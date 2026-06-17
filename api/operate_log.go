package api

import (
	"bot/repo"
	"context"
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
)

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
		Operator:   requestOperator(r),
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

// requestOperator 从 HTTP 请求中提取操作人信息。
func requestOperator(r *http.Request) string {
	if r == nil {
		return "unknown"
	}
	for _, key := range []string{
		"X-Operator",
		"X-User",
		"X-Username",
	} {
		if v := strings.TrimSpace(r.Header.Get(key)); v != "" {
			return v
		}
	}
	return "unknown"
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
