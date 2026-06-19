package api

import (
	"bot/internal/dto"
	"bot/internal/repo"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TopicAPI struct {
	Topic      *repo.GroupTopicRepo
	operateLog *operateLogger
}

func NewTopicAPI(repo *repo.Repo) *TopicAPI {
	return &TopicAPI{
		Topic:      repo.GroupTopic,
		operateLog: newOperateLogger(repo),
	}
}

// 创建群话题
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

	// 1.插入群话题配置到数据库
	if err := api.Topic.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.记录操作日志
	api.operateLog.write(r, repo.OpAdd, moduleTopic, topicTargetID(row.TgGroupId, row.TopicId), row, "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

// 删除群话题
func (api *TopicAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	// 1.删除群话题配置从数据库
	if err := api.Topic.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.记录操作日志
	api.operateLog.write(r, repo.OpDelete, moduleTopic, strconv.Itoa(id), "", "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

// 更新群话题
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

	// 1.更新群话题配置到数据库
	if err := api.Topic.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.记录操作日志
	api.operateLog.write(r, repo.OpUpdate, moduleTopic, strconv.Itoa(req.Id), req, "")

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

// 查询群话题列表
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

func topicTargetID(tgGroupId, topicId int64) string {
	return strconv.FormatInt(tgGroupId, 10) + ":" + strconv.FormatInt(topicId, 10)
}
