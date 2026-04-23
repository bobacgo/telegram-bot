package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/telebot.v4"
)

const (
	StatusUsable  = iota // 可使用的 Bot
	StatusNetwork        // Bot网络异常（临时的）
	StatusBan            // Bot被封禁（检测结果）
	StatusClosed         // Bot已关闭（人为）
)

const (
	BotTypeNormal  = iota // 普通 Bot
	BotTypeChannel        // 频道 Bot
	BotTypeAlert          // 告警 Bot
)

type UserTopicInfo struct {
	UserID   int64
	Username string
	TopicID  int
	GroupID  int64
}

type Bot struct {
	tgBot  *telebot.Bot
	status atomic.Int32

	BotId    int64
	Username string
	DB       DB
	BotType  int

	// 健康检测相关
	failCount          int // sendMessage 连续失败次数
	getMeFailCount     int // getMe 连续失败次数
	lastHeartbeatMsgId int // 上一次心跳消息 ID，用于发送前删除

	// 用户ID -> topic映射
	userTopics sync.Map // map[int64]*UserTopicInfo
}

func NewBot(token string, db DB) *Bot {
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

	b := &Bot{
		tgBot:    bot,
		BotId:    bot.Me.ID,
		Username: bot.Me.Username,
		DB:       db,
	}
	// 恢复用户topic信息
	b.restoreUserTopics()
	return b
}

func (b *Bot) Start() {
	b.initHandlers()
	b.tgBot.Start()
}

func (b *Bot) Stop() {
	b.tgBot.Stop()
}

// IsHealthy 检测Bot是否可用
func (b *Bot) IsHealthy() bool {
	return b.status.Load() == StatusUsable
}

// IsReady 检测Bot是否准备好（可用或网络异常）
func (b *Bot) IsReady() bool {
	return b.status.Load() == StatusUsable || b.status.Load() == StatusNetwork
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

func (tb *Bot) SendMsg(ctx context.Context, data *SendMsgReq) (*telebot.Message, error) {
	slog.DebugContext(ctx, "Bot SendMsg", "msg", data)
	what, opts := data.makeMsg(data)
	return tb.tgBot.Send(&telebot.Chat{ID: data.ChatId}, what, opts)
}

// SendHeartbeat 发送心跳检测消息到监控群（静默模式）
// 返回发送的消息（用于后续删除）和错误
func (tb *Bot) SendHeartbeat(chatId int64, idx int) (*telebot.Message, error) {
	text := fmt.Sprintf("🤖 %d [%s] 心跳检测 - %s", idx, tb.Username, time.Now().Format("2006-01-02 15:04:05"))
	msg, err := tb.tgBot.Send(
		&telebot.Chat{ID: chatId},
		text,
		telebot.Silent, // 静默模式，不会产生通知
	)
	return msg, err
}

// DeleteMessage 删除消息
func (tb *Bot) DeleteMessage(chatId int64, msgId int) error {
	return tb.tgBot.Delete(&telebot.Message{
		ID:   msgId,
		Chat: &telebot.Chat{ID: chatId},
	})
}

type MediaType string

const (
	MediaTypeText      MediaType = "text"
	MediaTypePhoto     MediaType = "photo"
	MediaTypeAnimation MediaType = "animation"
	MediaTypeVideo     MediaType = "video"

	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
)

type SendMsgReq struct {
	BotUsername string    `json:"bot_username"`   // webappURL需要这个参数来生成链接
	ChatId      int64     `json:"chat_id"`        // 目标 Chat ID (包括用户、频道、群组等)
	MediaType   MediaType `json:"media_type"`     // 消息类型：text、photo、animation、video、audio、document
	Caption     string    `json:"caption"`        // 消息文本或媒体说明
	Url         string    `json:"url"`            // 媒体 URL，图片/视频/文件的链接地址
	Width       int32     `json:"width"`          // 媒体宽度（仅图片/视频有效），用于调整发送时的显示大小
	Height      int32     `json:"height"`         // 媒体高度（仅图片/视频有效），用于调整发送时的显示大小
	Duration    int32     `json:"duration"`       // 媒体时长（仅视频有效），单位秒，用于调整发送时的显示大小
	Btns        []*MsgBtn `json:"btns,omitempty"` // 消息按钮列表 telebot.ModeMarkdownV2
	ParseMode   string    `json:"parse_mode"`     // 文本解析模式：MarkdownV2、HTML等，影响Caption字段的解析方式
}

// MsgBtn 定义了消息按钮的结构
type MsgBtn struct {
	IsWebapp bool   `json:"is_webapp"`
	Unique   string `json:"unique"`
	Data     string `json:"data"`
	Text     string `json:"text"`
	Url      string `json:"url"`
}

func (msg *SendMsgReq) makeMsg(data *SendMsgReq) (any, *telebot.SendOptions) {
	var what any

	switch data.MediaType {
	case MediaTypePhoto:
		what = &telebot.Photo{
			File: telebot.File{
				FileURL: data.Url,
			},
			Caption: data.Caption,
			Width:   int(data.Width),
			Height:  int(data.Height),
		}
	case MediaTypeAnimation:
		what = &telebot.Animation{
			File: telebot.File{
				FileURL: data.Url,
			},
			Caption:  data.Caption,
			Width:    int(data.Width),
			Height:   int(data.Height),
			Duration: int(data.Duration),
		}
	case MediaTypeVideo:
		what = &telebot.Video{
			File: telebot.File{
				FileURL: data.Url,
			},
			Caption:  data.Caption,
			Width:    int(data.Width),
			Height:   int(data.Height),
			Duration: int(data.Duration),
		}
	case MediaTypeAudio:
		what = &telebot.Audio{
			File: telebot.File{
				FileURL: data.Url,
			},
			Caption:  data.Caption,
			Duration: int(data.Duration),
		}
	case MediaTypeDocument:
		what = &telebot.Document{
			File: telebot.File{
				FileURL: data.Url,
			},
			Caption: data.Caption,
		}
	default:
		what = data.Caption
	}

	menu := makeBtn(data.BotUsername, data.Btns)
	opts := &telebot.SendOptions{
		ParseMode:   data.ParseMode, // 启用 HTML 解析
		ReplyMarkup: menu,
	}
	return what, opts
}

func makeBtn(botUsername string, btns []*MsgBtn) *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	var rows []telebot.Row
	for _, v := range btns {
		var btn telebot.Btn
		if v.IsWebapp {
			btn = menu.WebApp(v.Text, &telebot.WebApp{URL: genWebappURL(v.Url, botUsername)})
		} else {
			btn = menu.URL(v.Text, v.Url)
		}
		btn.Unique = v.Unique
		btn.Data = v.Data

		rows = append(rows, menu.Row(btn)) // 每个按钮一行
	}

	if len(rows) > 0 {
		menu.Inline(rows...) // 一次性设置所有行
	}
	return menu
}

// https://xxx?bot_name=lances_bot
func genWebappURL(rawURL, botName string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		slog.Error("url parse error", "url", rawURL, "err", err)
		return rawURL
	}
	params := u.Query()
	params.Set("bot_name", botName)
	u.RawQuery = params.Encode()
	return u.String()
}
