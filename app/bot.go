package app

import "context"

const TelegramBotTable = "bot"

const (
	Id            string = "id"
	BotTgId       string = "bot_tg_id"
	Username      string = "username"
	Token         string = "token"
	WebhookSecret string = "webhook_secret"
	Owner         string = "owner"
	Type          string = "type"
	Status        string = "status"
	CreatedAt     string = "created_at"
	UpdatedAt     string = "updated_at"
)

type TelegramBot struct {
	Id            int
	BotTgId       int64  // bot tgid
	Username      string // bot username
	Token         string
	WebhookSecret string // telegram 通过 webhook 接口请求认证的密钥 header X-Telegram-Bot-Api-Secret-Token
	Owner         string // bot owner tg username
	Type          int    // 类型
	Status        int    // 状态 1.启用 2.禁用 3.封禁
	CreatedAt     int64
	UpdatedAt     int64
}

func (row *TelegramBot) Mapping() []*Mapping {
	return []*Mapping{
		{Id, &row.Id, row.Id},
		{BotTgId, &row.BotTgId, row.BotTgId},
		{Username, &row.Username, row.Username},
		{Token, &row.Token, row.Token},
		{WebhookSecret, &row.WebhookSecret, row.WebhookSecret},
		{Owner, &row.Owner, row.Owner},
		{Type, &row.Type, row.Type},
		{Status, &row.Status, row.Status},
		{CreatedAt, &row.CreatedAt, row.CreatedAt},
		{UpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type BotRepo struct {
	db *DB
}

func (repo *BotRepo) Insert(ctx context.Context, row *TelegramBot) error {
	_, err := repo.db.Insert(ctx, TelegramBotTable, []string{Id}, row)
	return err
}

func (repo *BotRepo) Delete(ctx context.Context, id int) error {
	_, err := repo.db.Delete(ctx, TelegramBotTable, Wheres{{Id + " = ?", id}})
	return err
}

type BotUpdateReq struct {
	Id            int     `json:"id"`
	Username      string  `json:"username"`
	Token         string  `json:"token"`
	WebhookSecret *string `json:"webhook_secret"`
	Owner         string  `json:"owner"`
	Type          int     `json:"type"`
	Status        int     `json:"status"`
}

func (repo *BotRepo) Update(ctx context.Context, row *BotUpdateReq) error {
	m := map[string]any{}
	if row.Username != "" {
		m[Username] = row.Username
	}
	if row.Token != "" {
		m[Token] = row.Token
	}
	if row.WebhookSecret != nil {
		m[WebhookSecret] = *row.WebhookSecret
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

	_, err := repo.db.Update(ctx, TelegramBotTable, Wheres{{Id + " = ?", row.Id}}, m)
	return err
}

type TelegramBotQuery struct {
	Owner  string
	Type   int
	Status int
}

func (repo *BotRepo) List(ctx context.Context, filter *TelegramBotQuery) ([]*TelegramBot, error) {
	where := make(Wheres, 0)
	if filter.Owner != "" {
		where.And(Owner, filter.Owner)
	}
	if filter.Type != 0 {
		where.And(Type, filter.Type)
	}
	if filter.Status != 0 {
		where.And(Status, filter.Status)
	}

	query := Query[*TelegramBot]{
		NewRow:  func() *TelegramBot { return &TelegramBot{} },
		Where:   where,
		OrderBy: CreatedAt + " DESC",
	}

	rows, err := Find(ctx, repo.db, TelegramBotTable, query)
	return rows, err
}
