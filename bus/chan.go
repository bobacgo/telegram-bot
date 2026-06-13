package bus

import "gopkg.in/telebot.v4"

type Bus struct {
	cfgChan     chan any
	webhookChan chan *TgUpdateEvent
}

func NewBus() *Bus {
	return &Bus{
		cfgChan:     make(chan any, 1),
		webhookChan: make(chan *TgUpdateEvent, 1000),
	}
}

func (b *Bus) Start() {
	// No need to do anything here, channels are already initialized.
}

func (b *Bus) Stop() {
	close(b.cfgChan)
	close(b.webhookChan)
}

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

func (b *Bus) InConfig() chan<- any {
	return b.cfgChan
}

func (b *Bus) OutConfig() <-chan any {
	return b.cfgChan
}
