package bot

import (
	"bot/internal/repo"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type HealthChannel struct {
	mgr *Manager
}

func (h *HealthChannel) Cfg() *HeartbeatConfig {
	return &HeartbeatConfig{
		HealthName:   "channel",
		InitialDelay: 1 * time.Minute,
		Interval:     5 * time.Minute,
		EmptyWait:    2 * time.Minute,
	}
}

func (h *HealthChannel) List() []int64 {
	return h.mgr.ReadyBotIds()
}

func (h *HealthChannel) Ping(idx int, id int64) error {
	bot := h.mgr.GetChannelBot()
	if bot == nil { // bot 可能已被处理了，忽略
		return fmt.Errorf("channel bot is nil channelId = %d", id)
	}
	channelCfg := h.mgr.GetChannelById(id)
	if channelCfg == nil {
		return fmt.Errorf("channel id %d not found", id)
	}

	err := h.checkChannel(bot, channelCfg)
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
		slog.Warn("bot cannot access channel or token is invalid", "error", err)
	case errors.Is(err, ErrNetwork):
		slog.Warn("network error detected, marking bot as network error", "error", err)
	default:
		slog.Error("unexpected error during health check", "error", err)
	}
}

func (h *HealthChannel) checkChannel(bot *Bot, channel *repo.TelegramChannel) error {
	if _, err := bot.GetChatById(channel.TgChannelId); err != nil {
		slog.Error("channel health check failed", "bot_id", bot.cfg.BotTgId, "channel_id", channel.TgChannelId, "title", channel.Title, "err", err)
		return fmt.Errorf("check channel failed: bot_id=%d channel_id=%d title=%q: %w", bot.cfg.BotTgId, channel.TgChannelId, channel.Title, err)
	}
	slog.Debug("channel health check passed", "bot_id", bot.cfg.BotTgId, "channel_id", channel.TgChannelId, "title", channel.Title)
	return nil
}
