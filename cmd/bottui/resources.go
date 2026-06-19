package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"bot/internal/dto"
	"bot/internal/repo"

	"charm.land/bubbles/v2/table"
)

type operation string

const (
	operationCreate operation = "create"
	operationUpdate operation = "update"
	operationDelete operation = "delete"
)

type fieldSpec struct {
	Key         string
	Label       string
	Value       string
	Placeholder string
	Password    bool
	Required    bool
}

type record struct {
	ID     string
	Label  string
	Cells  []string
	Values map[string]string
}

type resourceSpec struct {
	Key              string
	Title            string
	Columns          []table.Column
	CanCreate        bool
	CanUpdate        bool
	CanDelete        bool
	DeleteNeedsInput bool
	List             func(context.Context, *apiClient) ([]record, error)
	Fields           func(operation, record) []fieldSpec
	Submit           func(context.Context, *apiClient, operation, map[string]string, record) (string, error)
}

func resources() []resourceSpec {
	return []resourceSpec{
		botResource(),
		channelResource(),
		groupResource(),
		topicResource(),
		authResource(),
		operateLogResource(),
	}
}

func botResource() resourceSpec {
	return resourceSpec{
		Key:       "bot",
		Title:     "Bot 管理",
		CanCreate: true,
		CanUpdate: true,
		CanDelete: true,
		Columns: []table.Column{
			{Title: "ID", Width: 6},
			{Title: "TG_ID", Width: 14},
			{Title: "USERNAME", Width: 18},
			{Title: "OWNER", Width: 14},
			{Title: "TYPE", Width: 6},
			{Title: "STATUS", Width: 9},
			{Title: "HEALTH_GROUP", Width: 16},
			{Title: "UPDATED_AT", Width: 19},
		},
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListBot(ctx, botFilter{})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				values := map[string]string{
					"id":              strconv.Itoa(row.Id),
					"bot_tg_id":       strconv.FormatInt(row.BotTgId, 10),
					"username":        row.Username,
					"token":           row.Token,
					"webhook_secret":  row.WebhookSecret,
					"owner":           row.Owner,
					"type":            strconv.Itoa(row.Type),
					"health_group_id": strconv.FormatInt(row.HealthGroupId, 10),
					"status":          strconv.Itoa(row.Status),
				}
				result = append(result, record{
					ID:     values["id"],
					Label:  fmt.Sprintf("bot #%d %s", row.Id, row.Username),
					Values: values,
					Cells: []string{
						values["id"],
						values["bot_tg_id"],
						row.Username,
						row.Owner,
						values["type"],
						statusText(row.Status),
						values["health_group_id"],
						formatUnix(row.UpdatedAt),
					},
				})
			}
			return result, nil
		},
		Fields: func(op operation, rec record) []fieldSpec {
			if op == operationCreate {
				return []fieldSpec{
					textField("token", "token", "", true),
					textField("webhook_secret", "webhook secret", "", false),
					textField("owner", "owner", "", false),
					textField("type", "type", "1", false),
					textField("health_group_id", "health group id", "", true),
					textField("status", "status", "1", false),
					textField("proxy_url", "proxy url", "", false),
				}
			}
			return []fieldSpec{
				textField("id", "id", rec.Values["id"], true),
				textField("username", "username", rec.Values["username"], false),
				textField("token", "token", rec.Values["token"], false),
				textField("webhook_secret", "webhook secret", rec.Values["webhook_secret"], false),
				textField("owner", "owner", rec.Values["owner"], false),
				textField("type", "type", rec.Values["type"], false),
				textField("health_group_id", "health group id", rec.Values["health_group_id"], false),
				textField("status", "status", rec.Values["status"], false),
			}
		},
		Submit: func(ctx context.Context, client *apiClient, op operation, values map[string]string, rec record) (string, error) {
			switch op {
			case operationCreate:
				req := &dto.BotCreateReq{
					Token:         values["token"],
					WebhookSecret: values["webhook_secret"],
					Owner:         values["owner"],
					Type:          atoi(values["type"]),
					HealthGroupId: atoi64(values["health_group_id"]),
					Status:        atoi(values["status"]),
				}
				if err := client.CreateBot(ctx, req, values["proxy_url"]); err != nil {
					return "", err
				}
				return "创建成功: bot", nil
			case operationUpdate:
				secret := values["webhook_secret"]
				req := &repo.BotUpdateReq{
					Id:            atoi(values["id"]),
					Username:      values["username"],
					Token:         values["token"],
					WebhookSecret: &secret,
					Owner:         values["owner"],
					Type:          atoi(values["type"]),
					HealthGroupId: atoi64(values["health_group_id"]),
					Status:        atoi(values["status"]),
				}
				if err := client.UpdateBot(ctx, req); err != nil {
					return "", err
				}
				return "更新成功: bot id=" + values["id"], nil
			case operationDelete:
				if err := client.DeleteBot(ctx, atoi(rec.ID)); err != nil {
					return "", err
				}
				return "删除成功: " + rec.Label, nil
			}
			return "", nil
		},
	}
}

