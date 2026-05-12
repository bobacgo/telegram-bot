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

type TopicAPI struct {
	Topic *repo.GroupTopicRepo
}

func NewTopicAPI(repo *repo.Repo) *TopicAPI {
	return &TopicAPI{
		Topic: repo.GroupTopic,
	}
}

func (api *TopicAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.TopicCreateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := req.Validate(); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().Unix()
	row := &repo.TelegramGroupTopic{
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
	var req repo.TopicUpdateReq
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

	filter := &repo.TelegramGroupTopicQuery{}
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
