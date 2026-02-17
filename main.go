package main

import "log"

func main() {
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	tokens := cfg.BotTokens()
	if len(tokens) == 0 {
		log.Fatal("no bot tokens found in config")
	}

	SetProxyConfig(cfg.Proxy)
	SetCustomerGroupIDs(cfg.CustomerChatIDs())

	mgr := NewBotManager(tokens)
	mgr.Start()
	defer mgr.Stop()

	select {}
}
