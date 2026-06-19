package api

import (
	"bot/internal/bus"
	"bot/internal/repo"
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Addr string `yaml:"addr"` // API 服务器监听地址
}

type ApiResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ApiResp{Code: status, Msg: msg})
}

type API struct {
	cfg *Config
	srv *http.Server

	BotAPI        *BotAPI
	ChannelAPI    *ChannelAPI
	GroupAPI      *GroupAPI
	TopicAPI      *TopicAPI
	OperateLogAPI *OperateLogAPI
	AuthAPI       *AuthAPI
	AuthCache     *authCache
}

func NewAPI(cfg *Config, bus *bus.Bus, repo *repo.Repo) *API {
	if cfg == nil {
		cfg = &Config{}
	}
	authCache := newAuthCache(repo.Auth)
	a := &API{
		cfg:           cfg,
		BotAPI:        NewBotAPI(bus, repo),
		ChannelAPI:    NewChannelAPI(bus, repo),
		GroupAPI:      NewGroupAPI(repo),
		TopicAPI:      NewTopicAPI(repo),
		OperateLogAPI: NewOperateLogAPI(repo),
		AuthAPI:       NewAuthAPI(authCache),
		AuthCache:     authCache,
	}
	a.srv = &http.Server{
		Addr:    cfg.Addr,
		Handler: a.router(),
	} // 注册路由
	return a
}

func (api *API) router() *http.ServeMux {
	mux := http.NewServeMux()

	// telegram、webhook回调

	mux.HandleFunc("POST /api/telegram/webhook", api.BotAPI.Webhook) // telegram、webhook回调

	// TODO : 统一处理API请求，鉴权、日志, panic等
	mux.HandleFunc("POST /api/bot/create", api.auth(api.BotAPI.Create))
	mux.HandleFunc("PUT /api/bot/update", api.auth(api.BotAPI.Update))
	mux.HandleFunc("DELETE /api/bot/delete", api.auth(api.BotAPI.Delete))
	mux.HandleFunc("GET /api/bot/list", api.auth(api.BotAPI.List))

	mux.HandleFunc("POST /api/channel/create", api.auth(api.ChannelAPI.Create))
	mux.HandleFunc("PUT /api/channel/update", api.auth(api.ChannelAPI.Update))
	mux.HandleFunc("DELETE /api/channel/delete", api.auth(api.ChannelAPI.Delete))
	mux.HandleFunc("GET /api/channel/list", api.auth(api.ChannelAPI.List))

	mux.HandleFunc("POST /api/group/create", api.auth(api.GroupAPI.Create))
	mux.HandleFunc("PUT /api/group/update", api.auth(api.GroupAPI.Update))
	mux.HandleFunc("DELETE /api/group/delete", api.auth(api.GroupAPI.Delete))
	mux.HandleFunc("GET /api/group/list", api.auth(api.GroupAPI.List))

	mux.HandleFunc("POST /api/topic/create", api.auth(api.TopicAPI.Create))
	mux.HandleFunc("PUT /api/topic/update", api.auth(api.TopicAPI.Update))
	mux.HandleFunc("DELETE /api/topic/delete", api.auth(api.TopicAPI.Delete))
	mux.HandleFunc("GET /api/topic/list", api.auth(api.TopicAPI.List))

	mux.HandleFunc("GET /api/operate_log/list", api.auth(api.OperateLogAPI.List))

	mux.HandleFunc("POST /api/auth/create", api.admin(api.AuthAPI.Create))
	mux.HandleFunc("PUT /api/auth/update", api.admin(api.AuthAPI.Update))
	mux.HandleFunc("DELETE /api/auth/delete", api.admin(api.AuthAPI.Delete))
	mux.HandleFunc("GET /api/auth/list", api.admin(api.AuthAPI.List))
	return mux
}

// auth 是 API 的一个方法，用于包装需要认证的处理函数。它调用 requireAuth 中间件，传入 AuthRepo 和下一个处理函数。
func (api *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return requireAuth(api.AuthCache, next)
}

func (api *API) admin(next http.HandlerFunc) http.HandlerFunc {
	return requireAdmin(api.AuthCache, next)
}

func (api *API) Run() {
	if err := api.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("api listen err: %v", err)
	}
}

func (api *API) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := api.srv.Shutdown(ctx); err != nil {
		log.Printf("api shutdown err: %v", err)
	}
	slog.Info("shutting down API server...")
}
