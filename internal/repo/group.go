package repo

import "context"

const TelegramGroupTable = "telegram_group"

const TgGroupId string = "tg_group_id"

const (
	GroupStatusUsable = 1
	GroupStatusClosed = 2
)

type TelegramGroup struct {
	Id        int
	TgGroupId int64
	Title     string
	Username  string // optional
	Owner     string
	Type      int
	Status    int // 状态 1.启用 2.禁用
	CreatedAt int64
	UpdatedAt int64
}

func (row *TelegramGroup) Mapping() []*Mapping {
	return []*Mapping{
		{Id, &row.Id, row.Id},
		{TgGroupId, &row.TgGroupId, row.TgGroupId},
		{Title, &row.Title, row.Title},
		{Username, &row.Username, row.Username},
		{Owner, &row.Owner, row.Owner},
		{Type, &row.Type, row.Type},
		{Status, &row.Status, row.Status},
		{CreatedAt, &row.CreatedAt, row.CreatedAt},
		{UpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type GroupRepo struct {
	db *DB
}

func NewGroupRepo(db *DB) *GroupRepo {
	return &GroupRepo{
		db: db,
	}
}

func (repo *GroupRepo) Insert(ctx context.Context, row *TelegramGroup) error {
	_, err := repo.db.Insert(ctx, TelegramGroupTable, []string{Id}, row)
	return err
}

func (repo *GroupRepo) Delete(ctx context.Context, id int) error {
	_, err := repo.db.Delete(ctx, TelegramGroupTable, Wheres{{Id + " = ?", id}})
	return err
}

type GroupFindOneReq struct {
	Id        int
	TgGroupId int64
}

func (repo *GroupRepo) FindOne(ctx context.Context, req GroupFindOneReq) (*TelegramGroup, error) {
	where := make(Wheres, 0)
	if req.Id != 0 {
		where.And(Id+" = ?", req.Id)
	}
	if req.TgGroupId != 0 {
		where.And(TgGroupId+" = ?", req.TgGroupId)
	}

	query := Query[*TelegramGroup]{
		NewRow: func() *TelegramGroup { return &TelegramGroup{} },
		Where:  where,
	}
	return FindOne(ctx, repo.db, TelegramGroupTable, query)
}

type GroupUpdateReq struct {
	Id        int     `json:"id"`
	TgGroupId int64   `json:"tg_group_id"`
	Title     string  `json:"title"`
	Username  *string `json:"username"`
	Owner     string  `json:"owner"`
	Type      int     `json:"type"`
	Status    int     `json:"status"`
}

func (repo *GroupRepo) Update(ctx context.Context, row *GroupUpdateReq) error {
	m := map[string]any{}
	if row.TgGroupId != 0 {
		m[TgGroupId] = row.TgGroupId
	}
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

	_, err := repo.db.Update(ctx, TelegramGroupTable, Wheres{{Id + " = ?", row.Id}}, m)
	return err
}

type TelegramGroupQuery struct {
	Owner  string
	Type   int
	Status int
}

func (repo *GroupRepo) List(ctx context.Context, filter *TelegramGroupQuery) ([]*TelegramGroup, error) {
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

	query := Query[*TelegramGroup]{
		NewRow:  func() *TelegramGroup { return &TelegramGroup{} },
		Where:   where,
		OrderBy: CreatedAt + " DESC",
	}

	rows, err := Find(ctx, repo.db, TelegramGroupTable, query)
	return rows, err
}
