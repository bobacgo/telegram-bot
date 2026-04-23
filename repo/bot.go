package repo

import (
	"database/sql"
)

const TelegramBotTable = "telegram_bot"

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
	Status        int    // 状态
	CreatedAt     int64
	UpdatedAt     int64
}

func (row *TelegramBot) TableName() string {
	return TelegramBotTable
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
	db *sql.DB
}

func (repo *BotRepo) Insert(row *TelegramBot) error {
	return Insert(repo.db, row)
}

func (repo *BotRepo) Delete(id int) error {
	return Delete(repo.db, TelegramBotTable, id)
}

func (repo *BotRepo) Update(row *TelegramBot) error {
	m := map[string]any{}
	if row.Username != "" {
		m[Username] = row.Username
	}
	if row.Token != "" {
		m[Token] = row.Token
	}
	if row.WebhookSecret != "" {
		m[WebhookSecret] = row.WebhookSecret
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

	return Update(repo.db, TelegramBotTable, row.Id, m)
}

type TelegramBotFilter struct {
	Owner  string
	Type   int
	Status int
}

func (repo *BotRepo) List(filter *TelegramBotFilter) ([]*TelegramBot, error) {
	where, args := "", []any{}

	if filter.Owner != "" {
		where += " AND " + Owner + " = ?"
		args = append(args, filter.Owner)
	}
	if filter.Type != 0 {
		where += " AND " + Type + " = ?"
		args = append(args, filter.Type)
	}
	if filter.Status != 0 {
		where += " AND " + Status + " = ?"
		args = append(args, filter.Status)
	}

	if where != "" {
		where = "1=1" + where
	}

	return List(repo.db, where, args, func() *TelegramBot {
		return &TelegramBot{}
	})
}
