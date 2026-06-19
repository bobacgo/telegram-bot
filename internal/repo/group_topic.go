package repo

import "context"

const TelegramGroupTopicTable = "telegram_group_topic"

const (
	TopicId string = "topic_id"
	Name    string = "name"
)

type TelegramGroupTopic struct {
	TgGroupId int64
	TopicId   int64 // message_thread_id
	Name      string
	CreatedAt int64
	UpdatedAt int64
}

func (row *TelegramGroupTopic) Mapping() []*Mapping {
	return []*Mapping{
		{TgGroupId, &row.TgGroupId, row.TgGroupId},
		{TopicId, &row.TopicId, row.TopicId},
		{Name, &row.Name, row.Name},
		{CreatedAt, &row.CreatedAt, row.CreatedAt},
		{UpdatedAt, &row.UpdatedAt, row.UpdatedAt},
	}
}

type GroupTopicRepo struct {
	db *DB
}

func NewGroupTopicRepo(db *DB) *GroupTopicRepo {
	return &GroupTopicRepo{
		db: db,
	}
}

func (repo *GroupTopicRepo) Insert(ctx context.Context, row *TelegramGroupTopic) error {
	_, err := repo.db.Insert(ctx, TelegramGroupTopicTable, nil, row)
	return err
}

func (repo *GroupTopicRepo) Delete(ctx context.Context, id int) error {
	_, err := repo.db.Delete(ctx, TelegramGroupTopicTable, Wheres{{Id + " = ?", id}})
	return err
}

type TopicUpdateReq struct {
	Id        int    `json:"id"`
	TgGroupId int64  `json:"tg_group_id"`
	TopicId   int64  `json:"topic_id"`
	Name      string `json:"name"`
}

func (repo *GroupTopicRepo) Update(ctx context.Context, row *TopicUpdateReq) error {
	m := map[string]any{}
	if row.TgGroupId != 0 {
		m[TgGroupId] = row.TgGroupId
	}
	if row.TopicId != 0 {
		m[TopicId] = row.TopicId
	}
	if row.Name != "" {
		m[Name] = row.Name
	}

	_, err := repo.db.Update(ctx, TelegramGroupTopicTable, Wheres{{Id + " = ?", row.Id}}, m)
	return err
}

type TelegramGroupTopicQuery struct {
	TgGroupId int64
	TopicId   int64
	Name      string
}

func (repo *GroupTopicRepo) List(ctx context.Context, filter *TelegramGroupTopicQuery) ([]*TelegramGroupTopic, error) {
	where := make(Wheres, 0)
	if filter.TgGroupId != 0 {
		where.And(TgGroupId+" = ?", filter.TgGroupId)
	}
	if filter.TopicId != 0 {
		where.And(TopicId+" = ?", filter.TopicId)
	}
	if filter.Name != "" {
		where.And(Name+" = ?", filter.Name)
	}

	query := Query[*TelegramGroupTopic]{
		NewRow:  func() *TelegramGroupTopic { return &TelegramGroupTopic{} },
		Where:   where,
		OrderBy: CreatedAt + " DESC",
	}

	rows, err := Find(ctx, repo.db, TelegramGroupTopicTable, query)
	return rows, err
}
