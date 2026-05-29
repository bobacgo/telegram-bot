package api

import (
	"bot/bus"
	"bot/dto"
	"bot/pkg"
	"bot/repo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

type BotAPI struct {
	Bot          *repo.BotRepo
	secretBotMap map[string]int64
	bus          *bus.Bus
}

func NewBotAPI(bus *bus.Bus, repo *repo.Repo) *BotAPI {
	secretBotMap, err := repo.Bot.FindSecretBotMap(context.Background())
	if err != nil {
		slog.Error("failed to load webhook secret map", "error", err)
		secretBotMap = map[string]int64{}
	}
	return &BotAPI{
		Bot:          repo.Bot,
		secretBotMap: secretBotMap,
		bus:          bus,
	}
}

func (api *BotAPI) refreshSecretBotMap() {
	secretBotMap, err := api.Bot.FindSecretBotMap(context.Background())
	if err != nil {
		slog.Error("failed to refresh webhook secret map", "error", err)
		return
	}
	api.secretBotMap = secretBotMap
}

func (api *BotAPI) Webhook(w http.ResponseWriter, r *http.Request) {
	secretKey := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	botID, ok := api.secretBotMap[secretKey]
	if !ok {
		writeErr(w, http.StatusForbidden, "invalid secret key")
		return
	}

	body := http.MaxBytesReader(w, r.Body, 1<<20)
	defer body.Close()

	var upd telebot.Update
	if err := json.NewDecoder(body).Decode(&upd); err != nil {
		if errors.Is(err, io.EOF) {
			writeErr(w, http.StatusBadRequest, "empty body")
			return
		}
		writeErr(w, http.StatusBadRequest, "invalid update body")
		return
	}

	if api.bus == nil {
		writeErr(w, http.StatusServiceUnavailable, "event bus not ready")
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

	if err := api.Bot.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.refreshSecretBotMap()

	// 如果是开启状态，启动 bot 实例
	// TODO 启动 bot 实例

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok"})
}

func (api *BotAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := api.Bot.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.refreshSecretBotMap()

	// TODO 停止 bot 实例

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

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

	if err := api.Bot.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.refreshSecretBotMap()

	// TODO 如果更新了 token 或 webhook secret，重启 bot 实例

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

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

	rows, err := api.Bot.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}
