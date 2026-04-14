package main

import (
	"errors"
	"log/slog"
	"time"

	"gopkg.in/telebot.v4"
)

type HealthGetMe struct {
	mgr *BotManager
}

func (h *HealthGetMe) Cfg() *HeartbeatConfig {
	return &HeartbeatConfig{
		InitialDelay: 30 * time.Second,
		Interval:     2 * time.Minute,
		EmptyWait:    1 * time.Minute,
	}
}

func (h *HealthGetMe) List() []int64 {
	return h.mgr.ReadyBotIds()
}

func (h *HealthGetMe) Ping(idx int, id int64) error {
	bot := h.mgr.GetBotById(id)
	if bot == nil || !bot.IsReady() { // bot 可能已被处理了，忽略
		return nil
	}
	err := h.mgr.getMe(bot.tgBot)
	return Unwrap(err) // 识别错误，转成哨兵错误
}

func (h *HealthGetMe) OnError(id int64, err error) {
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
		slog.Error("bot is banned or token is invalid, marking as banned", "error", err)
		bot.status.Store(StatusBan) // 更新 bot 状态
		h.mgr.sendAlert(0, nil)     // TODO 发送告警群
	case errors.Is(err, ErrNetwork): // 如果网络问题需要连续失败多次才标记为异常，避免偶发的网络问题导致误判
		bot.getMeFailCount++
		if bot.getMeFailCount < attemptLimit {
			slog.Error("bot health check failed, retrying", "bot_id", id, "attempt", bot.getMeFailCount, "error", err)
			return
		}

		slog.Warn("network error detected for bot, marking as network issue", "error", err)
		bot.status.Store(StatusNetwork)
		h.mgr.sendAlert(0, nil) // TODO 发送告警群
	case errors.Is(err, ErrRateLimit):
		slog.Warn("rate limit hit for bot during health check, keeping current status", "error", err)
	default:
		slog.Error("unexpected error during health check ping", "error", err)
	}

	bot.getMeFailCount = 0 // 重置失败计数，等待下一次检测
}

func (mgr *BotManager) getMe(bot *telebot.Bot) error {
	// 使用 Raw API 调用 getMe 方法
	// 检测 token 是否还正常，进而确认Bot是否可用
	// Note:
	//   这个不能 100% 代表 bot 正常，有些 bot 异常但是 bot 不能发收发消息
	_, err := bot.Raw("getMe", map[string]string{})
	return err
}
