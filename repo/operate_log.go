package repo

import "context"

const OperateLogTable = "operate_log"

// 操作类型常量
const (
	OpAdd    = 1 // 创建
	OpUpdate = 2 // 更新
	OpDelete = 3 // 删除
)

const (
	Operator   string = "operator"
	OperateAt  string = "operate_at"
	IpAddress  string = "ip_address"
	OpType     string = "op_type"
	ModuleName string = "module_name"
	TargetId   string = "target_id"
	Content    string = "content"
	Remark     string = "remark"
)

type OperateLog struct {
	Id        int64  `json:"id"`         // 日志主键
	Operator  string `json:"operator"`   // 操作者名称
	OperateAt int64  `json:"operate_at"` // 操作时间，Unix时间戳
	IpAddress string `json:"ip_address"` // 操作者IP地址

	OpType     int    `json:"op_type"`     // 操作类型，1-创建，2-更新，3-删除
	ModuleName string `json:"module_name"` // 模块名称
	TargetId   string `json:"target_id"`   // 目标ID，例如被操作的对象的ID
	Content    string `json:"content"`     // 操作内容 json字符串，记录操作前后的数据等信息
	Remark     string `json:"remark"`      // 备注
}

func (row *OperateLog) Mapping() []*Mapping {
	return []*Mapping{
		{Id, &row.Id, row.Id},
		{Operator, &row.Operator, row.Operator},
		{OperateAt, &row.OperateAt, row.OperateAt},
		{IpAddress, &row.IpAddress, row.IpAddress},
		{OpType, &row.OpType, row.OpType},
		{ModuleName, &row.ModuleName, row.ModuleName},
		{TargetId, &row.TargetId, row.TargetId},
		{Content, &row.Content, row.Content},
		{Remark, &row.Remark, row.Remark},
	}
}

type OperateLogRepo struct {
	db *DB
}

func NewOperateLogRepo(db *DB) *OperateLogRepo {
	return &OperateLogRepo{db: db}
}

func (repo *OperateLogRepo) Insert(ctx context.Context, row *OperateLog) error {
	_, err := repo.db.Insert(ctx, OperateLogTable, []string{Id}, row)
	return err
}
