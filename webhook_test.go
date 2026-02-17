package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestGetWebhook(t *testing.T) {
	token := "7984808294:AAGo"
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", token)

	// ✅ 设置代理（支持 socks5 或 http）
	proxyURL, _ := url.Parse("socks5://127.0.0.1:7890")
	// proxyURL, _ := url.Parse("http://127.0.0.1:8080")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// ✅ 使用代理发请求
	resp, err := client.Get(apiURL)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}

func TestDeleteWebhook(t *testing.T) {
	token := "7984808294:AAGoRl6zwCpB2_bJUb7YCIKUyoWal4DoKnI"
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook?drop_pending_updates=true", token)

	// ✅ 设置代理（支持 socks5 或 http）
	proxyURL, _ := url.Parse("socks5://127.0.0.1:7890")
	// proxyURL, _ := url.Parse("http://127.0.0.1:8080")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// ✅ 使用代理发请求
	resp, err := client.Post(apiURL, "application/json", nil)
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if ok, _ := result["ok"].(bool); ok {
		t.Log("✅ webhook 删除成功")
	} else {
		t.Errorf("❌ 删除失败: %+v", result)
	}
}
