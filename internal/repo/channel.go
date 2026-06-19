package repo

import "context"

const TelegramChannelTable = "telegram_channel"

const (
	TgChannelId string = "tg_channel_id"
	Title       string = "title"
)

const (
	ChannelStatusUsable = 1
	ChannelStatusClosed = 2
)

type TelegramChannel struct {
	Id          int
	TgChannelId int64
	Title       string
	Username    string // optional
	Owner       string
	Type        int
	Status      int // 状态 1.启用 2.禁用
	CreatedAt   int64
	UpdatedAt   int64
}

func (row *TelegramChannel) Mapping() []*Mapping {
	return []*Mapping{
		{Id, &row.Id, row.Id},
		{TgChannelId, &row.TgChannelId, row.TgChannelId},
		{Title, &row.Title, row.Title},
		{Username, &row.Username, row.Username},
		{Owner, &row.Owner, row.Owner},
		{Type, &row.Type, row.Type},
		{Status, &row.Status, row.Status},
		{CreatedAt, &row.CreatedAt, row.CreatedAt},
		{UpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type ChannelRepo struct {
	db *DB
}

func NewChannelRepo(db *DB) *ChannelRepo {
	return &ChannelRepo{
		db: db,
	}
}

func (repo *ChannelRepo) Insert(ctx context.Context, row *TelegramChannel) error {
	_, err := repo.db.Insert(ctx, TelegramChannelTable, []string{Id}, row)
	return err
}

func (repo *ChannelRepo) Delete(ctx context.Context, id int) error {
	_, err := repo.db.Delete(ctx, TelegramChannelTable, Wheres{{Id + " = ?", id}})
	return err
}

type ChannelFindOneReq struct {
	Id          int
	TgChannelId int64
}

func (repo *ChannelRepo) FindOne(ctx context.Context, req ChannelFindOneReq) (*TelegramChannel, error) {
	where := make(Wheres, 0)
	if req.Id != 0 {
		where.And(Id+" = ?", req.Id)
	}
	if req.TgChannelId != 0 {
		where.And(TgChannelId+" = ?", req.TgChannelId)
	}

	query := Query[*TelegramChannel]{
		NewRow: func() *TelegramChannel { return &TelegramChannel{} },
		Where:  where,
	}
	return FindOne(ctx, repo.db, TelegramChannelTable, query)
}

type ChannelUpdateReq struct {
	Id          int     `json:"id"`
	TgChannelId int64   `json:"tg_channel_id"`
	Title       string  `json:"title"`
	Username    *string `json:"username"`
	Owner       string  `json:"owner"`
	Type        int     `json:"type"`
	Status      int     `json:"status"`
}

func (repo *ChannelRepo) Update(ctx context.Context, row *ChannelUpdateReq) error {
	m := map[string]any{}
	if row.Title != "" {
		m[Title] = row.Title
	}
	if row.Username != nil {
		m[Username] = *row.Username
	}
	if row.Owner != "" {
		m[Owner] = row.Owner
	}
	if row.Type != 0 {
		m[Type] = row.Type
	}
	if row.Status != 0 {
		m[Status] = row.Status
	}

	_, err := repo.db.Update(ctx, TelegramChannelTable, Wheres{{Id + " = ?", row.Id}}, m)
	return err
}

type TelegramChannelQuery struct {
	Owner  string
	Type   int
	Status int
}

func (repo *ChannelRepo) List(ctx context.Context, filter *TelegramChannelQuery) ([]*TelegramChannel, error) {
	where := make(Wheres, 0)
	if filter.Owner != "" {
		where.And(Owner+" = ?", filter.Owner)
	}
	if filter.Type != 0 {
		where.And(Type+" = ?", filter.Type)
	}
	if filter.Status != 0 {
		where.And(Status+" = ?", filter.Status)
	}

	query := Query[*TelegramChannel]{
		NewRow:  func() *TelegramChannel { return &TelegramChannel{} },
		Where:   where,
		OrderBy: CreatedAt + " DESC",
	}

	rows, err := Find(ctx, repo.db, TelegramChannelTable, query)
	return rows, err
}
