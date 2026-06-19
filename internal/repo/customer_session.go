package repo

import "context"

const CustomerSessionTable = "customer_session"

const (
	SessionBotTgId   string = "bot_tg_id"
	SessionUserId    string = "user_id"
	SessionUsername  string = "username"
	SessionTgGroupId string = "tg_group_id"
	SessionTopicId   string = "topic_id"
	SessionStatus    string = "status"
	SessionCreatedAt string = "created_at"
	SessionUpdatedAt string = "updated_at"
)

const (
	CustomerSessionStatusOpen   = 1
	CustomerSessionStatusClosed = 2
)

type CustomerSession struct {
	Id        int64  `json:"id"`
	BotTgId   int64  `json:"bot_tg_id"`
	UserId    int64  `json:"user_id"`
	Username  string `json:"username"`
	TgGroupId int64  `json:"tg_group_id"`
	TopicId   int    `json:"topic_id"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func (row *CustomerSession) Mapping() []*Mapping {
	return []*Mapping{
		{Id, &row.Id, row.Id},
		{SessionBotTgId, &row.BotTgId, row.BotTgId},
		{SessionUserId, &row.UserId, row.UserId},
		{SessionUsername, &row.Username, row.Username},
		{SessionTgGroupId, &row.TgGroupId, row.TgGroupId},
		{SessionTopicId, &row.TopicId, row.TopicId},
		{SessionStatus, &row.Status, row.Status},
		{SessionCreatedAt, &row.CreatedAt, row.CreatedAt},
		{SessionUpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type CustomerSessionRepo struct {
	db *DB
}

func NewCustomerSessionRepo(db *DB) *CustomerSessionRepo {
	return &CustomerSessionRepo{db: db}
}

func (repo *CustomerSessionRepo) Insert(ctx context.Context, row *CustomerSession) error {
	_, err := repo.db.Insert(ctx, CustomerSessionTable, []string{Id}, row)
	return err
}

func (repo *CustomerSessionRepo) Touch(ctx context.Context, botTgId, userId int64, updatedAt int64) error {
	_, err := repo.db.Update(ctx, CustomerSessionTable, Wheres{
		{SessionBotTgId + " = ?", botTgId},
		{"AND " + SessionUserId + " = ?", userId},
	}, map[string]any{SessionUpdatedAt: updatedAt})
	return err
}

type CustomerSessionFindOneReq struct {
	BotTgId   int64
	UserId    int64
	TgGroupId int64
	TopicId   int
	Status    int
}

func (repo *CustomerSessionRepo) FindOne(ctx context.Context, req CustomerSessionFindOneReq) (*CustomerSession, error) {
	where := repo.makeWhere(req)
	query := Query[*CustomerSession]{
		NewRow: func() *CustomerSession { return &CustomerSession{} },
		Where:  where,
	}
	return FindOne(ctx, repo.db, CustomerSessionTable, query)
}

func (repo *CustomerSessionRepo) List(ctx context.Context, req CustomerSessionFindOneReq) ([]*CustomerSession, error) {
	where := repo.makeWhere(req)
	query := Query[*CustomerSession]{
		NewRow:  func() *CustomerSession { return &CustomerSession{} },
		Where:   where,
		OrderBy: SessionUpdatedAt + " DESC",
	}
	return Find(ctx, repo.db, CustomerSessionTable, query)
}

func (repo *CustomerSessionRepo) makeWhere(req CustomerSessionFindOneReq) Wheres {
	where := make(Wheres, 0)
	if req.BotTgId != 0 {
		where.And(SessionBotTgId+" = ?", req.BotTgId)
	}
	if req.UserId != 0 {
		where.And(SessionUserId+" = ?", req.UserId)
	}
	if req.TgGroupId != 0 {
		where.And(SessionTgGroupId+" = ?", req.TgGroupId)
	}
	if req.TopicId != 0 {
		where.And(SessionTopicId+" = ?", req.TopicId)
	}
	if req.Status != 0 {
		where.And(SessionStatus+" = ?", req.Status)
	}
	return where
}
