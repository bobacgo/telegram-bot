package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"gopkg.in/telebot.v4"
)

func TestGetChatInfo(t *testing.T) {
	//token := "8290090910:AAHlsOp1479LSxY0w3q5m1ZE8PlNXd3noJE" // 频道管理 Bot
	token := "8436918284:AAHl5o__uhllcYI6ucit6tQCb30Xt7OY-CQ" // 群管理 Bot
	proxyURL, _ := url.Parse("http://127.0.0.1:7890")
	pref := telebot.Settings{
		Token: token,
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
	//chanInfo, err := bot.ChatByID(-1003803025720) // 公开频道
	//chanInfo, err := bot.ChatByID(-1002073613488) // 私有频道
	chanInfo, err := bot.ChatByID(-1003955399688) // 私有频道
	if err != nil {
		log.Fatalf("failed to get chat id, bot_id:%s err: %v", token, err)
	}

	res, _ := json.MarshalIndent(chanInfo, "", "  ")
	fmt.Printf("%s\n", string(res))
}

func TestNilSlice(t *testing.T) {
	var arr []int = nil
	var arrTs = func(arr ...int) {
		log.Printf("arr:%v", arr)
	}
	arrTs(arr...)
}
