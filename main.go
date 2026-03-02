package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// load config
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	tokens := cfg.BotTokens()
	if len(tokens) == 0 {
		log.Fatal("no bot tokens found in config")
	}

	SetProxyConfig(cfg.Proxy)
	SetCustomerConfig(cfg.Customer)

	// load storage
	db, err := loadDB(cfg.DBs)
	if err != nil {
		log.Fatalf("failed to load database: %v", err)
	}

	// start bots
	mgr := NewBotManager(tokens, db)
	mgr.Start()

	slog.Info("bots started successfully", "bot_count", len(tokens))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down gracefully...")
	mgr.Stop()
	if err := db.Close(); err != nil {
		slog.Error("failed to close database", "error", err)
	}
	slog.Info("shutdown complete")
}

func loadDB(cfgs []DBConfig) (DB, error) {
	var kvs []KVStore
	for _, cfg := range cfgs {
		fkv, err := NewFileKVStore(cfg.Path, FileKVStoreOptions{
			SyncOnWrite:        cfg.SyncOnWrite,
			SyncThreshold:      cfg.SyncThreshold,
			CompactDeleteCount: cfg.CompactDeleteCount,
			CompactCooldown:    time.Duration(cfg.CompactCooldown) * time.Second,
			SyncCooldown:       time.Duration(cfg.SyncCooldown) * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("init KVStore %q: %w", filepath.Base(cfg.Path), err)
		}
		kvs = append(kvs, fkv)
	}
	db := NewDB(kvs)
	return db, nil
}
