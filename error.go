package main

import (
	"errors"
	"strings"

	"gopkg.in/telebot.v4"
)

var (
	ErrRateLimit = errors.New("Too Many Requests")
	ErrNetwork   = errors.New("Network Error")
	ErrBotBanned = errors.New("Bot is banned or token is invalid")
)

// telegram: retry after 8 (429)
var rateLimitPatterns = []string{
	"retry after",
	"Too Many Requests",
	"429",
	"rate limit exceeded",
}

// 这类错误是临时性的，网络恢复后机器人可以继续使用
var networkErrorPatterns = []string{
	"timeout",
	"i/o timeout",
	"connection refused",
	"connection reset",
	"no such host",
	"network is unreachable",
	"dial tcp",
	"context deadline exceeded",
	"TLS handshake timeout",
	"EOF",
	"certificate",
}

// 401 Unauthorized - token 无效或机器人被删除
// 403 Forbidden - 机器人被封禁、踢出群组等
var bannedErrorPatterns = []string{
	"Unauthorized",
	"bot was blocked",
	"bot was kicked",
	"Forbidden",
}

// 识别错误，转成哨兵错误
func Unwrap(err error) error {
	if err == nil {
		return nil
	}

	// telebot 错误类型检查
	var teleErr *telebot.Error
	if errors.As(err, &teleErr) {
		switch teleErr.Code {
		case 401: // Unauthorized - token 无效
			return ErrBotBanned
		case 403: // Forbidden - 被封禁/踢出
			return ErrBotBanned
		}
	}

	switch {
	case contains(err.Error(), bannedErrorPatterns...):
		return ErrBotBanned
	case contains(err.Error(), rateLimitPatterns...):
		return ErrRateLimit
	case contains(err.Error(), networkErrorPatterns...):
		return ErrNetwork
	default:
		return err
	}
}

func contains(s string, substrs ...string) bool {
	s = strings.ToLower(s)
	for _, substr := range substrs {
		if strings.Contains(s, strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
