package main

import (
	"path/filepath"
	"testing"

	"github.com/OpenALC/LiteTalk/client/internal/config"
)

// 验证「最小化时休眠界面」设置项的读写与持久化
func TestHibernateOnMinimize(t *testing.T) {
	m := config.NewWithPath(filepath.Join(t.TempDir(), "config.json"))
	app := newAppWithConfig(m)

	if app.GetHibernateOnMinimize() {
		t.Fatal("默认值应为 false（默认关闭）")
	}

	app.SetHibernateOnMinimize(true)
	if !app.GetHibernateOnMinimize() {
		t.Fatal("设置 true 后读取应为 true")
	}
	// 持久化：新实例读盘恢复。
	app2 := newAppWithConfig(config.NewWithPath(m.Path()))
	if h := app2.cfg.Get().HibernateOnMinimize; h == nil || !*h {
		t.Fatal("休眠开关应已持久化到 config.json")
	}

	app.SetHibernateOnMinimize(false)
	if app.GetHibernateOnMinimize() {
		t.Fatal("设置 false 后读取应为 false")
	}
}

// 验证版本号默认值
func TestGetVersion(t *testing.T) {
	app := NewApp()
	if app.GetVersion() == "" {
		t.Fatal("版本号不应为空")
	}
}

// 验证服务器地址配置：保存 / WS 推导 / 来源
func TestServerConfig(t *testing.T) {
	m := config.NewWithPath(filepath.Join(t.TempDir(), "config.json"))
	app := newAppWithConfig(m)

	// 默认值。
	v := app.GetServerConfig()
	if v.BaseURL != config.DefaultBaseURL || v.Source != "default" {
		t.Fatalf("默认解析错误: %+v", v)
	}

	// 保存后生效（file 来源），WS 自动推导 wss。
	if err := app.SetServerConfig("https://talk.example.com"); err != nil {
		t.Fatalf("SetServerConfig: %v", err)
	}
	v = app.GetServerConfig()
	if v.BaseURL != "https://talk.example.com" || v.WSURL != "wss://talk.example.com/ws" {
		t.Fatalf("保存后解析错误: %+v", v)
	}
	if v.Source != "file" {
		t.Fatalf("来源应为 file，得到 %s", v.Source)
	}

	// 非法地址拒绝保存。
	if err := app.SetServerConfig("ftp://x"); err == nil {
		t.Fatal("非法 scheme 应报错")
	}
	if err := app.SetServerConfig("nope"); err == nil {
		t.Fatal("非法地址应报错")
	}
}

// 验证连通性测试对非法地址的快速失败。
func TestServerConnectionBadURL(t *testing.T) {
	app := newAppWithConfig(config.NewWithPath(filepath.Join(t.TempDir(), "config.json")))
	r := app.TestServerConnection("not-a-url")
	if r.OK || r.ErrorMsg == "" {
		t.Fatalf("非法地址应返回失败结果: %+v", r)
	}
}
