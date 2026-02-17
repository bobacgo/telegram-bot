package main

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var proxyConfig ProxyConfig

func SetProxyConfig(cfg ProxyConfig) {
	proxyConfig = cfg
}

func HttpClient() *http.Client {
	clt := &http.Client{
		Timeout: 3 * time.Second,
	}
	if !proxyConfig.Enabled {
		return clt
	}

	if proxyConfig.URL == "" {
		proxyConfig.URL = "http://127.0.0.1:7890"
	}

	proxyURL, err := url.Parse(proxyConfig.URL)
	if err != nil {
		slog.Warn("invalid proxy url, using direct connection", "url", proxyConfig.URL, "err", err)
		return clt
	}
	clt.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	slog.Info("using proxy", "url", proxyConfig.URL)
	return clt
}
