package bot

import (
	"bot/internal/repo"
	"log/slog"
)

// 添加新 Bot 实例
func (mgr *Manager) AddBot(cfg *repo.TelegramBot) {
	b := NewBot(cfg, mgr.webhookURL, mgr.repo)
	mgr.bots.Store(b.cfg.BotTgId, b)

	slog.Info("[add] bot added", "bot_id", cfg.BotTgId, "bot_name", cfg.Username)
}

// 添加并启动新 Bot 实例
func (mgr *Manager) AddBotAndStart(cfg *repo.TelegramBot) {
	mgr.AddBot(cfg)

	// 启动新添加的 bot 实例
	bAny, ok := mgr.bots.Load(cfg.BotTgId)
	if !ok {
		slog.Error("[add] bot not found after adding", "bot_id", cfg.BotTgId)
		return
	}
	b, _ := bAny.(*Bot)
	go b.Start()
}

// 更新 Bot 实例配置
func (mgr *Manager) UpdateBot(cfg *repo.TelegramBot) {
	mgr.RemoveBot(cfg.BotTgId)
	mgr.AddBotAndStart(cfg)
}

// 删除 Bot 实例并停止运行
func (mgr *Manager) RemoveBot(botId int64) {
	bAny, ok := mgr.bots.Load(botId)
	if !ok {
		slog.Error("[remove] bot not found", "bot_id", botId)
		return
	}
	mgr.bots.Delete(botId)

	b := bAny.(*Bot)
	b.Stop()

	slog.Warn("[remove] bot removed", "bot_id", botId, "bot_name", b.cfg.Username)
}

// 获取所有的 Bot 实例列表（包括不可用的）
func (mgr *Manager) Bots() []*Bot {
	res := make([]*Bot, 0)
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		res = append(res, bot)
		return true
	})
	return res
}

// 获取可使用的Bot列表
func (mgr *Manager) UsableBots() []*Bot {
	activeBots := make([]*Bot, 0)
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.IsHealthy() {
			activeBots = append(activeBots, bot)
		}
		return true
	})

	return activeBots
}

// 获取可使用的 bot 和 网络异常的 bot 列表
func (mgr *Manager) ReadyBotIds() []int64 {
	readyBots := make([]int64, 0)
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.IsReady() {
			readyBots = append(readyBots, bot.cfg.BotTgId)
		}
		return true
	})
	return readyBots
}

// 通过 bot ID 获取 Bot 实例
func (mgr *Manager) GetBotById(botId int64) *Bot {
	bAny, ok := mgr.bots.Load(botId)
	if !ok {
		return nil
	}
	return bAny.(*Bot)
}

// 获取告警 Bot 实例
func (mgr *Manager) GetAlertBot() *Bot {
	var alertBot *Bot
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.cfg.Type == BotTypeAlert && bot.IsHealthy() {
			alertBot = bot
			return false // 找到后停止遍历
		}
		return true
	})

	return alertBot
}

// 获取频道 Bot 实例
func (mgr *Manager) GetChannelBot() *Bot {
	var channelBot *Bot
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.cfg.Type == BotTypeChannel && bot.IsHealthy() {
			channelBot = bot
			return false // 找到后停止遍历
		}
		return true
	})
	return channelBot
}
