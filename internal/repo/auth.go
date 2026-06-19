package repo

import "context"

const AuthTable = "auth"

const (
	AuthStatusUsable   = 1 // 启用
	AuthStatusDisabled = 2 // 禁用
)

const (
	AuthUsername  string = "username"
	AuthToken     string = "token"
	AuthStatus    string = "status"
	AuthCreatedAt string = "created_at"
)

type Auth struct {
	Username  string `json:"username"`
	Token     string `json:"token"`
	Status    int    `json:"status"`
	CreatedAt int64  `json:"created_at"`
}

func (row *Auth) Mapping() []*Mapping {
	return []*Mapping{
		{AuthUsername, &row.Username, row.Username},
		{AuthToken, &row.Token, row.Token},
		{AuthStatus, &row.Status, row.Status},
		{AuthCreatedAt, &row.CreatedAt, row.CreatedAt},
	}
}

type AuthRepo struct {
	db *DB
}

func NewAuthRepo(db *DB) *AuthRepo {
	return &AuthRepo{db: db}
}

func (repo *AuthRepo) Insert(ctx context.Context, row *Auth) error {
	_, err := repo.db.Insert(ctx, AuthTable, nil, row)
	return err
}

type AuthUpdateReq struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	Status   int    `json:"status"`
}

func (repo *AuthRepo) Update(ctx context.Context, req *AuthUpdateReq) error {
	m := map[string]any{}
	if req.Token != "" {
		m[AuthToken] = req.Token
	}
	if req.Status != 0 {
		m[AuthStatus] = req.Status
	}

	_, err := repo.db.Update(ctx, AuthTable, Wheres{{AuthUsername + " = ?", req.Username}}, m)
	return err
}

func (repo *AuthRepo) Delete(ctx context.Context, username string) error {
	_, err := repo.db.Delete(ctx, AuthTable, Wheres{{AuthUsername + " = ?", username}})
	return err
}

type AuthFindOneReq struct {
	Username string
	Token    string
	Status   int
}

func (repo *AuthRepo) FindOne(ctx context.Context, req AuthFindOneReq) (*Auth, error) {
	where := make(Wheres, 0)
	if req.Username != "" {
		where.And(AuthUsername+" = ?", req.Username)
	}
	if req.Token != "" {
		where.And(AuthToken+" = ?", req.Token)
	}
	if req.Status != 0 {
		where.And(AuthStatus+" = ?", req.Status)
	}

	query := Query[*Auth]{
		NewRow: func() *Auth { return &Auth{} },
		Where:  where,
	}
	return FindOne(ctx, repo.db, AuthTable, query)
}

type AuthQuery struct {
	Username string
	Token    string
	Status   int
}

func (repo *AuthRepo) List(ctx context.Context, filter *AuthQuery) ([]*Auth, error) {
	where := make(Wheres, 0)
	if filter.Username != "" {
		where.And(AuthUsername+" = ?", filter.Username)
	}
	if filter.Token != "" {
		where.And(AuthToken+" = ?", filter.Token)
	}
	if filter.Status != 0 {
		where.And(AuthStatus+" = ?", filter.Status)
	}

	query := Query[*Auth]{
		NewRow:  func() *Auth { return &Auth{} },
		Where:   where,
		OrderBy: AuthCreatedAt + " DESC",
	}
	return Find(ctx, repo.db, AuthTable, query)
}
