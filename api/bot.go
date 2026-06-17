package api

import (
	"bot/bus"
	"bot/dto"
	"bot/pkg"
	"bot/repo"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

type BotAPI struct {
	botRepo      *repo.BotRepo
	secretBotMap map[string]int64
	bus          *bus.Bus
	operateLog   *operateLogger
}

func NewBotAPI(bus *bus.Bus, repo *repo.Repo) *BotAPI {
	secretBotMap, err := repo.Bot.FindSecretBotMap(context.Background())
	if err != nil {
		slog.Error("failed to load webhook secret map", "error", err)
		secretBotMap = map[string]int64{}
	}
	return &BotAPI{
		botRepo:      repo.Bot,
		secretBotMap: secretBotMap,
		bus:          bus,
		operateLog:   newOperateLogger(repo),
	}
}

// 刷新 webhook secret map，适用于 bot 创建、更新、删除等操作后，确保 webhook 验证使用最新的 secret 配置
func (api *BotAPI) refreshSecretBotMap() {
	secretBotMap, err := api.botRepo.FindSecretBotMap(context.Background())
	if err != nil {
		slog.Error("failed to refresh webhook secret map", "error", err)
		return
	}
	api.secretBotMap = secretBotMap
}

// Webhook 处理 Telegram 服务器发送的更新请求，验证 secret 后将更新事件发送到总线供 BotManager 消费
func (api *BotAPI) Webhook(w http.ResponseWriter, r *http.Request) {
	secretKey := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	botID, ok := api.secretBotMap[secretKey]
	if !ok {
		writeErr(w, http.StatusForbidden, "invalid secret key")
		return
	}

	var upd telebot.Update
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid update body")
		return
	}

	evt := &bus.TgUpdateEvent{BotID: botID, Update: &upd}
	select {
	case api.bus.InWebhook() <- evt:
		writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok"})
	default:
		writeErr(w, http.StatusServiceUnavailable, "webhook queue is full")
	}
}

// 创建 bot，验证 token 获取 bot 信息，存储配置，并发送全量同步事件以启动 bot 实例
func (api *BotAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.BotCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	proxyUrl := r.Header.Get("proxy_url")
	me, err := pkg.BotGetMe(req.Token, proxyUrl)
	if err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("failed to get bot info: %v", err))
		return
	}

	now := time.Now().Unix()
	row := &repo.TelegramBot{
		BotTgId:       me.ID,
		Username:      me.Username,
		Token:         strings.TrimSpace(req.Token),
		WebhookSecret: strings.TrimSpace(req.WebhookSecret),
		Owner:         strings.TrimSpace(req.Owner),
		Type:          req.Type,
		HealthGroupId: req.HealthGroupId,
		Status:        req.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 1.插入 bot 配置到数据库
	if err := api.botRepo.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.刷新 webhook secret map，确保最新的 secret 配置生效
	api.refreshSecretBotMap()

	// 3.发送事件以启动 bot 实例
	api.bus.InConfig() <- &bus.ConfigEvent{
		OpType:  bus.OpAdd,
		CfgType: bus.CfgBot,
		ChatId:  row.BotTgId,
	}

	// 4.记录操作日志
	api.operateLog.write(r, repo.OpAdd, moduleBot, strconv.FormatInt(row.BotTgId, 10), row, "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok"})
}

// 删除 bot
func (api *BotAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	row, err := api.botRepo.FindOne(r.Context(), repo.BotFindOneReq{Id: id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 1.删除 bot 配置从数据库
	if err := api.botRepo.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.刷新 webhook secret map，确保最新的 secret 配置生效
	api.refreshSecretBotMap()

	// 3.停止 bot 实例，发送全量同步事件以移除失效 bot
	api.bus.InConfig() <- &bus.ConfigEvent{
		OpType:  bus.OpDelete,
		CfgType: bus.CfgBot,
		ChatId:  row.BotTgId,
	}
	// 4.记录操作日志
	api.operateLog.write(r, repo.OpDelete, moduleBot, strconv.FormatInt(row.BotTgId, 10), "", "")

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

// 更新 bot
func (api *BotAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req repo.BotUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.Id <= 0 {
		writeErr(w, http.StatusBadRequest, "id is required and must be > 0")
		return
	}
	// 1.更新 bot 配置到数据库
	if err := api.botRepo.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.刷新 webhook secret map，确保最新的 secret 配置生效
	api.refreshSecretBotMap()

	row, err := api.botRepo.FindOne(r.Context(), repo.BotFindOneReq{Id: req.Id})
	if err == nil {
		// 3.发送全量同步事件，管理器会自动对比并应用配置更新或重启 bot
		api.bus.InConfig() <- &bus.ConfigEvent{
			OpType:  bus.OpUpdate,
			CfgType: bus.CfgBot,
			ChatId:  row.BotTgId,
		}
		// 4.记录操作日志
		api.operateLog.write(r, repo.OpUpdate, moduleBot, strconv.FormatInt(row.BotTgId, 10), row, "")
	} else {
		slog.Error("[Update] failed to find bot", "id", req.Id)
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

// 查询 bot 列表，支持根据 owner、type、status 等参数过滤，返回符合条件的 bot 配置列表
func (api *BotAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &repo.TelegramBotQuery{
		Owner: urlValues.Get("owner"),
	}
	if v := urlValues.Get("type"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.Type = n
	}
	if v := urlValues.Get("status"); v != "" {
		stringsList := strings.Split(v, ",")
		for _, s := range stringsList {
			n, _ := strconv.Atoi(s)
			filter.Status = append(filter.Status, n)
		}
	}

	rows, err := api.botRepo.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}
