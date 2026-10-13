//go:build !windows

// Package tray 的非 Windows 空实现：
// Linux/macOS 不引入 systray（appindicator / objc CGo），保 CI 可编译。
package tray

// Callbacks 托盘菜单回调（与 Windows 实现同形）。
type Callbacks struct {
	OnReady      func()      // 就绪回调（空实现永不触发 → app 侧 trayReady 恒 false）
	OnShow       func()      // 显示主界面
	OnToggleMute func() bool // 切换静音，返回切换后状态
	OnQuit       func()      // 真正退出应用
	IsMuted      func() bool // 查询当前静音态
}

// Run 空实现。
func Run(Callbacks) {}

// SetMuted 空实现。
func SetMuted(bool) {}

// Quit 空实现。
func Quit() {}
