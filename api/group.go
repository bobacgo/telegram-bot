package api

import (
	"bot/dto"
	"bot/repo"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GroupAPI struct {
	Group *repo.GroupRepo
}

func NewGroupAPI(repo *repo.Repo) *GroupAPI {
	return &GroupAPI{
		Group: repo.Group,
	}
}

func (api *GroupAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.GroupCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &repo.TelegramGroup{
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
	var req repo.GroupUpdateReq
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

	filter := &repo.TelegramGroupQuery{
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
