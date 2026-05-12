package repo

type Repo struct {
	db         *DB
	Bot        *BotRepo
	Channel    *ChannelRepo
	Group      *GroupRepo
	GroupTopic *GroupTopicRepo
}

func NewRepo(cfg *DBConfig) *Repo {
	// 初始化数据库连接
	db := NewDB(cfg)
	return &Repo{
		db:         db,
		Bot:        NewBotRepo(db),
		Channel:    NewChannelRepo(db),
		Group:      NewGroupRepo(db),
		GroupTopic: NewGroupTopicRepo(db),
	}
}

func (repo *Repo) Close() error {
	return repo.db.Close()
}
