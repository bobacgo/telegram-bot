package dto

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

type ChannelCreateReq struct {
	TgChannelId int64  `json:"tg_channel_id"`
	Title       string `json:"title"`
	Username    string `json:"username"`
	Owner       string `json:"owner"`
	Type        int    `json:"type"`
	Status      int    `json:"status"`
}

func (req *ChannelCreateReq) Validate() error {
	if req.TgChannelId == 0 {
		return fmt.Errorf("tg_channel_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if req.Status == 0 {
		req.Status = 1
	}
	return nil
}

type GroupCreateReq struct {
	TgGroupId int64  `json:"tg_group_id"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	Owner     string `json:"owner"`
	Type      int    `json:"type"`
	Status    int    `json:"status"`
}

func (req *GroupCreateReq) Validate() error {
	if req.TgGroupId == 0 {
		return fmt.Errorf("tg_group_id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if req.Status == 0 {
		req.Status = 1
	}
	return nil
}

type TopicCreateReq struct {
	TgGroupId int64  `json:"tg_group_id"`
	TopicId   int64  `json:"topic_id"`
	Name      string `json:"name"`
}

func (req *TopicCreateReq) Validate() error {
	if req.TgGroupId == 0 {
		return fmt.Errorf("tg_group_id is required")
	}
	if req.TopicId == 0 {
		return fmt.Errorf("topic_id is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}
