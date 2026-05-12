package api

import (
	"bot/repo"
	"encoding/json"
	"log"
	"net/http"
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
	mux *http.ServeMux

	BotAPI     *BotAPI
	ChannelAPI *ChannelAPI
	GroupAPI   *GroupAPI
	TopicAPI   *TopicAPI
}

func NewAPI(cfg *Config, repo *repo.Repo) *API {
	a := &API{
		cfg:        cfg,
		BotAPI:     NewBotAPI(repo),
		ChannelAPI: NewChannelAPI(repo),
		GroupAPI:   NewGroupAPI(repo),
		TopicAPI:   NewTopicAPI(repo),
	}
	a.mux = a.Router() // 注册路由
	return a
}

func (api *API) Router() *http.ServeMux {
	mux := http.NewServeMux()
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
	if err := http.ListenAndServe(api.cfg.Addr, api.mux); err != nil {
		log.Fatalf("api listen err: %v", err)
	}
}

func (api *API) Shutdown() error {
	return nil
}
