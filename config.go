package main

import (
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

const DomainMe = "https://t.me"

type Config struct {
	Proxy    ProxyConfig    `yaml:"proxy"`
	Bots     []BotConfig    `yaml:"bot"`
	Customer CustomerConfig `yaml:"customer"`
	DBs      []DBConfig     `yaml:"db"`
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

type DBConfig struct {
	Path               string `yaml:"path"`                 // 存储文件路径
	SyncOnWrite        bool   `yaml:"sync_on_write"`        // 每次写入后立即同步到磁盘
	SyncThreshold      int    `yaml:"sync_threshold"`       // 触发 fsync 的操作次数阈值
	CompactDeleteCount int    `yaml:"compact_delete_count"` // 触发压缩的删除次数阈值
	CompactCooldown    int    `yaml:"compact_cooldown"`     // 压缩冷却时间（秒）
	SyncCooldown       int    `yaml:"sync_cooldown"`        // 同步冷却时间（秒）
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
