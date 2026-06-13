package main

import (
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	tb "gopkg.in/telebot.v4"
)

func TestReplyMarKup(t *testing.T) {

	token := "8441906451:AAGMpRGiyFi3HRe-06cfchlqKf8pmlS-OdA" // 群管理 Bot
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	pref := tb.Settings{
		Token: token,
		Client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	}
	b, err := tb.NewBot(pref)

	if err != nil {
		log.Fatal(err)
	}

	// 创建回复键盘
	menu := &tb.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: false,
	}

	// 按钮
	btnCurrentBot := menu.Text("🤖 Актуальный бот")

	// 布局
	menu.Reply(
		menu.Row(btnCurrentBot),
	)

	b.Handle("/start", func(c tb.Context) error {
		return c.Send(
			"请选择功能",
			menu,
		)
	})

	// 监听按钮点击
	b.Handle(&btnCurrentBot, func(c tb.Context) error {
		return c.Send("你点击了当前机器人")
	})

	b.Start()
}
