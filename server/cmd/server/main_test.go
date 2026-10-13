package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 验证 healthz handler 返回 200 与正确的 JSON 结构
func TestHealthzHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
	rec := httptest.NewRecorder()

	healthzHandler(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("期望状态码 200，实际 %d", res.StatusCode)
	}

	if ct := res.Header.Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("期望 Content-Type application/json; charset=utf-8，实际 %q", ct)
	}

	var body struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("响应体不是合法 JSON: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("期望 status=ok，实际 %q", body.Status)
	}
	if body.Version == "" {
		t.Fatal("version 不应为空")
	}
}
