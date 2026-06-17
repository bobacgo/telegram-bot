package api

import (
	"bot/bus"
	"bot/repo"
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

	BotAPI     *BotAPI
	ChannelAPI *ChannelAPI
	GroupAPI   *GroupAPI
	TopicAPI   *TopicAPI
}

func NewAPI(cfg *Config, bus *bus.Bus, repo *repo.Repo) *API {
	if cfg == nil {
		cfg = &Config{}
	}
	a := &API{
		cfg:        cfg,
		BotAPI:     NewBotAPI(bus, repo),
		ChannelAPI: NewChannelAPI(bus, repo),
		GroupAPI:   NewGroupAPI(repo),
		TopicAPI:   NewTopicAPI(repo),
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
	mux.HandleFunc("POST /api/bot/create", api.BotAPI.Create)
	mux.HandleFunc("PUT /api/bot/update", api.BotAPI.Update)
	mux.HandleFunc("DELETE /api/bot/delete", api.BotAPI.Delete)
	mux.HandleFunc("GET /api/bot/list", api.BotAPI.List)

	mux.HandleFunc("POST /api/channel/create", api.ChannelAPI.Create)
	mux.HandleFunc("PUT /api/channel/update", api.ChannelAPI.Update)
	mux.HandleFunc("DELETE /api/channel/delete", api.ChannelAPI.Delete)
	mux.HandleFunc("GET /api/channel/list", api.ChannelAPI.List)

	mux.HandleFunc("POST /api/group/create", api.GroupAPI.Create)
	mux.HandleFunc("PUT /api/group/update", api.GroupAPI.Update)
	mux.HandleFunc("DELETE /api/group/delete", api.GroupAPI.Delete)
	mux.HandleFunc("GET /api/group/list", api.GroupAPI.List)

	mux.HandleFunc("POST /api/topic/create", api.TopicAPI.Create)
	mux.HandleFunc("PUT /api/topic/update", api.TopicAPI.Update)
	mux.HandleFunc("DELETE /api/topic/delete", api.TopicAPI.Delete)
	mux.HandleFunc("GET /api/topic/list", api.TopicAPI.List)
	return mux
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
