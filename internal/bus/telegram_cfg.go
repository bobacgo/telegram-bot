package bus

// 操作类型
const (
	OpAdd    int = 1
	OpUpdate int = 2
	OpDelete int = 3
)

// 配置方式
const (
	CfgBot     int = 1
	CfgChannel int = 2
)

type ConfigEvent struct {
	OpType  int
	CfgType int
	ChatId  int64
}

func (b *Bus) InConfig() chan<- any {
	return b.cfgChan
}

func (b *Bus) OutConfig() <-chan any {
	return b.cfgChan
}