func channelResource() resourceSpec {
	return resourceSpec{
		Key:       "channel",
		Title:     "频道管理",
		CanCreate: true,
		CanUpdate: true,
		CanDelete: true,
		Columns:   commonColumns("TG_CHANNEL_ID"),
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListChannel(ctx, commonFilter{})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				result = append(result, commonRecord(row.Id, row.TgChannelId, row.Title, row.Username, row.Owner, row.Type, row.Status, row.UpdatedAt))
			}
			return result, nil
		},
		Fields: func(op operation, rec record) []fieldSpec {
			return commonFields(op, rec, "tg_channel_id")
		},
		Submit: func(ctx context.Context, client *apiClient, op operation, values map[string]string, rec record) (string, error) {
			switch op {
			case operationCreate:
				row, err := client.CreateChannel(ctx, &dto.ChannelCreateReq{
					TgChannelId: atoi64(values["tg_channel_id"]),
					Title:       values["title"],
					Username:    values["username"],
					Owner:       values["owner"],
					Type:        atoi(values["type"]),
					Status:      atoi(values["status"]),
				})
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("创建成功: channel id=%d", row.Id), nil
			case operationUpdate:
				username := values["username"]
				if err := client.UpdateChannel(ctx, &repo.ChannelUpdateReq{
					Id:       atoi(values["id"]),
					Title:    values["title"],
					Username: &username,
					Owner:    values["owner"],
					Type:     atoi(values["type"]),
					Status:   atoi(values["status"]),
				}); err != nil {
					return "", err
				}
				return "更新成功: channel id=" + values["id"], nil
			case operationDelete:
				if err := client.DeleteChannel(ctx, atoi(rec.ID)); err != nil {
					return "", err
				}
				return "删除成功: " + rec.Label, nil
			}
			return "", nil
		},
	}
}

func groupResource() resourceSpec {
	return resourceSpec{
		Key:       "group",
		Title:     "群组管理",
		CanCreate: true,
		CanUpdate: true,
		CanDelete: true,
		Columns:   commonColumns("TG_GROUP_ID"),
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListGroup(ctx, commonFilter{})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				result = append(result, commonRecord(row.Id, row.TgGroupId, row.Title, row.Username, row.Owner, row.Type, row.Status, row.UpdatedAt))
			}
			return result, nil
		},
		Fields: func(op operation, rec record) []fieldSpec {
			return commonFields(op, rec, "tg_group_id")
		},
		Submit: func(ctx context.Context, client *apiClient, op operation, values map[string]string, rec record) (string, error) {
			switch op {
			case operationCreate:
				row, err := client.CreateGroup(ctx, &dto.GroupCreateReq{
					TgGroupId: atoi64(values["tg_group_id"]),
					Title:     values["title"],
					Username:  values["username"],
					Owner:     values["owner"],
					Type:      atoi(values["type"]),
					Status:    atoi(values["status"]),
				})
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("创建成功: group id=%d", row.Id), nil
			case operationUpdate:
				username := values["username"]
				if err := client.UpdateGroup(ctx, &repo.GroupUpdateReq{
					Id:        atoi(values["id"]),
					TgGroupId: atoi64(values["tg_group_id"]),
					Title:     values["title"],
					Username:  &username,
					Owner:     values["owner"],
					Type:      atoi(values["type"]),
					Status:    atoi(values["status"]),
				}); err != nil {
					return "", err
				}
				return "更新成功: group id=" + values["id"], nil
			case operationDelete:
				if err := client.DeleteGroup(ctx, atoi(rec.ID)); err != nil {
					return "", err
				}
				return "删除成功: " + rec.Label, nil
			}
			return "", nil
		},
	}
}

