package main

import (
	"fmt"
	"log"
	"log/slog"

	"gopkg.in/telebot.v4"
)

type Bot struct {
	tgBot *telebot.Bot

	BotId    int64
	Username string
}

func NewBot(token string) *Bot {
	// TODO : webhook support

	pref := telebot.Settings{
		Token:   token,
		Offline: false,
		// Verbose: true,
		OnError: func(err error, c telebot.Context) {
			slog.Error("telegram bot error", "err", err, "bot", c.Bot())
		},
		Client: HttpClient(),
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("failed to create bot, bot_id:%s err: %v", token, err)
	}
	return &Bot{
		tgBot:    bot,
		BotId:    bot.Me.ID,
		Username: bot.Me.Username,
	}
}

func (b *Bot) Start() {
	b.initHandlers()
	b.tgBot.Start()
}

func (b *Bot) Stop() {
	b.tgBot.Stop()
}

func (b *Bot) initHandlers() {
	// TODO : add bot handlers

	b.tgBot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Send("Welcome! This is a Telegram bot.")
	})

	b.tgBot.Handle("/cid", func(ctx telebot.Context) error {
		return ctx.Send(fmt.Sprintf("Your chat ID is: %d", ctx.Chat().ID))
	})

	// b.tgBot.Handle(telebot.OnText, func(c telebot.Context) error {
	// 	slog.Info("received message", "from", c.Sender().Username, "text", c.Text())
	// 	return c.Send("Hello! This is an automated response.")
	// })

	b.tgBot.Handle(telebot.OnText, b.OnText)
}

func (b *Bot) GetChatById(chatId int64) (*telebot.Chat, error) {
	chat, err := b.tgBot.ChatByID(chatId)
	if err != nil {
		slog.Error("failed to get chat by id", "chat_id", chatId, "err", err)
		return nil, err
	}
	return chat, nil
}

func (b *Bot) GetChatUsername(username string) (*telebot.Chat, error) {
	chat, err := b.tgBot.ChatByUsername(username)
	if err != nil {
		slog.Error("failed to get chat by username", "username", username, "err", err)
		return nil, err
	}
	return chat, nil
}

func (b *Bot) GetChatMember(chatId, userId int64) (*telebot.ChatMember, error) {
	member, err := b.tgBot.ChatMemberOf(&telebot.Chat{ID: chatId}, &telebot.User{ID: userId})
	if err != nil {
		slog.Error("failed to get chat member", "chat_id", chatId, "user_id", userId, "err", err)
		return nil, err
	}
	return member, nil
}
