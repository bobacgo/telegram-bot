package bot

import (
	"bot/repo"
	"log/slog"
)

// AddChannel 添加或更新频道配置。
func (mgr *Manager) AddChannel(cfg *repo.TelegramChannel) {
	mgr.channels.Store(cfg.TgChannelId, cfg)
	slog.Info("[add] channel added", "channel_id", cfg.TgChannelId, "title", cfg.Title)
}

// RemoveChannel 删除频道配置。
func (mgr *Manager) RemoveChannel(channelId int64) {
	mgr.channels.Delete(channelId)
	slog.Warn("[remove] channel removed", "channel_id", channelId)
}

// 返回可用的channel 配置
func (mgr *Manager) GetChannelById(channelId int64) *repo.TelegramChannel {
	var channel *repo.TelegramChannel
	mgr.channels.Range(func(k, v any) bool {
		channel = v.(*repo.TelegramChannel)
		if channel.TgChannelId == channelId {
			return false
		}
		return true
	})
	return channel
}
