package main

import "testing"

// 验证「最小化时休眠界面」设置项的读写
func TestHibernateOnMinimize(t *testing.T) {
	app := NewApp()

	if app.GetHibernateOnMinimize() {
		t.Fatal("默认值应为 false（默认关闭）")
	}

	app.SetHibernateOnMinimize(true)
	if !app.GetHibernateOnMinimize() {
		t.Fatal("设置 true 后读取应为 true")
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
