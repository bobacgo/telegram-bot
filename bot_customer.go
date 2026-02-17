package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/telebot.v4"
)

var customerGroupIDs []int64

func SetCustomerGroupIDs(ids []int64) {
	customerGroupIDs = ids
}

// CustomerBot 是一个专门用于客服的 Telegram 机器人
func (b *Bot) OnText(c telebot.Context) error {
	slog.Info("received message", "from", c.Sender().Username, "chat_id", c.Chat().ID, "text", c.Text())

	// 1. 用户私聊过来的消息 → 转发到客服群
	if c.Chat().Type == telebot.ChatPrivate {
		if len(customerGroupIDs) == 0 {
			slog.Warn("no customer groups configured")
			return nil
		}
		user, msg := c.Sender(), c.Message()

		caption := fmt.Sprintf("%s (%d): %s", user.Username, user.ID, msg.Text)
		_, err := b.tgBot.Send(&telebot.Chat{ID: b.getCustomerGroupID(user.ID)}, caption)
		return err
	}

	// 2. 客服群里消息，并且是回复消息 → 转发回用户
	if slices.Contains(customerGroupIDs, c.Chat().ID) && c.Message().ReplyTo != nil {
		userID := b.toUserID(c.Message().ReplyTo.Text)
		if userID == 0 {
			return nil // 没识别出用户ID
		}

		// 把客服的消息转发给对应用户
		_, err := b.tgBot.Send(&telebot.Chat{ID: userID}, c.Text())
		return err
	}
	return nil
}

func (b *Bot) toUserID(text string) int64 {
	// 从文本中提取用户ID
	// 假设文本格式为 "用户名 (ID): 消息内容"
	// 提取 "ID" 部分
	startIdx := strings.LastIndex(text, "(")
	endIdx := strings.LastIndex(text, ")")
	if startIdx == -1 || endIdx == -1 || endIdx == len(text)-1 {
		return 0
	}
	idStr := text[startIdx+1 : endIdx]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		slog.Error("parse userID failed", "idStr", idStr, "err", err)
		return 0
	}
	return id
}

func (b *Bot) getCustomerGroupID(userID int64) int64 {
	if len(customerGroupIDs) == 1 {
		return customerGroupIDs[0]
	}
	idx := int(userID % int64(len(customerGroupIDs)))
	if idx < 0 {
		idx = -idx
	}
	return customerGroupIDs[idx]
}
