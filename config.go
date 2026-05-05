package main

import (
	"bot/pkg"
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

const DomainMe = "https://t.me"

type Config struct {
	Proxy    pkg.ProxyConfig `yaml:"proxy"`
	AdminBot AdminBotConfig  `yaml:"admin_bot"`
	Customer CustomerConfig  `yaml:"customer"`
	DBs      []DBConfig      `yaml:"db"`
	BizBots  []BizBot        `yaml:"biz_bots"` // 这个以后从其他地方拿
}

type AdminBotConfig struct {
	AdminBot []*AdminBot `yaml:"bot"` // 管理员机器人
}

type AdminBot struct {
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

type BotKind int

const (
	BotKindDefault  BotKind = 1 // 默认
	BotKindCustomer BotKind = 2 // 客服 bot
)

type BizBot struct {
	Kind          BotKind `yaml:"kind"`           // 机器人用途
	Token         string  `yaml:"token"`          // 机器人 token
	WebhookSecret string  `yaml:"webhook_secret"` // Webhook 验证密钥
	Status        int     `yaml:"status"`         // 机器人状态，1=正常，2=关闭
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
	res := make([]string, 0, len(c.BizBots))
	for _, b := range c.BizBots {
		if b.Token != "" && !slices.Contains(res, b.Token) {
			res = append(res, b.Token)
		}
	}
	return res
}
