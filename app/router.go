package app

import "net/http"

func (api *API) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/bot/create", api.BotCreate)
	mux.HandleFunc("PUT /api/bot/update", api.BotUpdate)
	mux.HandleFunc("DELETE /api/bot/delete", api.BotDelete)
	mux.HandleFunc("GET /api/bot/list", api.BotList)
	return mux
}
