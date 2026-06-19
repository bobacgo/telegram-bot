package bot

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

func TestInChannel(t *testing.T) {
	token := "8290090910:AAHlsOp1479LSxY0w3q5m1ZE8PlNXd3noJE" // 群管理 Bot
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
	//chid := -1003679108610
	var chid int64 = -1003803025720

	chanInfo, err := bot.ChatMemberOf(&telebot.Chat{ID: chid}, &telebot.User{ID: 8238467092}) // 私有频道
	if err != nil {
		log.Fatalf("failed to get chat id, bot_id:%s err: %v", token, err)
	}

	res, _ := json.MarshalIndent(chanInfo, "", "  ")
	fmt.Printf("%s\n", string(res))
}
