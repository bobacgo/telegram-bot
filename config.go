package main

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Proxy    ProxyConfig    `yaml:"proxy"`
	Bots     []BotConfig    `yaml:"bot"`
	Customer CustomerConfig `yaml:"customer"`
}

type ProxyConfig struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
}

type BotConfig struct {
	Token string `yaml:"token"`
}

type CustomerConfig struct {
	SessionLimit int             `yaml:"session_limit"`
	Groups       []CustomerGroup `yaml:"groups"`
}

type CustomerGroup struct {
	ChatID int64 `yaml:"chat_id"`
}

func (c *CustomerConfig) GroupChatIDs() []int64 {
	res := make([]int64, 0, len(c.Groups))
	for _, g := range c.Groups {
		if g.ChatID != 0 && !slices.Contains(res, g.ChatID) {
			res = append(res, g.ChatID)
		}
	}
	slices.Sort(res)
	return res
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
