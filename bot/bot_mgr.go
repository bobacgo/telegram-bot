package bot

import (
	"bot/repo"
	"context"
	"log/slog"
	"sync"
	"time"
)

// BotManager manages multiple Bot instances.
// 多个Bot共享同一业务的存储实例（如user_topic存储）
type BotManager struct {
	bots            sync.Map         // map[botid]*Bot
	activeBotTokens map[string]int64 // token -> botID mapping for cleanup
	// DB              DB       // map[string]kv存储实例，按业务维度区分
	repo *repo.Repo
	mu   sync.Mutex
}

// NewBotManager 创建BotManager
// tokens: Bot token列表
// db: 数据库实例，按业务维度管理存储
func NewBotManager(repo *repo.Repo, webhookURL string) *BotManager {
	mgr := &BotManager{
		repo:            repo,
		activeBotTokens: make(map[string]int64),
	}

	cfgs := mgr.getBotCfg()
	for _, cfg := range cfgs {
		b := NewBot(cfg, webhookURL)
		mgr.bots.Store(b.botCfg.BotTgId, b)
		mgr.mu.Lock()
		mgr.activeBotTokens[cfg.Token] = b.botCfg.BotTgId
		mgr.mu.Unlock()
	}
	ctx := context.Background()

	go runHealthCheck(ctx, &HealthGetMe{mgr: mgr})      // bot 健康检测 - GetMe接口检测
	go runHealthCheck(ctx, &HealthTryMessage{mgr: mgr}) // bot 健康检测 - 发送消息检测
	go runHealthCheck(ctx, &HealthChannel{mgr: mgr})    // 频道 健康检测
	return mgr
}

func (mgr *BotManager) Start() {
	for _, bot := range mgr.Bots() {
		go bot.Start()
		slog.Info("[start] bot started", "bot_id", bot.botCfg.BotTgId, "bot_name", bot.botCfg.Username)
	}
}

func (mgr *BotManager) Stop() {
	for _, bot := range mgr.Bots() {
		bot.Stop()
		slog.Info("[stop] bot stopped", "bot_id", bot.botCfg.BotTgId, "bot_name", bot.botCfg.Username)
	}
}

func (mgr *BotManager) Bots() []*Bot {
	res := make([]*Bot, 0)
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		res = append(res, bot)
		return true
	})
	return res
}

// 获取可使用的Bot列表
func (mgr *BotManager) UsableBots() []*Bot {
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
func (mgr *BotManager) ReadyBotIds() []int64 {
	readyBots := make([]int64, 0)
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.IsReady() {
			readyBots = append(readyBots, bot.botCfg.BotTgId)
		}
		return true
	})
	return readyBots
}

// 通过 bot ID 获取 Bot 实例
func (mgr *BotManager) GetBotById(botId int64) *Bot {
	bAny, ok := mgr.bots.Load(botId)
	if !ok {
		return nil
	}
	return bAny.(*Bot)
}

// 获取告警 Bot 实例
func (mgr *BotManager) GetAlertBot() *Bot {
	var alertBot *Bot
	mgr.bots.Range(func(k, v any) bool {
		bot := v.(*Bot)
		if bot.botCfg.Type == BotTypeAlert {
			alertBot = bot
			return false // 找到后停止遍历
		}
		return true
	})

	return alertBot
}

func (mgr *BotManager) AddBot(botId int64, bot *Bot) {
	go bot.Start()

	time.Sleep(time.Millisecond * 100) // wait for bot to start
	mgr.bots.Store(botId, bot)
	slog.Info("[add] bot added", "bot_id", botId, "bot_name", bot.botCfg.Username)
}

func (mgr *BotManager) RemoveBot(botId int64) {
	bAny, ok := mgr.bots.Load(botId)
	if !ok {
		slog.Error("[remove] bot not found", "bot_id", botId)
		return
	}
	mgr.bots.Delete(botId)

	b := bAny.(*Bot)
	b.Stop()

	slog.Warn("[remove] bot removed", "bot_id", botId, "bot_name", b.botCfg.Username)
}

func (mgr *BotManager) getBotCfg() []*repo.TelegramBot {
	f := &repo.TelegramBotQuery{Status: []int{StatusUsable, StatusNetwork}}
	botCfgs, err := mgr.repo.Bot.List(context.Background(), f)
	if err != nil {
		slog.Error("failed to get bot configs from repo", "err", err)
		return nil
	}
	return botCfgs
}
