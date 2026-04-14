package main

import (
	"errors"
	"log/slog"
	"time"
)

type HealthChannel struct {
	mgr *BotManager
}

func (h *HealthChannel) Cfg() *HeartbeatConfig {
	return &HeartbeatConfig{
		InitialDelay: 1 * time.Minute,
		Interval:     5 * time.Minute,
		EmptyWait:    2 * time.Minute,
	}
}

func (h *HealthChannel) List() []int64 {
	return h.mgr.ReadyBotIds()
}

func (h *HealthChannel) Ping(idx int, id int64) error {
	bot := h.mgr.GetBotById(id)
	if bot == nil || !bot.IsReady() { // bot 可能已被处理了，忽略
		return nil
	}
	err := h.checkChannel(bot)
	return Unwrap(err) // 识别错误，转成哨兵错误
}

func (h *HealthChannel) OnError(id int64, err error) {
	if id == 0 || err == nil {
		return
	}

	bot := h.mgr.GetBotById(id)
	if bot == nil {
		slog.Error("bot not find", "id", id)
		return
	}

	switch {
	case errors.Is(err, ErrBotBanned):
		slog.Warn("bot is banned or token is invalid, marking as banned", "error", err)
	case errors.Is(err, ErrNetwork):
		slog.Warn("network error detected, marking bot as network error", "error", err)
	default:
		slog.Error("unexpected error during health check", "error", err)
	}
}

func (h *HealthChannel) checkChannel(bot *Bot) error {
	// 通过频道id获取频道信息，如果失败则认为频道不可用
	return nil
}
