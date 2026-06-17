package api

import (
	"bot/bus"
	"bot/dto"
	"bot/repo"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ChannelAPI struct {
	Channel    *repo.ChannelRepo
	bus        *bus.Bus
	operateLog *operateLogger
}

func NewChannelAPI(bus *bus.Bus, repo *repo.Repo) *ChannelAPI {
	return &ChannelAPI{
		Channel:    repo.Channel,
		bus:        bus,
		operateLog: newOperateLogger(repo),
	}
}

// 创建频道
func (api *ChannelAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.ChannelCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &repo.TelegramChannel{
		TgChannelId: req.TgChannelId,
		Title:       strings.TrimSpace(req.Title),
		Username:    strings.TrimSpace(req.Username),
		Owner:       strings.TrimSpace(req.Owner),
		Type:        req.Type,
		Status:      req.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// 1.插入频道配置到数据库
	if err := api.Channel.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.发送事件以启动频道实例
	api.notifyConfig(bus.OpAdd, row.TgChannelId)
	// 3.记录操作日志
	api.operateLog.write(r, repo.OpAdd, moduleChannel, strconv.FormatInt(row.TgChannelId, 10), row, "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

// 删除频道
func (api *ChannelAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	row, err := api.Channel.FindOne(r.Context(), repo.ChannelFindOneReq{Id: id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 1.删除频道配置从数据库
	if err := api.Channel.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.发送事件以停止频道实例
	api.notifyConfig(bus.OpDelete, row.TgChannelId)
	// 3.记录操作日志
	api.operateLog.write(r, repo.OpDelete, moduleChannel, strconv.FormatInt(row.TgChannelId, 10), "", "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

// 更新频道
func (api *ChannelAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req repo.ChannelUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Id <= 0 {
		writeErr(w, http.StatusBadRequest, "id is required and must be > 0")
		return
	}

	if req.Title != "" {
		req.Title = strings.TrimSpace(req.Title)
	}
	if req.Owner != "" {
		req.Owner = strings.TrimSpace(req.Owner)
	}
	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		req.Username = &username
	}

	// 1.更新频道配置到数据库
	if err := api.Channel.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	row, err := api.Channel.FindOne(r.Context(), repo.ChannelFindOneReq{Id: req.Id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 2.发送事件以更新频道实例
	api.notifyConfig(bus.OpUpdate, row.TgChannelId)
	// 3.记录操作日志
	api.operateLog.write(r, repo.OpUpdate, moduleChannel, strconv.FormatInt(row.TgChannelId, 10), row, "")

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

// 查询频道列表
func (api *ChannelAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &repo.TelegramChannelQuery{
		Owner: urlValues.Get("owner"),
	}
	if v := urlValues.Get("type"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.Type = n
	}
	if v := urlValues.Get("status"); v != "" {
		n, _ := strconv.Atoi(v)
		filter.Status = n
	}

	rows, err := api.Channel.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}

func (api *ChannelAPI) notifyConfig(opType int, channelId int64) {
	if api.bus == nil || channelId == 0 {
		return
	}
	api.bus.InConfig() <- &bus.ConfigEvent{
		OpType:  opType,
		CfgType: bus.CfgChannel,
		ChatId:  channelId,
	}
}
