package repo

import (
	"database/sql"
	"fmt"
	"strings"
)

type Row interface {
	TableName() string
	Mapping() []*Mapping
}

type Mapping struct {
	Column string
	Result any // query result (pointer)
	Value  any // insert, update value
}

func Insert(db *sql.DB, row Row) error {
	table, mappings := row.TableName(), row.Mapping()

	columns := make([]string, 0, len(mappings))
	placeholders := make([]string, 0, len(mappings))
	values := make([]any, 0, len(mappings))

	for _, m := range mappings {
		if m.Column == "id" {
			continue // skip id column for insert
		}

		columns = append(columns, m.Column)
		placeholders = append(placeholders, "?")
		values = append(values, m.Value)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := db.Exec(query, values...)
	return err
}

func Delete(db *sql.DB, table string, id int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", table)
	_, err := db.Exec(query, id)
	return err
}

func Update(db *sql.DB, table string, id int, m map[string]any) error {
	if len(m) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClauses := make([]string, 0, len(m))
	values := make([]any, 0, len(m))
	var idValue any = id

	for column, value := range m {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		values = append(values, value)
	}

	if idValue == nil {
		return fmt.Errorf("id value is required for update")
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ?",
		table,
		strings.Join(setClauses, ", "),
	)

	values = append(values, idValue) // id value goes last for where clause
	_, err := db.Exec(query, values...)
	return err
}

func List[T Row](db *sql.DB, where string, args []any, newRow func() T) ([]T, error) {
	if where != "" {
		where = "WHERE " + where
	}

	prototype := newRow()
	mappings := prototype.Mapping()
	columns := make([]string, len(mappings))
	for i, m := range mappings {
		columns[i] = m.Column
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s %s",
		strings.Join(columns, ", "),
		prototype.TableName(),
		where,
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]T, 0)
	for rows.Next() {
		item := newRow()
		itemMappings := item.Mapping()
		scanArgs := make([]any, len(itemMappings))
		for i, m := range itemMappings {
			scanArgs[i] = m.Result
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
