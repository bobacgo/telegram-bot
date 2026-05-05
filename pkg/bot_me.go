package pkg

import (
	"fmt"
	"log/slog"

	"gopkg.in/telebot.v4"
)

// getMe 用于验证 token 是否有效，以及获取 Bot 的基本信息
func BotGetMe(token string, proxyUrl string) (*telebot.User, error) {
	pref := telebot.Settings{
		Token:  token,
		Client: HttpClt(proxyUrl),
		OnError: func(err error, c telebot.Context) {
			var (
				senderId    int64
				chatId      int64
				recipientId string
				text        string
			)
			if c != nil {
				senderId = c.Sender().ID
				chatId = c.Chat().ID
				text = c.Text()
				if c.Recipient() != nil {
					recipientId = c.Recipient().Recipient()
				}
			}

			slog.Error("telegram bot error", "sender_id", senderId, "chat_id", chatId, "recipient_id", recipientId, "text", text, "err", err)
		},
	}
	botClient, err := telebot.NewBot(pref)
	if err != nil {
		slog.Error("telegram bot New error", "err", err)
		return nil, err
	}
	if !botClient.Me.IsBot {
		return nil, fmt.Errorf("bot.Me: not bot")
	}

	return botClient.Me, nil
}
