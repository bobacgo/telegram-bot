package app

import (
	"bot/pkg"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ApiResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ApiResp{Code: status, Msg: msg})
}

type API struct {
	BotAPI     BotAPI
	ChannelAPI ChannelAPI
	GroupAPI   GroupAPI
	TopicAPI   TopicAPI
}

func NewAPI(db *DB) *API {
	return &API{
		BotAPI:     *NewBotAPI(db),
		ChannelAPI: *NewChannelAPI(db),
		GroupAPI:   *NewGroupAPI(db),
		TopicAPI:   *NewTopicAPI(db),
	}
}

// === Bot API ===

type BotAPI struct {
	Bot *BotRepo
}

func NewBotAPI(db *DB) *BotAPI {
	return &BotAPI{
		Bot: &BotRepo{db: db},
	}
}

func (api *BotAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req BotCreateReq
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
	row := &TelegramBot{
		BotTgId:       me.ID,
		Username:      me.Username,
		Token:         strings.TrimSpace(req.Token),
		WebhookSecret: strings.TrimSpace(req.WebhookSecret),
		Owner:         strings.TrimSpace(req.Owner),
		Type:          req.Type,
		Status:        req.Status,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := api.Bot.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
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

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})

}

func (api *BotAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req BotUpdateReq
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

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

func (api *BotAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &TelegramBotQuery{
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

	rows, err := api.Bot.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}

// === Channel API ===

type ChannelAPI struct {
	Channel *ChannelRepo
}

func NewChannelAPI(db *DB) *ChannelAPI {
	return &ChannelAPI{
		Channel: &ChannelRepo{db: db},
	}
}

func (api *ChannelAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req ChannelCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &TelegramChannel{
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

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

func (api *ChannelAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := api.Channel.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

func (api *ChannelAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req ChannelUpdateReq
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

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

func (api *ChannelAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &TelegramChannelQuery{
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

// === Group API ===

type GroupAPI struct {
	Group *GroupRepo
}

func NewGroupAPI(db *DB) *GroupAPI {
	return &GroupAPI{
		Group: &GroupRepo{db: db},
	}
}

func (api *GroupAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req GroupCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &TelegramGroup{
		TgGroupId: req.TgGroupId,
		Title:     strings.TrimSpace(req.Title),
		Username:  strings.TrimSpace(req.Username),
		Owner:     strings.TrimSpace(req.Owner),
		Type:      req.Type,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := api.Group.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

func (api *GroupAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := api.Group.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

func (api *GroupAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req GroupUpdateReq
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

	if err := api.Group.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

func (api *GroupAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &TelegramGroupQuery{
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

	rows, err := api.Group.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}

// === Topic API ===

type TopicAPI struct {
	Topic *TopicRepo
}

func NewTopicAPI(db *DB) *TopicAPI {
	return &TopicAPI{
		Topic: &TopicRepo{db: db},
	}
}

func (api *TopicAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req TopicCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &TelegramGroupTopic{
		TgGroupId: req.TgGroupId,
		TopicId:   req.TopicId,
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := api.Topic.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

func (api *TopicAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := api.Topic.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

func (api *TopicAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req TopicUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Id <= 0 {
		writeErr(w, http.StatusBadRequest, "id is required and must be > 0")
		return
	}
	if req.Name != "" {
		req.Name = strings.TrimSpace(req.Name)
	}

	if err := api.Topic.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

func (api *TopicAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	filter := &TelegramGroupTopicQuery{}
	if v := urlValues.Get("tg_group_id"); v != "" {
		n, _ := strconv.ParseInt(v, 10, 64)
		filter.TgGroupId = n
	}
	if v := urlValues.Get("topic_id"); v != "" {
		n, _ := strconv.ParseInt(v, 10, 64)
		filter.TopicId = n
	}
	filter.Name = strings.TrimSpace(urlValues.Get("name"))

	rows, err := api.Topic.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}
