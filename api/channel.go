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
	Channel *repo.ChannelRepo
	bus     *bus.Bus
}

func NewChannelAPI(bus *bus.Bus, repo *repo.Repo) *ChannelAPI {
	return &ChannelAPI{
		Channel: repo.Channel,
		bus:     bus,
	}
}

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

	if err := api.Channel.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.notifyConfig(bus.OpAdd, row.TgChannelId)
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

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

	if err := api.Channel.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	api.notifyConfig(bus.OpDelete, row.TgChannelId)
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

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

	if err := api.Channel.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	row, err := api.Channel.FindOne(r.Context(), repo.ChannelFindOneReq{Id: req.Id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.notifyConfig(bus.OpUpdate, row.TgChannelId)

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

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
