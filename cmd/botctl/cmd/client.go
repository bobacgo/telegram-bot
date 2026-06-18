package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"bot/repo"
)

type apiClient struct {
	baseURL string
	token   string
	http    *http.Client
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

type apiEnvelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func newAPIClient(addr string, token string, timeout time.Duration) *apiClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &apiClient{
		baseURL: normalizeBaseURL(addr),
		token:   token,
		http: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *apiClient) CreateAuth(ctx context.Context, row *repo.Auth) (*repo.Auth, error) {
	var res repo.Auth
	if err := c.do(ctx, http.MethodPost, "/api/auth/create", nil, row, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *apiClient) UpdateAuth(ctx context.Context, req *repo.AuthUpdateReq) (*repo.Auth, error) {
	var res repo.Auth
	if err := c.do(ctx, http.MethodPut, "/api/auth/update", nil, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *apiClient) DeleteAuth(ctx context.Context, username string) error {
	query := url.Values{}
	query.Set("username", username)
	return c.do(ctx, http.MethodDelete, "/api/auth/delete", query, nil, nil)
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
		query.Set("status", fmt.Sprint(filter.Status))
	}

	var rows []*repo.Auth
	if err := c.do(ctx, http.MethodGet, "/api/auth/list", query, nil, &rows); err != nil {
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
		query.Set("page", fmt.Sprint(filter.Page))
	}
	if filter.PageSize > 0 {
		query.Set("page_size", fmt.Sprint(filter.PageSize))
	}

	var rows []*repo.OperateLog
	if err := c.do(ctx, http.MethodGet, "/api/operate_log/list", query, nil, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (c *apiClient) do(ctx context.Context, method string, apiPath string, query url.Values, body any, out any) error {
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

func normalizeBaseURL(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "http://127.0.0.1:8080"
	}
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return strings.TrimRight(addr, "/")
}

func missingTokenError(configPath string) error {
	return fmt.Errorf("missing token: run `botctl config set --token <token>` first, config file: %s", configPath)
}
