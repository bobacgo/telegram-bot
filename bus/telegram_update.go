package bus

import "gopkg.in/telebot.v4"

type TgUpdateEvent struct {
	BotID  int64
	Update *telebot.Update
}

func (b *Bus) InWebhook() chan<- *TgUpdateEvent {
	return b.webhookChan
}

func (b *Bus) OutWebhook() <-chan *TgUpdateEvent {
	return b.webhookChan
}
