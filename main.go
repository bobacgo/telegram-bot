package main

import "log"

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	store, err := NewFileKVStore("data/kv.log", FileKVStoreOptions{SyncOnWrite: true})
	if err != nil {
		log.Fatalf("failed to init kv store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("failed to close kv store: %v", err)
		}
	}()

	tokens := cfg.BotTokens()
	if len(tokens) == 0 {
		log.Fatal("no bot tokens found in config")
	}

	SetProxyConfig(cfg.Proxy)
	SetCustomerConfig(cfg.Customer)

	mgr := NewBotManager(tokens, store)
	mgr.Start()
	defer mgr.Stop()

	select {}
}
