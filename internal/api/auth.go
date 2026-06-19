package api

import (
	"bot/internal/repo"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 普通鉴权
func requireAuth(cache *authCache, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Token")
		if token == "" {
			writeErr(w, http.StatusUnauthorized, "missing auth token")
			return
		}
		tokenInfo, ok := cache.validateToken(token)
		if !ok {
			slog.Error("invalid auth token")
			writeErr(w, http.StatusUnauthorized, "invalid auth token")
			return
		}

		r.Header.Set("username", tokenInfo.Username)
		next(w, r)
	}
}

// 管理员鉴权，要求 token 对应的用户名必须是 admin
func requireAdmin(cache *authCache, next http.HandlerFunc) http.HandlerFunc {
	return requireAuth(cache, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("username") != "admin" {
			writeErr(w, http.StatusForbidden, "admin token required")
			return
		}
		next(w, r)
	})
}

type authCache struct {
	repo    *repo.AuthRepo
	mu      sync.RWMutex
	byToken map[string]*repo.Auth
}

func newAuthCache(authRepo *repo.AuthRepo) *authCache {
	cache := &authCache{
		repo:    authRepo,
		byToken: map[string]*repo.Auth{},
	}
	cache.load()
	return cache
}

func (c *authCache) load() {
	rows, err := c.repo.List(context.Background(), &repo.AuthQuery{})
	if err != nil {
		slog.Error("load auth cache failed", "err", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, row := range rows {
		copied := *row
		c.byToken[row.Token] = &copied
	}
}

func (c *authCache) validateToken(token string) (*repo.Auth, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	row := c.byToken[token]
	return row, row != nil && row.Status == repo.AuthStatusUsable
}

func (c *authCache) set(row *repo.Auth) {
	c.mu.Lock()
	defer c.mu.Unlock()
	copied := *row
	c.byToken[row.Token] = &copied
}

func (c *authCache) delete(username string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, row := range c.byToken {
		if row.Username == username {
			delete(c.byToken, row.Token)
			break
		}
	}
}

type AuthAPI struct {
	repo  *repo.AuthRepo
	cache *authCache
}

func NewAuthAPI(cache *authCache) *AuthAPI {
	return &AuthAPI{
		repo:  cache.repo,
		cache: cache,
	}
}

func (api *AuthAPI) Create(w http.ResponseWriter, r *http.Request) {
	var req repo.Auth
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Token = strings.TrimSpace(req.Token)
	if req.Username == "" || req.Token == "" {
		writeErr(w, http.StatusBadRequest, "username and token are required")
		return
	}
	if req.Status == 0 {
		req.Status = repo.AuthStatusUsable
	}
	if req.CreatedAt == 0 {
		req.CreatedAt = time.Now().Unix()
	}

	if err := api.repo.Insert(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.cache.set(&req)
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: req})
}

func (api *AuthAPI) Update(w http.ResponseWriter, r *http.Request) {
	var req repo.AuthUpdateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Token = strings.TrimSpace(req.Token)
	if req.Username == "" {
		writeErr(w, http.StatusBadRequest, "username is required")
		return
	}

	if err := api.repo.Update(r.Context(), &req); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	row, err := api.repo.FindOne(r.Context(), repo.AuthFindOneReq{Username: req.Username})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.cache.set(row)
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: row})
}

func (api *AuthAPI) Delete(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("username"))
	if username == "" {
		writeErr(w, http.StatusBadRequest, "username is required")
		return
	}
	if err := api.repo.Delete(r.Context(), username); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	api.cache.delete(username)
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: map[string]any{"deleted": username}})
}

func (api *AuthAPI) List(w http.ResponseWriter, r *http.Request) {
	urlValues := r.URL.Query()
	status, _ := strconv.Atoi(urlValues.Get("status"))
	filter := &repo.AuthQuery{
		Username: strings.TrimSpace(urlValues.Get("username")),
		Token:    strings.TrimSpace(urlValues.Get("token")),
		Status:   status,
	}

	rows, err := api.repo.List(r.Context(), filter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ApiResp{Code: 0, Msg: "ok", Data: rows})
}
