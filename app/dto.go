package app

import (
	"fmt"
	"strings"
)

type BotCreateReq struct {
	Token         string `json:"token"`
	WebhookSecret string `json:"webhook_secret"`
	Owner         string `json:"owner"`
	Type          int    `json:"type"`
	Status        int    `json:"status"`
}

func (req *BotCreateReq) Validate() error {
	if strings.TrimSpace(req.Token) == "" {
		return fmt.Errorf("token is required")
	}

	if req.WebhookSecret == "" {
		// 如果没有提供 webhook secret，则默认使用 token 的前半部分作为 webhook secret
		req.WebhookSecret = req.Token[:len(req.Token)/2]
	}

	if req.Status == 0 {
		req.Status = 1 // 默认启用
	}
	return nil
}
