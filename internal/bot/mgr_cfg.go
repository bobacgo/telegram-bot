package bot

import (
	"bot/internal/bus"
	"bot/internal/repo"
	"context"
	"log/slog"
)

// 获取 bot 配置列表，初始化 BotManager 时调用
func (mgr *Manager) getBotCfg() []*repo.TelegramBot {
	f := &repo.TelegramBotQuery{Status: []int{StatusUsable, StatusNetwork}}
	botCfgs, err := mgr.repo.Bot.List(context.Background(), f)
	if err != nil {
		slog.Error("failed to get bot configs from repo", "err", err)
		return nil
	}
	return botCfgs
}

// 获取频道配置列表，初始化 BotManager 时调用
func (mgr *Manager) getChannelCfg() []*repo.TelegramChannel {
	f := &repo.TelegramChannelQuery{Status: repo.ChannelStatusUsable}
	channelCfgs, err := mgr.repo.Channel.List(context.Background(), f)
	if err != nil {
		slog.Error("failed to get channel configs from repo", "err", err)
		return nil
	}
	return channelCfgs
}

// 处理Bot配置变更事件，保持BotManager中的Bot实例与数据库中的配置同步
func (mgr *Manager) watchConfig() {
	for event := range mgr.bus.OutConfig() {
		cfg, ok := event.(*bus.ConfigEvent)
		if !ok {
			slog.Warn("received invalid config event", "event", event)
			continue
		}

		slog.Info("received config event", "cfg", cfg)

		switch cfg.CfgType {
		case bus.CfgBot:
			mgr.syncBot(cfg.OpType, cfg.ChatId)
		case bus.CfgChannel:
			mgr.syncChannel(cfg.OpType, cfg.ChatId)
		default:
			slog.Warn("unknown config type in event", "cfg_type", cfg.CfgType)
		}
	}
}

func (mgr *Manager) syncBot(opType int, botId int64) {

	slog.Info("syncing bot configuration", "op_type", opType, "bot_id", botId)

	switch opType {
	case bus.OpAdd:
		cfg, err := mgr.repo.Bot.FindOne(context.Background(), repo.BotFindOneReq{BotTgId: botId})
		if err != nil {
			slog.Error("failed to get bot config from repository", "bot_id", botId, "err", err)
			return
		}
		mgr.AddBotAndStart(cfg)
	case bus.OpUpdate:
		cfg, err := mgr.repo.Bot.FindOne(context.Background(), repo.BotFindOneReq{BotTgId: botId})
		if err != nil {
			slog.Error("failed to get bot config from repository", "bot_id", botId, "err", err)
			return
		}
		mgr.UpdateBot(cfg)
	case bus.OpDelete:
		mgr.RemoveBot(botId)
	default:
		slog.Warn("unknown operation type for bot config sync", "op_type", opType)
	}
}

func (mgr *Manager) syncChannel(opType int, channelId int64) {
	slog.Info("syncing channel configuration", "op_type", opType, "channel_id", channelId)

	switch opType {
	case bus.OpAdd, bus.OpUpdate:
		cfg, err := mgr.repo.Channel.FindOne(context.Background(), repo.ChannelFindOneReq{TgChannelId: channelId})
		if err != nil {
			slog.Error("failed to get channel config from repository", "channel_id", channelId, "err", err)
			return
		}
		if cfg.Status != repo.ChannelStatusUsable {
			mgr.RemoveChannel(channelId)
			return
		}
		mgr.AddChannel(cfg)
	case bus.OpDelete:
		mgr.RemoveChannel(channelId)
	default:
		slog.Warn("unknown operation type for channel config sync", "op_type", opType)
	}
}
