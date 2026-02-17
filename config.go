package main

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy     ProxyConfig      `yaml:"proxy"`
	Bots      []BotConfig      `yaml:"bot"`
	Customers []CustomerConfig `yaml:"customer"`
}

type ProxyConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

type CustomerConfig struct {
	ChatID int64 `yaml:"chat_id"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (c *Config) BotTokens() []string {
	if c == nil {
		return nil
	}
	res := make([]string, 0, len(c.Bots))
	for _, b := range c.Bots {
		if b.Token != "" && !slices.Contains(res, b.Token) {
			res = append(res, b.Token)
		}
	}
	return res
}

func (c *Config) CustomerChatIDs() []int64 {
	if c == nil {
		return nil
	}
	res := make([]int64, 0, len(c.Customers))
	for _, v := range c.Customers {
		if v.ChatID != 0 && !slices.Contains(res, v.ChatID) {
			res = append(res, v.ChatID)
		}
	}
	slices.Sort(res)
	return res
}
