package app

import "net/http"

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
