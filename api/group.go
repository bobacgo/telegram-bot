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
	Group      *repo.GroupRepo
	operateLog *operateLogger
}

func NewGroupAPI(repo *repo.Repo) *GroupAPI {
	return &GroupAPI{
		Group:      repo.Group,
		operateLog: newOperateLogger(repo),
	}
}

// 创建群组
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

	// 1.插入群组配置到数据库
	if err := api.Group.Insert(r.Context(), row); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.记录操作日志
	api.operateLog.write(r, repo.OpAdd, moduleGroup, strconv.FormatInt(row.TgGroupId, 10), row, "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

// 删除群组
func (api *GroupAPI) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	row, err := api.Group.FindOne(r.Context(), repo.GroupFindOneReq{Id: id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 1.删除群组配置从数据库
	if err := api.Group.Delete(r.Context(), id); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 2.记录操作日志
	api.operateLog.write(r, repo.OpDelete, moduleGroup, strconv.FormatInt(row.TgGroupId, 10), row, "")
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": id}})
}

// 更新群组
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

	// 1.更新群组配置到数据库
	if err := api.Group.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	row, err := api.Group.FindOne(r.Context(), repo.GroupFindOneReq{Id: req.Id})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 2.记录操作日志
	api.operateLog.write(r, repo.OpUpdate, moduleGroup, strconv.FormatInt(row.TgGroupId, 10), row, "")

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"id": req.Id}})
}

// 查询群组列表
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
