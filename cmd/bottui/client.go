package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bot/internal/dto"
	"bot/internal/repo"
)

type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
}

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type botFilter struct {
	Owner  string
	Type   int
	Status string
}

type commonFilter struct {
	Owner  string
	Type   int
	Status int
}

type topicFilter struct {
	TgGroupId int64
	TopicId   int64
	Name      string
}

type authFilter struct {
	Username string
	Token    string
	Status   int
}

type operateLogFilter struct {
	ModuleName string
	TargetId   string
	Page       int
	PageSize   int
}

func newAPIClient(addr string, token string) *apiClient {
	return &apiClient{
		baseURL: normalizeBaseURL(addr),
		token:   strings.TrimSpace(token),
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *apiClient) CreateBot(ctx context.Context, req *dto.BotCreateReq, proxyURL string) error {
	headers := map[string]string{}
	if strings.TrimSpace(proxyURL) != "" {
		headers["proxy_url"] = strings.TrimSpace(proxyURL)
	}
	return c.do(ctx, http.MethodPost, "/api/bot/create", nil, req, nil, headers)
}

func (c *apiClient) UpdateBot(ctx context.Context, req *repo.BotUpdateReq) error {
	return c.do(ctx, http.MethodPut, "/api/bot/update", nil, req, nil, nil)
}

func (c *apiClient) DeleteBot(ctx context.Context, id int) error {
	query := url.Values{}
	query.Set("id", strconv.Itoa(id))
	return c.do(ctx, http.MethodDelete, "/api/bot/delete", query, nil, nil, nil)
}

func (c *apiClient) ListBot(ctx context.Context, filter botFilter) ([]*repo.TelegramBot, error) {
	query := url.Values{}
	if filter.Owner != "" {
		query.Set("owner", filter.Owner)
	}
	if filter.Type != 0 {
		query.Set("type", strconv.Itoa(filter.Type))
	}
	if filter.Status != "" {
		query.Set("status", filter.Status)
	}
	var rows []*repo.TelegramBot
	if err := c.do(ctx, http.MethodGet, "/api/bot/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) CreateChannel(ctx context.Context, req *dto.ChannelCreateReq) (*repo.TelegramChannel, error) {
	var row repo.TelegramChannel
	if err := c.do(ctx, http.MethodPost, "/api/channel/create", nil, req, &row, nil); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *apiClient) UpdateChannel(ctx context.Context, req *repo.ChannelUpdateReq) error {
	return c.do(ctx, http.MethodPut, "/api/channel/update", nil, req, nil, nil)
}

func (c *apiClient) DeleteChannel(ctx context.Context, id int) error {
	query := url.Values{}
	query.Set("id", strconv.Itoa(id))
	return c.do(ctx, http.MethodDelete, "/api/channel/delete", query, nil, nil, nil)
}

func (c *apiClient) ListChannel(ctx context.Context, filter commonFilter) ([]*repo.TelegramChannel, error) {
	query := commonQuery(filter)
	var rows []*repo.TelegramChannel
	if err := c.do(ctx, http.MethodGet, "/api/channel/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) CreateGroup(ctx context.Context, req *dto.GroupCreateReq) (*repo.TelegramGroup, error) {
	var row repo.TelegramGroup
	if err := c.do(ctx, http.MethodPost, "/api/group/create", nil, req, &row, nil); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *apiClient) UpdateGroup(ctx context.Context, req *repo.GroupUpdateReq) error {
	return c.do(ctx, http.MethodPut, "/api/group/update", nil, req, nil, nil)
}

func (c *apiClient) DeleteGroup(ctx context.Context, id int) error {
	query := url.Values{}
	query.Set("id", strconv.Itoa(id))
	return c.do(ctx, http.MethodDelete, "/api/group/delete", query, nil, nil, nil)
}

func (c *apiClient) ListGroup(ctx context.Context, filter commonFilter) ([]*repo.TelegramGroup, error) {
	query := commonQuery(filter)
	var rows []*repo.TelegramGroup
	if err := c.do(ctx, http.MethodGet, "/api/group/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) CreateTopic(ctx context.Context, req *dto.TopicCreateReq) (*repo.TelegramGroupTopic, error) {
	var row repo.TelegramGroupTopic
	if err := c.do(ctx, http.MethodPost, "/api/topic/create", nil, req, &row, nil); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *apiClient) UpdateTopic(ctx context.Context, req *repo.TopicUpdateReq) error {
	return c.do(ctx, http.MethodPut, "/api/topic/update", nil, req, nil, nil)
}

func (c *apiClient) DeleteTopic(ctx context.Context, id int) error {
	query := url.Values{}
	query.Set("id", strconv.Itoa(id))
	return c.do(ctx, http.MethodDelete, "/api/topic/delete", query, nil, nil, nil)
}

func (c *apiClient) ListTopic(ctx context.Context, filter topicFilter) ([]*repo.TelegramGroupTopic, error) {
	query := url.Values{}
	if filter.TgGroupId != 0 {
		query.Set("tg_group_id", strconv.FormatInt(filter.TgGroupId, 10))
	}
	if filter.TopicId != 0 {
		query.Set("topic_id", strconv.FormatInt(filter.TopicId, 10))
	}
	if filter.Name != "" {
		query.Set("name", filter.Name)
	}
	var rows []*repo.TelegramGroupTopic
	if err := c.do(ctx, http.MethodGet, "/api/topic/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) CreateAuth(ctx context.Context, req *repo.Auth) (*repo.Auth, error) {
	var row repo.Auth
	if err := c.do(ctx, http.MethodPost, "/api/auth/create", nil, req, &row, nil); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *apiClient) UpdateAuth(ctx context.Context, req *repo.AuthUpdateReq) (*repo.Auth, error) {
	var row repo.Auth
	if err := c.do(ctx, http.MethodPut, "/api/auth/update", nil, req, &row, nil); err != nil {
		return nil, err
	}
	return &row, nil
}

func (c *apiClient) DeleteAuth(ctx context.Context, username string) error {
	query := url.Values{}
	query.Set("username", username)
	return c.do(ctx, http.MethodDelete, "/api/auth/delete", query, nil, nil, nil)
}

func (c *apiClient) ListAuth(ctx context.Context, filter authFilter) ([]*repo.Auth, error) {
	query := url.Values{}
	if filter.Username != "" {
		query.Set("username", filter.Username)
	}
	if filter.Token != "" {
		query.Set("token", filter.Token)
	}
	if filter.Status != 0 {
		query.Set("status", strconv.Itoa(filter.Status))
	}
	var rows []*repo.Auth
	if err := c.do(ctx, http.MethodGet, "/api/auth/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) ListOperateLog(ctx context.Context, filter operateLogFilter) ([]*repo.OperateLog, error) {
	query := url.Values{}
	if filter.ModuleName != "" {
		query.Set("module_name", filter.ModuleName)
	}
	if filter.TargetId != "" {
		query.Set("target_id", filter.TargetId)
	}
	if filter.Page > 0 {
		query.Set("page", strconv.Itoa(filter.Page))
	}
	if filter.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(filter.PageSize))
	}
	var rows []*repo.OperateLog
	if err := c.do(ctx, http.MethodGet, "/api/operate_log/list", query, nil, &rows, nil); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) do(ctx context.Context, method string, apiPath string, query url.Values, body any, out any, headers map[string]string) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	endpoint, err := c.endpoint(apiPath, query)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Token", c.token)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var envelope apiEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest || envelope.Code != 0 {
		if envelope.Msg == "" {
			envelope.Msg = resp.Status
		}
		return errors.New(envelope.Msg)
	}
	if out == nil || len(envelope.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(envelope.Data, out); err != nil {
		return fmt.Errorf("decode response data: %w", err)
	}
	return nil
}

func (c *apiClient) endpoint(apiPath string, query url.Values) (string, error) {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	rel, err := url.Parse(apiPath)
	if err != nil {
		return "", err
	}
	endpoint := base.ResolveReference(rel)
	if len(query) > 0 {
		endpoint.RawQuery = query.Encode()
	}
	return endpoint.String(), nil
}

func commonQuery(filter commonFilter) url.Values {
	query := url.Values{}
	if filter.Owner != "" {
		query.Set("owner", filter.Owner)
	}
	if filter.Type != 0 {
		query.Set("type", strconv.Itoa(filter.Type))
	}
	if filter.Status != 0 {
		query.Set("status", strconv.Itoa(filter.Status))
	}
	return query
}
