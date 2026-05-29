package repo

import (
	"context"
	"strconv"
	"strings"
)

const TelegramBotTable = "bot"

const (
	Id            string = "id"
	BotTgId       string = "bot_tg_id"
	Username      string = "username"
	Token         string = "token"
	WebhookSecret string = "webhook_secret"
	Owner         string = "owner"
	Type          string = "type"
	HealthGroupId string = "health_group_id"
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
	HealthGroupId int64  // 心跳检测群 ID ｜ 一个群最多支持 20 个 BOT
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
		{HealthGroupId, &row.HealthGroupId, row.HealthGroupId},
		{Status, &row.Status, row.Status},
		{CreatedAt, &row.CreatedAt, row.CreatedAt},
		{UpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type BotRepo struct {
	db *DB
}

func NewBotRepo(db *DB) *BotRepo {
	return &BotRepo{
		db: db,
	}
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
	HealthGroupId int64   `json:"health_group_id"`
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
	if row.HealthGroupId != 0 {
		m[HealthGroupId] = row.HealthGroupId
	}

	_, err := repo.db.Update(ctx, TelegramBotTable, Wheres{{Id + " = ?", row.Id}}, m)
	return err
}

func (repo *BotRepo) UpdateStatus(ctx context.Context, botTgId int64, status int32) error {
	_, err := repo.db.Update(ctx, TelegramBotTable, Wheres{{BotTgId + " = ?", botTgId}}, map[string]any{Status: status})
	return err
}

type TelegramBotQuery struct {
	Owner  string
	Type   int
	Status []int
}

func (repo *BotRepo) List(ctx context.Context, f *TelegramBotQuery) ([]*TelegramBot, error) {
	where := make(Wheres, 0)
	if f.Owner != "" {
		where.And(Owner, f.Owner)
	}
	if f.Type != 0 {
		where.And(Type, f.Type)
	}
	if len(f.Status) > 0 {
		statusStrArr := make([]string, len(f.Status))
		for i, s := range f.Status {
			statusStrArr[i] = strconv.Itoa(s)
		}
		where.And(Status+" IN ("+strings.Join(statusStrArr, ",")+")", nil)
	}

	query := Query[*TelegramBot]{
		NewRow:  func() *TelegramBot { return &TelegramBot{} },
		Where:   where,
		OrderBy: CreatedAt + " DESC",
	}

	rows, err := Find(ctx, repo.db, TelegramBotTable, query)
	return rows, err
}

// 获取开启的 bot secret 列表，用于 telegram webhook 验证
func (repo *BotRepo) FindSecretList(ctx context.Context) ([]string, error) {
	sql := "SELECT " + WebhookSecret + " FROM " + TelegramBotTable + " WHERE " + Status + " = 1"
	rows, err := repo.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	secrets := make([]string, 0)
	for rows.Next() {
		var secret string
		if err := rows.Scan(&secret); err != nil {
			return nil, err
		}
		secrets = append(secrets, secret)
	}
	return secrets, nil
}

// 获取开启 bot 的 webhook secret -> bot_tg_id 映射
func (repo *BotRepo) FindSecretBotMap(ctx context.Context) (map[string]int64, error) {
	sql := "SELECT " + WebhookSecret + ", " + BotTgId + " FROM " + TelegramBotTable + " WHERE " + Status + " = 1 AND " + WebhookSecret + " != ''"
	rows, err := repo.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int64)
	for rows.Next() {
		var (
			secret string
			botID  int64
		)
		if err := rows.Scan(&secret, &botID); err != nil {
			return nil, err
		}
		res[secret] = botID
	}
	return res, nil
}
