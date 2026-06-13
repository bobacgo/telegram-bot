package bot

import (
	"bot/bus"
	"bot/repo"
	"context"
	"log/slog"
	"sync"
	"time"
)

// Manager manages multiple Bot instances.
// 多个Bot共享同一业务的存储实例（如user_topic存储）
type Manager struct {
	bots sync.Map // map[botid]*Bot

	// DB              DB       // map[string]kv存储实例，按业务维度区分
	repo       *repo.Repo
	bus        *bus.Bus
	cancel     context.CancelFunc // 用于停止健康检查goroutines
	webhookURL string
}

// NewManager 创建BotManager
// tokens: Bot token列表
// db: 数据库实例，按业务维度管理存储
func NewManager(webhookURL string, bus *bus.Bus, repo *repo.Repo) *Manager {
	mgr := &Manager{
		webhookURL: webhookURL,
		bus:        bus,
		repo:       repo,
	}

	cfgs := mgr.getBotCfg()
	for _, cfg := range cfgs {
		mgr.AddBot(cfg)
	}
	return mgr
}

// 监听 webhook 更新事件，分发到对应的 Bot 实例处理
func (mgr *Manager) onWebhook() {
	for evt := range mgr.bus.OutWebhook() {
		if evt == nil || evt.Update == nil {
			continue
		}

		b := mgr.GetBotById(evt.BotID)
		if b == nil {
			slog.Warn("drop webhook update: bot not found", "bot_id", evt.BotID, "update_id", evt.Update.ID)
			continue
		}
		b.tgBot.ProcessUpdate(*evt.Update)
	}
}

func (mgr *Manager) Start() {
	for _, bot := range mgr.Bots() {
		slog.Info("[start] bot started", "bot_id", bot.cfg.BotTgId, "bot_name", bot.cfg.Username)
		go bot.Start()
	}

	// 创建可以被取消的context用于健康检查goroutines
	ctx, cancel := context.WithCancel(context.Background())
	mgr.cancel = cancel

	go mgr.watchConfig() // 接收配置刷新
	go mgr.onWebhook()
	go runHealthCheck(ctx, &HealthGetMe{mgr: mgr})      // bot 健康检测 - GetMe接口检测
	go runHealthCheck(ctx, &HealthTryMessage{mgr: mgr}) // bot 健康检测 - 发送消息检测
	go runHealthCheck(ctx, &HealthChannel{mgr: mgr})    // 频道 健康检测
}

func (mgr *Manager) Stop() {
	slog.Info("[stop] stopping health checks...")
	// 停止健康检查goroutines
	if mgr.cancel != nil {
		mgr.cancel()
	}
	slog.Info("[stop] health checks stopped")

	slog.Info("[stop] stopping bots...")
	bots := mgr.Bots()

	// 使用goroutine + channel方式停止bot，带超时保护
	done := make(chan struct{})
	go func() {
		for _, bot := range bots {
			bot.Stop()
			slog.Info("[stop] bot stopped", "bot_id", bot.cfg.BotTgId, "bot_name", bot.cfg.Username)
		}
		close(done)
	}()

	// 等待5秒，如果bot停止超时则继续
	select {
	case <-done:
		slog.Info("[stop] all bots stopped")
	case <-time.After(5 * time.Second):
		slog.Warn("[stop] bot stop timeout, continuing anyway")
	}
}

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
		if bot.cfg.Type == BotTypeAlert {
			alertBot = bot
			return false // 找到后停止遍历
		}
		return true
	})

	return alertBot
}

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
