package main

import (
	"log/slog"
	"sync"
	"time"
)

// BotManager manages multiple Bot instances.
type BotManager struct {
	bots sync.Map // map[botid]*Bot
}

func NewBotManager(bot []string) *BotManager {
	mgr := &BotManager{}
	for _, token := range bot {
		b := NewBot(token)
		mgr.bots.Store(b.BotId, b)
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
