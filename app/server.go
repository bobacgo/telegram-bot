package app

import "net/http"

type HttpServeConfig struct {
	Addr string `yaml:"addr"` // API 服务器监听地址
}

func Run(cfg HttpServeConfig, db *DB) error {
	srv := NewAPI(db).Router()
	return http.ListenAndServe(cfg.Addr, srv)
}
