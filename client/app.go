package main

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// version 客户端版本号，可用 -ldflags "-X main.version=" 注入，默认 dev
var version = "dev"

// App 应用结构体：持有窗口上下文与设置项。
// 导出方法经 Wails Bind 暴露给前端调用。
type App struct {
	ctx context.Context

	// hibernateOnMinimize 设置项「最小化时休眠界面（省内存）」，默认关闭。
	hibernateOnMinimize bool
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// startup 启动回调：保存应用上下文
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// onBeforeClose 拦截窗口关闭：隐藏到托盘而非退出。
// 返回 true 阻止默认关闭流程；真正的退出走托盘菜单。
func (a *App) onBeforeClose(ctx context.Context) bool {
	runtime.WindowHide(ctx)
	return true
}

// SetHibernateOnMinimize 设置「最小化时休眠界面」开关
func (a *App) SetHibernateOnMinimize(enabled bool) {
	a.hibernateOnMinimize = enabled
	// TODO(M4)：enabled 时，最小化到托盘后前端切换到「休眠页」——
	// 卸载业务组件、停止 rAF/定时器/动画、解除 Wails 事件订阅；
	// 唤起时恢复休眠前的路由与状态（关键状态持久化 localStorage）。
}

// GetHibernateOnMinimize 读取「最小化时休眠界面」开关
func (a *App) GetHibernateOnMinimize() bool {
	return a.hibernateOnMinimize
}

// GetVersion 返回客户端版本号
func (a *App) GetVersion() string {
	return version
}
