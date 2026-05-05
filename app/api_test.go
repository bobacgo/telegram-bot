package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func assertResponse(t *testing.T, w *httptest.ResponseRecorder) {
	if w.Code != http.StatusOK {
		t.Fatalf("%s: expected status 200, got %d, body: %s", t.Name(), w.Code, w.Body.String())
	}

	var response ApiResp
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("%s: failed to unmarshal response: %v", t.Name(), err)
	}
	// 检查返回的数据格式
	if response.Code != 0 {
		t.Fatalf("%s: expected code 0, got %d, msg: %v", t.Name(), response.Code, response.Data)
	}
	byt, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("✅ %s passed, response: %v", t.Name(), string(byt))
}

func TestMain(m *testing.M) {
	if err := InitDB(DBConfig{
		DSN:     "root:@tcp(127.0.0.1:3306)/telegram?charset=utf8mb4&parseTime=True&loc=Local",
		Timeout: 5 * time.Second,
	}); err != nil {
		panic(err)
	}

	orm := NewSQL(db)
	http.DefaultServeMux = NewAPI(orm).Router()

	m.Run()
}

func TestBotList(t *testing.T) {
	// 创建 HTTP 请求
	req := httptest.NewRequest("GET", "/api/bot/list", nil)
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, req)
	// 检查响应
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	assertResponse(t, w)
}

func TestBotCreate(t *testing.T) {
	row := &BotCreateReq{
		Token:         "8441906451:AAGMpRGiyFi3HRe-06cfchlqKf8pmlS-OdA",
		WebhookSecret: "", // 不指定就默认使用 token 的前半部分
		Owner:         "test_owner",
		Type:          1,
		Status:        1,
	}
	data, _ := json.Marshal(row)

	req := httptest.NewRequest("POST", "/api/bot/create", bytes.NewReader(data))
	req.Header.Set("proxy_url", "http://127.0.0.1:7890")
	w := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}
	assertResponse(t, w)
}
