package api

import (
	"bot/internal/repo"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OperateLogAPI struct {
	OperateLog *repo.OperateLogRepo
}

func NewOperateLogAPI(repo *repo.Repo) *OperateLogAPI {
	return &OperateLogAPI{
		OperateLog: repo.OperateLog,
	}
}

// List 获取操作日志列表，支持按模块名称和目标 ID 过滤，并支持分页。
func (api *OperateLogAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()

	page, _ := strconv.Atoi(urlValues.Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(urlValues.Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := &repo.OperateLogQuery{
		ModuleName: strings.TrimSpace(urlValues.Get("module_name")),
		TargetId:   strings.TrimSpace(urlValues.Get("target_id")),
		Limit:      pageSize,
		Offset:     (page - 1) * pageSize,
	}

	rows, err := api.OperateLog.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}

const (
	moduleBot     = "bot"         // bot 管理
	moduleChannel = "channel"     // channel 管理
	moduleGroup   = "group"       // group 管理
	moduleTopic   = "group-topic" // group topic 管理
)

type operateLogger struct {
	repo *repo.OperateLogRepo
}

func newOperateLogger(repoAll *repo.Repo) *operateLogger {
	return &operateLogger{repo: repoAll.OperateLog}
}

func (l *operateLogger) write(r *http.Request, opType int, moduleName, targetId string, content any, remark string) {
	contentText := "{}"
	if content != nil {
		byt, err := json.Marshal(content)
		if err != nil {
			slog.Error("marshal operate log content failed", "module", moduleName, "target_id", targetId, "err", err)
		} else {
			contentText = string(byt)
		}
	}

	row := &repo.OperateLog{
		Operator:   r.Header.Get("username"),
		OperateAt:  time.Now().Unix(),
		IpAddress:  requestIP(r),
		OpType:     opType,
		ModuleName: moduleName,
		TargetId:   targetId,
		Content:    contentText,
		Remark:     remark,
	}
	if err := l.repo.Insert(context.Background(), row); err != nil {
		slog.Error("write operate log failed", "module", moduleName, "target_id", targetId, "err", err)
	}
}

// requestIP 从 HTTP 请求中提取客户端 IP 地址。
func requestIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwardedFor != "" {
		if idx := strings.Index(forwardedFor, ","); idx >= 0 {
			forwardedFor = forwardedFor[:idx]
		}
		return strings.TrimSpace(forwardedFor)
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
