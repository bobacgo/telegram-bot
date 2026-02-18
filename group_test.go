package main

import (
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gopkg.in/telebot.v4"
)

func TestGroup(t *testing.T) {
	token := "8441906451:AAGMpRGiyFi3HRe-06cfchlqKf8pmlS-OdA"
	var group_id int64 = -1003563520720
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	pref := telebot.Settings{
		Token: token,
		// Verbose: true,
		OnError: func(err error, c telebot.Context) {
			slog.Error("telegram bot error", "err", err, "bot", c.Bot())
		},
		Client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("failed to create bot, bot_id:%s err: %v", token, err)
	}
	topic, err := bot.CreateTopic(&telebot.Chat{ID: group_id}, &telebot.Topic{Name: "test topic"})
	if err != nil {
		log.Fatalf("failed to create topic, err: %v", err)
	}
	log.Printf("topic created: %+v", topic)
}

func TestSendToGroup(t *testing.T) {
	token := "8441906451:AAGMpRGiyFi3HRe-06cfchlqKf8pmlS-OdA"
	var group_id int64 = -1003563520720
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")

	pref := telebot.Settings{
		Token: token,
		OnError: func(err error, c telebot.Context) {
			slog.Error("telegram bot error", "err", err, "bot", c.Bot())
		},
		Client: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatalf("failed to create bot: %v", err)
	}

	// 创建一个 topic (群话题)
	topic, err := bot.CreateTopic(&telebot.Chat{ID: group_id}, &telebot.Topic{Name: "test topic"})
	if err != nil {
		log.Fatalf("failed to create topic: %v", err)
	}
	log.Printf("✅ topic created: %+v\n", topic)

	// 向指定话题发送消息
	msg := &telebot.Message{
		Text: "Hello from test topic!",
		Chat: &telebot.Chat{ID: group_id},
	}

	// 使用 Subject 方法指定要发送的话题 ID
	sendMsg := &telebot.SendOptions{
		ThreadID: topic.ThreadID,
	}

	result, err := bot.Send(&telebot.Chat{ID: group_id}, msg.Text, sendMsg)
	if err != nil {
		log.Fatalf("failed to send message to topic: %v", err)
	}
	log.Printf("✅ message sent to topic: message_id=%d\n", result.ID)

	// 验证：可以再发送一条消息
	result2, err := bot.Send(&telebot.Chat{ID: group_id}, "Second message in topic", sendMsg)
	if err != nil {
		log.Fatalf("failed to send second message: %v", err)
	}
	log.Printf("✅ second message sent: message_id=%d\n", result2.ID)
}
