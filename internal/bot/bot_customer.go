package bot

import (
	"bot/internal/repo"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/telebot.v4"
)

const customerRecordRoot = "data/customer"

type customerChatRecord struct {
	At        int64  `json:"at"`
	Direction string `json:"direction"`
	BotTgId   int64  `json:"bot_tg_id"`
	GroupId   int64  `json:"group_id"`
	TopicId   int    `json:"topic_id"`
	UserId    int64  `json:"user_id"`
	Username  string `json:"username"`
	SenderId  int64  `json:"sender_id"`
	Sender    string `json:"sender"`
	MessageId int    `json:"message_id"`
	Text      string `json:"text"`
}

// OnText 处理客服 bot 的用户私聊和客服群 topic 回复。
func (b *Bot) OnText(c telebot.Context) error {
	if c == nil || c.Message() == nil || c.Chat() == nil || c.Sender() == nil {
		return nil
	}

	switch {
	case c.Chat().Type == telebot.ChatPrivate:
		return b.onCustomerUserText(c)
	case c.Message().ThreadID != 0:
		return b.onCustomerGroupText(c)
	default:
		return nil
	}
}

func (b *Bot) onCustomerUserText(c telebot.Context) error {
	user, msg := c.Sender(), c.Message()
	groupID, err := b.customerGroupID(user.ID)
	if err != nil {
		slog.Warn("no customer groups configured", "err", err)
		return c.Send("Customer service is temporarily unavailable.")
	}

	session, err := b.getOrCreateUserTopic(user.ID, user.Username, groupID)
	if err != nil {
		slog.Error("failed to get or create customer session", "user_id", user.ID, "username", user.Username, "err", err)
		return err
	}

	name := displayName(user)
	text := fmt.Sprintf("%s: %s", name, msg.Text)
	_, err = b.tgBot.Send(&telebot.Chat{ID: session.GroupID}, text, &telebot.SendOptions{ThreadID: session.TopicID})
	if err != nil {
		slog.Error("failed to send customer message to topic", "user_id", user.ID, "topic_id", session.TopicID, "err", err)
		return err
	}

	b.appendCustomerRecord(session, customerChatRecord{
		At:        time.Now().Unix(),
		Direction: "user_to_service",
		BotTgId:   b.cfg.BotTgId,
		GroupId:   session.GroupID,
		TopicId:   session.TopicID,
		UserId:    user.ID,
		Username:  user.Username,
		SenderId:  user.ID,
		Sender:    name,
		MessageId: msg.ID,
		Text:      msg.Text,
	})
	b.touchCustomerSession(session)
	return nil
}

func (b *Bot) onCustomerGroupText(c telebot.Context) error {
	if c.Sender().ID == b.cfg.BotTgId {
		return nil
	}

	msg := c.Message()
	session := b.getUserTopicByThreadID(c.Chat().ID, msg.ThreadID)
	if session == nil {
		slog.Debug("customer topic not found", "chat_id", c.Chat().ID, "thread_id", msg.ThreadID)
		return nil
	}

	_, err := b.tgBot.Send(&telebot.Chat{ID: session.UserID}, c.Text())
	if err != nil {
		slog.Error("failed to send customer reply to user", "user_id", session.UserID, "err", err)
		return err
	}

	b.appendCustomerRecord(session, customerChatRecord{
		At:        time.Now().Unix(),
		Direction: "service_to_user",
		BotTgId:   b.cfg.BotTgId,
		GroupId:   session.GroupID,
		TopicId:   session.TopicID,
		UserId:    session.UserID,
		Username:  session.Username,
		SenderId:  c.Sender().ID,
		Sender:    displayName(c.Sender()),
		MessageId: msg.ID,
		Text:      c.Text(),
	})
	b.touchCustomerSession(session)
	return nil
}

