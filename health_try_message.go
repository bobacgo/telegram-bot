package main

import (
	"errors"
	"log/slog"
	"time"
)

type HealthTryMessage struct {
	mgr *BotManager
}

func (h *HealthTryMessage) Cfg() *HeartbeatConfig {
	return &HeartbeatConfig{
		InitialDelay: 1 * time.Minute,
		Interval:     5 * time.Minute,
		EmptyWait:    2 * time.Minute,
	}
}

func (h *HealthTryMessage) List() []int64 {
	return h.mgr.ReadyBotIds()
}

func (h *HealthTryMessage) Ping(idx int, id int64) error {
	bot := h.mgr.GetBotById(id)
	if bot == nil || !bot.IsReady() { // bot 可能已被处理了，忽略
		return nil
	}
	err := h.trySendMessage(idx, bot)
	return Unwrap(err) // 识别错误，转成哨兵错误
}

func (h *HealthTryMessage) OnError(id int64, err error) {
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
		bot.status.Store(StatusBan) // 更新 bot 状态
		h.mgr.sendAlert(0, nil)     // TODO 发送告警群
	case errors.Is(err, ErrNetwork):
		bot.failCount++
		if bot.failCount < attemptLimit {
			slog.Error("bot health check failed, retrying", "bot_id", id, "attempt", bot.failCount, "error", err)
			return
		}

		slog.Warn("network error detected for bot, marking as network issue", "error", err)
		bot.status.Store(StatusNetwork)
		h.mgr.sendAlert(0, nil) // TODO 发送告警群
	default:
		slog.Error("unexpected error during health check", "error", err)
	}

	bot.failCount = 0 // 重置失败计数，等待下一次检测
}

func (h *HealthTryMessage) trySendMessage(idx int, bot *Bot) error {
	// TODO 发送心跳消息，内容可以是固定的，也可以包含时间戳等信息
	// 发送到管理员或测试群，避免对用户造成干扰
	// 可以考虑先删除上一次的心跳消息，避免积累过多消息

	// 示例：发送到管理员
	// TODO 替换成实际的管理员 ID 或群 ID
	adminId := int64(123456789)
	msg, err := bot.SendHeartbeat(adminId, idx) // idx 可以根据需要传入当前检测的 bot 顺序等信息
	if err != nil {
		return err
	}

	// 删除上一次的心跳消息
	if bot.lastHeartbeatMsgId != 0 {
		_ = bot.DeleteMessage(adminId, int(bot.lastHeartbeatMsgId)) // 忽略删除错误
	}
	bot.lastHeartbeatMsgId = msg.ID // 保存当前心跳消息 ID，供下一次删除使用
	return nil
}