func topicResource() resourceSpec {
	return resourceSpec{
		Key:              "topic",
		Title:            "群话题管理",
		CanCreate:        true,
		CanUpdate:        true,
		CanDelete:        true,
		DeleteNeedsInput: true,
		Columns: []table.Column{
			{Title: "TG_GROUP_ID", Width: 16},
			{Title: "TOPIC_ID", Width: 12},
			{Title: "NAME", Width: 22},
			{Title: "UPDATED_AT", Width: 19},
		},
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListTopic(ctx, topicFilter{})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				values := map[string]string{
					"tg_group_id": strconv.FormatInt(row.TgGroupId, 10),
					"topic_id":    strconv.FormatInt(row.TopicId, 10),
					"name":        row.Name,
				}
				result = append(result, record{
					Label:  "topic " + values["tg_group_id"] + ":" + values["topic_id"],
					Values: values,
					Cells: []string{
						values["tg_group_id"],
						values["topic_id"],
						row.Name,
						formatUnix(row.UpdatedAt),
					},
				})
			}
			return result, nil
		},
		Fields: func(op operation, rec record) []fieldSpec {
			if op == operationCreate {
				return []fieldSpec{
					textField("tg_group_id", "tg group id", rec.Values["tg_group_id"], true),
					textField("topic_id", "topic id", rec.Values["topic_id"], true),
					textField("name", "name", rec.Values["name"], true),
				}
			}
			if op == operationDelete {
				return []fieldSpec{textField("id", "topic row id", "", true)}
			}
			return []fieldSpec{
				textField("id", "topic row id", "", true),
				textField("tg_group_id", "tg group id", rec.Values["tg_group_id"], false),
				textField("topic_id", "topic id", rec.Values["topic_id"], false),
				textField("name", "name", rec.Values["name"], false),
			}
		},
		Submit: func(ctx context.Context, client *apiClient, op operation, values map[string]string, rec record) (string, error) {
			switch op {
			case operationCreate:
				row, err := client.CreateTopic(ctx, &dto.TopicCreateReq{
					TgGroupId: atoi64(values["tg_group_id"]),
					TopicId:   atoi64(values["topic_id"]),
					Name:      values["name"],
				})
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("创建成功: topic %d:%d", row.TgGroupId, row.TopicId), nil
			case operationUpdate:
				if err := client.UpdateTopic(ctx, &repo.TopicUpdateReq{
					Id:        atoi(values["id"]),
					TgGroupId: atoi64(values["tg_group_id"]),
					TopicId:   atoi64(values["topic_id"]),
					Name:      values["name"],
				}); err != nil {
					return "", err
				}
				return "更新成功: topic id=" + values["id"], nil
			case operationDelete:
				if err := client.DeleteTopic(ctx, atoi(values["id"])); err != nil {
					return "", err
				}
				return "删除成功: topic id=" + values["id"], nil
			}
			return "", nil
		},
	}
}

func authResource() resourceSpec {
	return resourceSpec{
		Key:       "auth",
		Title:     "Token 管理",
		CanCreate: true,
		CanUpdate: true,
		CanDelete: true,
		Columns: []table.Column{
			{Title: "USERNAME", Width: 16},
			{Title: "TOKEN", Width: 48},
			{Title: "STATUS", Width: 10},
			{Title: "CREATED_AT", Width: 19},
		},
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListAuth(ctx, authFilter{})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				values := map[string]string{
					"username": row.Username,
					"token":    row.Token,
					"status":   strconv.Itoa(row.Status),
				}
				result = append(result, record{
					ID:     row.Username,
					Label:  "auth " + row.Username,
					Values: values,
					Cells:  []string{row.Username, row.Token, authStatusText(row.Status), formatUnix(row.CreatedAt)},
				})
			}
			return result, nil
		},
		Fields: func(op operation, rec record) []fieldSpec {
			return []fieldSpec{
				textField("username", "username", rec.Values["username"], true),
				textField("token", "token", rec.Values["token"], op == operationCreate),
				textField("status", "status", valueOr(rec.Values["status"], "1"), false),
			}
		},
		Submit: func(ctx context.Context, client *apiClient, op operation, values map[string]string, rec record) (string, error) {
			switch op {
			case operationCreate:
				row, err := client.CreateAuth(ctx, &repo.Auth{
					Username: values["username"],
					Token:    values["token"],
					Status:   atoi(values["status"]),
				})
				if err != nil {
					return "", err
				}
				return "创建成功: auth " + row.Username, nil
			case operationUpdate:
				row, err := client.UpdateAuth(ctx, &repo.AuthUpdateReq{
					Username: values["username"],
					Token:    values["token"],
					Status:   atoi(values["status"]),
				})
				if err != nil {
					return "", err
				}
				return "更新成功: auth " + row.Username, nil
			case operationDelete:
				if err := client.DeleteAuth(ctx, rec.ID); err != nil {
					return "", err
				}
				return "删除成功: " + rec.Label, nil
			}
			return "", nil
		},
	}
}

