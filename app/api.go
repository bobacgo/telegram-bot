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

type API struct {
	Bot *BotRepo
}

func NewAPI(db *DB) *API {
	return &API{
		Bot: &BotRepo{db: db},
	}
}

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

func (api *API) BotCreate(w http.ResponseWriter, r *http.Request) {
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

func (api *API) BotDelete(w http.ResponseWriter, r *http.Request) {
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

func (api *API) BotUpdate(w http.ResponseWriter, r *http.Request) {
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

func (api *API) BotList(w http.ResponseWriter, r *http.Request) {
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
