package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/telebot.v4"
)

var customerConfig CustomerConfig

const userTopicKey = "user_topic"

func SetCustomerConfig(cfg CustomerConfig) {
	customerConfig = cfg
}

// CustomerBot 是一个专门用于客服的 Telegram 机器人
// 为每个用户创建专属的topic，在topic中进行对话
func (b *Bot) OnText(c telebot.Context) error {
	slog.Info("received message", "from", c.Sender().Username, "chat_id", c.Chat().ID, "text", c.Text())

	// 1. 用户私聊过来的消息 → 创建topic并转发到客服群
	if c.Chat().Type == telebot.ChatPrivate {
		if len(customerConfig.GroupChatIDs()) == 0 {
			slog.Warn("no customer groups configured")
			return nil
		}
		user, msg := c.Sender(), c.Message()
		groupID := b.getCustomerGroupID(user.ID)

		// If the user doesn't have a topic yet, enforce the session limit.
		if _, ok := b.userTopics.Load(user.ID); !ok && customerConfig.SessionLimit > 0 {
			if b.groupSessionCount(groupID) >= customerConfig.SessionLimit {
				warnMsg := "The customer service session is full. Please try again later."
				if err := c.Send(warnMsg); err != nil {
					slog.Error("failed to send session limit message", "user_id", user.ID, "err", err)
				}
				return nil
			}
		}

		// 检查该用户是否已有topic
		userTopic := b.getOrCreateUserTopic(user.ID, user.Username, groupID)
		if userTopic == nil {
			slog.Error("failed to get or create topic for user", "user_id", user.ID, "username", user.Username)
			return nil
		}

		// 向topic中发送消息
		msgText := fmt.Sprintf("%s: %s", user.Username, msg.Text)
		sendOpts := &telebot.SendOptions{ThreadID: userTopic.TopicID}
		_, err := b.tgBot.Send(&telebot.Chat{ID: groupID}, msgText, sendOpts)
		if err != nil {
			slog.Error("failed to send message to topic", "user_id", user.ID, "topic_id", userTopic.TopicID, "err", err)
		}
		return err
	}

	// 2. 客服群里的topic消息 → 转发回用户
	if slices.Contains(customerConfig.GroupChatIDs(), c.Chat().ID) && c.Message().ThreadID != 0 {
		// 在topic中，获取来自topic reply的消息
		userTopic := b.getUserTopicByThreadID(c.Chat().ID, c.Message().ThreadID)
		if userTopic == nil {
			slog.Warn("topic not found", "chat_id", c.Chat().ID, "thread_id", c.Message().ThreadID)
			return nil
		}

		// 忽略bot自己的消息
		if c.Sender().ID == b.BotId {
			return nil
		}

		// 把客服的消息转发给对应用户
		_, err := b.tgBot.Send(&telebot.Chat{ID: userTopic.UserID}, c.Text())
		if err != nil {
			slog.Error("failed to send message to user", "user_id", userTopic.UserID, "err", err)
		}
		return err
	}
	return nil
}

// getOrCreateUserTopic 获取或创建用户的topic
func (b *Bot) getOrCreateUserTopic(userID int64, username string, groupID int64) *UserTopicInfo {
	// 检查是否已存在
	if val, ok := b.userTopics.Load(userID); ok {
		if topic, ok := val.(*UserTopicInfo); ok {
			return topic
		}
	}

	// 创建新topic
	topicName := fmt.Sprintf("%s (ID: %d)", username, userID)
	topic, err := b.tgBot.CreateTopic(&telebot.Chat{ID: groupID}, &telebot.Topic{Name: topicName})
	if err != nil {
		slog.Error("failed to create topic", "user_id", userID, "username", username, "err", err)
		return nil
	}

	userTopic := &UserTopicInfo{
		UserID:   userID,
		Username: username,
		TopicID:  topic.ThreadID,
		GroupID:  groupID,
	}

	b.userTopics.Store(userID, userTopic)
	b.saveUserTopic(userTopic)
	slog.Info("topic created for user", "user_id", userID, "username", username, "topic_id", topic.ThreadID, "group_id", groupID)
	return userTopic
}

func (b *Bot) userTopicStoreKey(userID int64) string {
	return fmt.Sprintf("%d:%d", b.BotId, userID)
}

func parseUserTopicStoreKey(key string) (botID int64, userID int64, ok bool) {
	parts := strings.Split(key, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	if parts[0] == "" || parts[1] == "" {
		return 0, 0, false
	}

	bid, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	uid, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, false
	}
	return bid, uid, true
}

func (b *Bot) saveUserTopic(topic *UserTopicInfo) {
	raw, err := json.Marshal(topic)
	if err != nil {
		slog.Error("failed to marshal topic", "user_id", topic.UserID, "err", err)
		return
	}

	if err := b.DB[userTopicKey].Set(b.userTopicStoreKey(topic.UserID), raw); err != nil {
		slog.Error("failed to persist topic", "user_id", topic.UserID, "err", err)
	}
}

// restoreUserTopics 从存储中恢复用户topic信息
func (b *Bot) restoreUserTopics() {
	if err := b.DB[userTopicKey].ForEach(func(key string, value []byte) error {
		botID, userID, ok := parseUserTopicStoreKey(key)
		if !ok || botID != b.BotId {
			return nil
		}

		topic := &UserTopicInfo{}
		if err := json.Unmarshal(value, topic); err != nil {
			slog.Warn("failed to decode persisted topic", "key", key, "err", err)
			return nil
		}
		if topic.UserID == 0 {
			topic.UserID = userID
		}

		b.userTopics.Store(topic.UserID, topic)
		return nil
	}); err != nil {
		slog.Error("failed to restore topics", "bot_id", b.BotId, "err", err)
		return
	}

	slog.Info("restored user topics", "bot_id", b.BotId, "count", b.userTopicCount())
}

// getUserTopicByThreadID 根据group和thread ID获取用户topic信息
func (b *Bot) getUserTopicByThreadID(groupID int64, threadID int) *UserTopicInfo {
	var result *UserTopicInfo
	b.userTopics.Range(func(key, value any) bool {
		if topic, ok := value.(*UserTopicInfo); ok {
			if topic.GroupID == groupID && topic.TopicID == threadID {
				result = topic
				return false // 找到了，停止遍历
			}
		}
		return true
	})
	return result
}

func (b *Bot) groupSessionCount(groupID int64) int {
	count := 0
	b.userTopics.Range(func(key, value any) bool {
		if topic, ok := value.(*UserTopicInfo); ok {
			if topic.GroupID == groupID {
				count++
			}
		}
		return true
	})
	return count
}

func (b *Bot) userTopicCount() int {
	count := 0
	b.userTopics.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}

// 根据群数量和用户ID计算分配到哪个客服群
// 这里使用简单的取模算法，确保同一个用户始终分配到同一个群
func (b *Bot) getCustomerGroupID(userID int64) int64 {
	gids := customerConfig.GroupChatIDs()
	if len(gids) == 1 {
		return gids[0]
	}
	idx := int(userID % int64(len(gids)))
	if idx < 0 {
		idx = -idx
	}
	return gids[idx]
}