func operateLogResource() resourceSpec {
	return resourceSpec{
		Key:   "operate_log",
		Title: "操作日志",
		Columns: []table.Column{
			{Title: "ID", Width: 8},
			{Title: "OPERATOR", Width: 12},
			{Title: "MODULE", Width: 12},
			{Title: "TARGET", Width: 16},
			{Title: "TYPE", Width: 8},
			{Title: "IP", Width: 15},
			{Title: "OPERATE_AT", Width: 19},
			{Title: "CONTENT", Width: 48},
		},
		List: func(ctx context.Context, client *apiClient) ([]record, error) {
			rows, err := client.ListOperateLog(ctx, operateLogFilter{Page: 1, PageSize: 100})
			if err != nil {
				return nil, err
			}
			result := make([]record, 0, len(rows))
			for _, row := range rows {
				id := strconv.FormatInt(row.Id, 10)
				result = append(result, record{
					ID:    id,
					Label: "operate log #" + id,
					Cells: []string{
						id,
						row.Operator,
						row.ModuleName,
						row.TargetId,
						opTypeText(row.OpType),
						row.IpAddress,
						formatUnix(row.OperateAt),
						truncate(row.Content, 80),
					},
				})
			}
			return result, nil
		},
	}
}

func commonColumns(tgIDTitle string) []table.Column {
	return []table.Column{
		{Title: "ID", Width: 6},
		{Title: tgIDTitle, Width: 16},
		{Title: "TITLE", Width: 22},
		{Title: "USERNAME", Width: 18},
		{Title: "OWNER", Width: 14},
		{Title: "TYPE", Width: 6},
		{Title: "STATUS", Width: 9},
		{Title: "UPDATED_AT", Width: 19},
	}
}

func commonFields(op operation, rec record, tgKey string) []fieldSpec {
	fields := []fieldSpec{}
	if op == operationUpdate {
		fields = append(fields, textField("id", "id", rec.Values["id"], true))
	}
	fields = append(fields,
		textField(tgKey, strings.ReplaceAll(tgKey, "_", " "), rec.Values[tgKey], op == operationCreate),
		textField("title", "title", rec.Values["title"], op == operationCreate),
		textField("username", "username", rec.Values["username"], false),
		textField("owner", "owner", rec.Values["owner"], false),
		textField("type", "type", valueOr(rec.Values["type"], "1"), false),
		textField("status", "status", valueOr(rec.Values["status"], "1"), false),
	)
	return fields
}

func commonRecord(id int, tgID int64, title, username, owner string, rowType, status int, updatedAt int64) record {
	values := map[string]string{
		"id":            strconv.Itoa(id),
		"tg_channel_id": strconv.FormatInt(tgID, 10),
		"tg_group_id":   strconv.FormatInt(tgID, 10),
		"title":         title,
		"username":      username,
		"owner":         owner,
		"type":          strconv.Itoa(rowType),
		"status":        strconv.Itoa(status),
	}
	return record{
		ID:     values["id"],
		Label:  fmt.Sprintf("#%d %s", id, title),
		Values: values,
		Cells: []string{
			values["id"],
			strconv.FormatInt(tgID, 10),
			title,
			username,
			owner,
			values["type"],
			statusText(status),
			formatUnix(updatedAt),
		},
	}
}

func textField(key, label, value string, required bool) fieldSpec {
	return fieldSpec{Key: key, Label: label, Value: value, Required: required}
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func atoi64(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func statusText(status int) string {
	switch status {
	case 1:
		return "enabled"
	case 2:
		return "disabled"
	case 3:
		return "blocked"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func authStatusText(status int) string {
	switch status {
	case repo.AuthStatusUsable:
		return "enabled"
	case repo.AuthStatusDisabled:
		return "disabled"
	default:
		return fmt.Sprintf("unknown(%d)", status)
	}
}

func opTypeText(opType int) string {
	switch opType {
	case repo.OpAdd:
		return "create"
	case repo.OpUpdate:
		return "update"
	case repo.OpDelete:
		return "delete"
	default:
		return fmt.Sprintf("unknown(%d)", opType)
	}
}

func formatUnix(ts int64) string {
	if ts <= 0 {
		return "-"
	}
	return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