func (b *Bot) getOrCreateUserTopic(userID int64, username string, groupID int64) (*UserTopicInfo, error) {
	if val, ok := b.userTopics.Load(userID); ok {
		if topic, ok := val.(*UserTopicInfo); ok {
			return topic, nil
		}
	}

	session, err := b.repo.CustomerSession.FindOne(context.Background(), repo.CustomerSessionFindOneReq{
		BotTgId: b.cfg.BotTgId,
		UserId:  userID,
		Status:  repo.CustomerSessionStatusOpen,
	})
	if err == nil {
		info := sessionToTopicInfo(session)
		b.userTopics.Store(userID, info)
		return info, nil
	}

	topicName := fmt.Sprintf("%s (ID: %d)", fallbackUsername(username, userID), userID)
	topic, err := b.tgBot.CreateTopic(&telebot.Chat{ID: groupID}, &telebot.Topic{Name: topicName})
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	session = &repo.CustomerSession{
		BotTgId:   b.cfg.BotTgId,
		UserId:    userID,
		Username:  username,
		TgGroupId: groupID,
		TopicId:   topic.ThreadID,
		Status:    repo.CustomerSessionStatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := b.repo.CustomerSession.Insert(context.Background(), session); err != nil {
		if existing, findErr := b.repo.CustomerSession.FindOne(context.Background(), repo.CustomerSessionFindOneReq{
			BotTgId: b.cfg.BotTgId,
			UserId:  userID,
			Status:  repo.CustomerSessionStatusOpen,
		}); findErr == nil {
			info := sessionToTopicInfo(existing)
			b.userTopics.Store(userID, info)
			return info, nil
		}
		return nil, err
	}

	_ = b.repo.GroupTopic.Insert(context.Background(), &repo.TelegramGroupTopic{
		TgGroupId: groupID,
		TopicId:   int64(topic.ThreadID),
		Name:      topicName,
		CreatedAt: now,
		UpdatedAt: now,
	})

	info := sessionToTopicInfo(session)
	b.userTopics.Store(userID, info)
	slog.Info("customer topic created", "bot_id", b.cfg.BotTgId, "user_id", userID, "group_id", groupID, "topic_id", topic.ThreadID)
	return info, nil
}

func (b *Bot) restoreUserTopics() {
	if b.cfg.Type != BotTypeCustomer || b.repo == nil || b.repo.CustomerSession == nil {
		return
	}

	rows, err := b.repo.CustomerSession.List(context.Background(), repo.CustomerSessionFindOneReq{
		BotTgId: b.cfg.BotTgId,
		Status:  repo.CustomerSessionStatusOpen,
	})
	if err != nil {
		slog.Error("failed to restore customer sessions", "bot_id", b.cfg.BotTgId, "err", err)
		return
	}
	for _, row := range rows {
		info := sessionToTopicInfo(row)
		b.userTopics.Store(info.UserID, info)
	}
	slog.Info("restored customer sessions", "bot_id", b.cfg.BotTgId, "count", len(rows))
}

func (b *Bot) getUserTopicByThreadID(groupID int64, threadID int) *UserTopicInfo {
	var result *UserTopicInfo
	b.userTopics.Range(func(key, value any) bool {
		topic, ok := value.(*UserTopicInfo)
		if ok && topic.GroupID == groupID && topic.TopicID == threadID {
			result = topic
			return false
		}
		return true
	})
	return result
}

func (b *Bot) customerGroupID(userID int64) (int64, error) {
	groups, err := b.repo.Group.List(context.Background(), &repo.TelegramGroupQuery{Status: repo.GroupStatusUsable})
	if err != nil {
		return 0, err
	}
	if len(groups) == 0 {
		return 0, errors.New("no enabled customer groups")
	}

	idx := int(userID % int64(len(groups)))
	if idx < 0 {
		idx = -idx
	}
	return groups[idx].TgGroupId, nil
}

func (b *Bot) appendCustomerRecord(session *UserTopicInfo, record customerChatRecord) {
	path := b.customerRecordPath(session)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		slog.Error("failed to create customer record dir", "path", path, "err", err)
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.Error("failed to open customer record file", "path", path, "err", err)
		return
	}
	defer f.Close()

	raw, err := json.Marshal(record)
	if err != nil {
		slog.Error("failed to marshal customer record", "err", err)
		return
	}
	if _, err := f.Write(append(raw, '\n')); err != nil {
		slog.Error("failed to write customer record", "path", path, "err", err)
	}
}

func (b *Bot) touchCustomerSession(session *UserTopicInfo) {
	if session == nil {
		return
	}
	if err := b.repo.CustomerSession.Touch(context.Background(), b.cfg.BotTgId, session.UserID, time.Now().Unix()); err != nil {
		slog.Error("failed to touch customer session", "bot_id", b.cfg.BotTgId, "user_id", session.UserID, "err", err)
	}
}

func (b *Bot) customerRecordPath(session *UserTopicInfo) string {
	groupFolder := strconv.FormatInt(session.GroupID, 10)
	fileName := fmt.Sprintf("%d_%d_%d.jsonl", b.cfg.BotTgId, session.UserID, session.TopicID)
	return filepath.Join(customerRecordRoot, groupFolder, fileName)
}

func sessionToTopicInfo(session *repo.CustomerSession) *UserTopicInfo {
	return &UserTopicInfo{
		UserID:   session.UserId,
		Username: session.Username,
		TopicID:  session.TopicId,
		GroupID:  session.TgGroupId,
	}
}

func fallbackUsername(username string, userID int64) string {
	if username != "" {
		return username
	}
	return strconv.FormatInt(userID, 10)
}

func displayName(user *telebot.User) string {
	if user == nil {
		return ""
	}
	if user.Username != "" {
		return "@" + user.Username
	}
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	if name != "" {
		return name
	}
	return strconv.FormatInt(user.ID, 10)
}
