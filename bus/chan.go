package bus

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
