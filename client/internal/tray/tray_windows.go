//go:build windows

// Package tray 提供系统托盘（Windows 纯 Go 实现，getlantern/systray）。
// 其他平台由 tray_other.go 提供空实现，避免 Linux CI 拉入 appindicator CGo 依赖。
package tray

import (
	"log"
	"sync"

	"github.com/getlantern/systray"
)

// Callbacks 托盘菜单回调（由 app 层注入）。
type Callbacks struct {
	OnReady      func()      // 图标与菜单就绪（app 置 trayReady，关闭链路兜底用）
	OnShow       func()      // 显示主界面
	OnToggleMute func() bool // 切换静音，返回切换后状态
	OnQuit       func()      // 真正退出应用
	IsMuted      func() bool // 查询当前静音态（构造勾选用）
}

var (
	cbMu sync.Mutex
	cbs  Callbacks

	menuMu sync.Mutex
	mShow  *systray.MenuItem
	mMute  *systray.MenuItem
	mQuit  *systray.MenuItem
)

// Run 启动托盘。Windows 下 systray.Run 需独占一个 goroutine 的消息循环，
// 由调用方以 go tray.Run(...) 启动；阻塞直到 Quit。
func Run(cb Callbacks) {
	cbMu.Lock()
	cbs = cb
	cbMu.Unlock()
	systray.Run(onReady, onExit)
}

// SetMuted 更新静音勾选状态（跨线程安全）。
func SetMuted(muted bool) {
	menuMu.Lock()
	defer menuMu.Unlock()
	if mMute == nil {
		return
	}
	if muted {
		mMute.Check()
	} else {
		mMute.Uncheck()
	}
}

// Quit 移除托盘图标并结束 systray 循环。
func Quit() {
	systray.Quit()
}

func onReady() {
	systray.SetIcon(iconICO())
	systray.SetTooltip("轻音 LiteTalk")

	menuMu.Lock()
	mShow = systray.AddMenuItem("显示主界面", "显示主界面")
	mMute = systray.AddMenuItemCheckbox("静音", "静音", false)
	mQuit = systray.AddMenuItem("退出", "退出")
	menuMu.Unlock()

	cbMu.Lock()
	cb := cbs
	cbMu.Unlock()
	if cb.IsMuted != nil && cb.IsMuted() {
		SetMuted(true)
	}

	// 图标+菜单都建好后才宣告托盘就绪（问题 4a/4c：就绪前关窗=真退出）。
	if cb.OnReady != nil {
		cb.OnReady()
	}

	for {
		select {
		case <-mShow.ClickedCh:
			if cb.OnShow != nil {
				cb.OnShow()
			}
		case <-mMute.ClickedCh:
			if cb.OnToggleMute != nil {
				SetMuted(cb.OnToggleMute())
			}
		case <-mQuit.ClickedCh:
			if cb.OnQuit != nil {
				cb.OnQuit()
			}
			systray.Quit()
			return
		}
	}
}

func onExit() {
	log.Println("tray: 已退出")
}
