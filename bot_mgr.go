package main

import (
	"log/slog"
	"sync"
	"time"
)

// BotManager manages multiple Bot instances.
// 多个Bot共享同一业务的存储实例（如user_topic存储）
type BotManager struct {
	bots            sync.Map // map[botid]*Bot
	DB              DB       // map[string]kv存储实例，按业务维度区分
	mu              sync.Mutex
	activeBotTokens map[string]int64 // token -> botID mapping for cleanup
}

// NewBotManager 创建BotManager
// tokens: Bot token列表
// db: 数据库实例，按业务维度管理存储
func NewBotManager(tokens []string, db DB) *BotManager {
	mgr := &BotManager{
		DB:              db,
		activeBotTokens: make(map[string]int64),
	}

	for _, token := range tokens {
		b := NewBot(token, db)
		mgr.bots.Store(b.BotId, b)
		mgr.mu.Lock()
		mgr.activeBotTokens[token] = b.BotId
		mgr.mu.Unlock()
	}
	return mgr
}

func (mgr *BotManager) Start() {
	for _, bot := range mgr.Bots() {
		go bot.Start()
		slog.Info("[start] bot started", "bot_id", bot.BotId, "bot_name", bot.Username)
	}
}

func (mgr *BotManager) Stop() {
	for _, bot := range mgr.Bots() {
		bot.Stop()
		slog.Info("[stop] bot stopped", "bot_id", bot.BotId, "bot_name", bot.Username)
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

func (mgr *BotManager) AddBot(botId int64, bot *Bot) {
	go bot.Start()

	time.Sleep(time.Millisecond * 100) // wait for bot to start
	mgr.bots.Store(botId, bot)
	slog.Info("[add] bot added", "bot_id", botId, "bot_name", bot.Username)
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

	slog.Warn("[remove] bot removed", "bot_id", botId, "bot_name", b.Username)
}
