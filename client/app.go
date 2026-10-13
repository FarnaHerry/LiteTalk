package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/OpenALC/LiteTalk/client/internal/audio"
	"github.com/OpenALC/LiteTalk/client/internal/state"
	"github.com/gen2brain/malgo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// version 客户端版本号，可用 -ldflags "-X main.version=" 注入，默认 dev
var version = "dev"

// App 应用结构体：窗口上下文、设置项、业务会话与 API 客户端。
// 导出方法经 Wails Bind 暴露给前端调用（落地方案 §4.6）。
type App struct {
	ctx context.Context

	// hibernateOnMinimize 设置项「最小化时休眠界面（省内存）」，默认关闭。
	hibernateOnMinimize bool

	api  *state.APIClient
	sess *state.Session // 登录并连接信令后创建

	mu         sync.Mutex
	refreshTok string
	user       state.UserProfile

	audioCtx *audio.Context // 设备枚举用，惰性创建
}

// NewApp 创建应用实例
func NewApp() *App {
	return &App{}
}

// startup 启动回调：保存上下文、创建 API 客户端。
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	base := os.Getenv("LITETALK_API")
	if base == "" {
		base = "http://127.0.0.1:8077"
	}
	a.api = state.NewAPIClient(base)
}

// onBeforeClose 拦截窗口关闭：隐藏到托盘而非退出。
func (a *App) onBeforeClose(ctx context.Context) bool {
	runtime.WindowHide(ctx)
	return true
}

// ---------- 窗口控制（自定义标题栏调用） ----------

// WindowMinimize 最小化窗口。
func (a *App) WindowMinimize() { runtime.WindowMinimise(a.ctx) }

// WindowMaximize 最大化/还原。
func (a *App) WindowMaximize() { runtime.WindowToggleMaximise(a.ctx) }

// WindowClose 关闭（转隐藏到托盘）。
func (a *App) WindowClose() { runtime.WindowHide(a.ctx) }

// ---------- 设置 ----------

// SetHibernateOnMinimize 设置「最小化时休眠界面」开关
func (a *App) SetHibernateOnMinimize(enabled bool) {
	a.hibernateOnMinimize = enabled
	// TODO(M4)：enabled 时，最小化到托盘后前端切换到「休眠页」——
	// 卸载业务组件、停止 rAF/定时器/动画、解除 Wails 事件订阅；
	// 唤起时恢复休眠前的路由与状态（关键状态持久化 localStorage）。
}

// GetHibernateOnMinimize 读取「最小化时休眠界面」开关
func (a *App) GetHibernateOnMinimize() bool { return a.hibernateOnMinimize }

// GetVersion 返回客户端版本号
func (a *App) GetVersion() string { return version }

// ---------- 认证（转发 REST，§5.4） ----------

// SendEmailCode 发送验证码（purpose: register / reset）。
func (a *App) SendEmailCode(email, purpose string) error {
	return a.api.SendEmailCode(email, purpose)
}

// Login 邮箱密码登录。成功后连接信令。
func (a *App) Login(email, password string) (*state.AuthResult, error) {
	r, err := a.api.Login(email, password)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.refreshTok = r.RefreshToken
	a.user = r.User
	a.mu.Unlock()
	if err := a.connectSession(r.AccessToken, r.User.Nickname); err != nil {
		log.Printf("connect session: %v", err) // 登录成功但信令暂不可达：UI 可重试
	}
	return r, nil
}

// Register 注册并自动登录。
func (a *App) Register(nickname, email, password, code string) (*state.AuthResult, error) {
	r, err := a.api.Register(nickname, email, password, code)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	a.refreshTok = r.RefreshToken
	a.user = r.User
	a.mu.Unlock()
	_ = a.connectSession(r.AccessToken, r.User.Nickname)
	return r, nil
}

// ResetPassword 邮箱验证码重置密码。
func (a *App) ResetPassword(email, code, newPassword string) error {
	return a.api.ResetPassword(email, code, newPassword)
}

// Logout 注销：断信令、吊销 refresh。
func (a *App) Logout() error {
	a.mu.Lock()
	rt := a.refreshTok
	a.refreshTok = ""
	a.mu.Unlock()
	a.disconnectSession()
	if rt != "" {
		return a.api.Logout(rt)
	}
	return nil
}

// GetMe 当前用户档案。
func (a *App) GetMe() (*state.UserProfile, error) {
	a.mu.Lock()
	u := a.user
	a.mu.Unlock()
	if u.ID != "" {
		return &u, nil
	}
	return a.api.Me()
}

// ---------- 音频设备（设置页） ----------

// AudioDevice 音频设备描述。
type AudioDevice struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// ListAudioDevices 枚举采集/播放设备。
func (a *App) ListAudioDevices() (map[string][]AudioDevice, error) {
	a.mu.Lock()
	if a.audioCtx == nil {
		ctx, err := audio.NewContext(nil)
		if err != nil {
			a.mu.Unlock()
			return nil, err
		}
		a.audioCtx = ctx
	}
	ctx := a.audioCtx
	a.mu.Unlock()

	caps, err := ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, err
	}
	plays, err := ctx.Devices(malgo.Playback)
	if err != nil {
		return nil, err
	}
	toDTO := func(list []audio.DeviceInfo) []AudioDevice {
		out := make([]AudioDevice, len(list))
		for i, d := range list {
			out[i] = AudioDevice{ID: d.ID, Name: d.Name, IsDefault: d.IsDefault}
		}
		return out
	}
	return map[string][]AudioDevice{
		"capture":  toDTO(caps),
		"playback": toDTO(plays),
	}, nil
}

// ---------- 语音会话 ----------

// connectSession 建立信令会话并把事件接到 Wails Events。
func (a *App) connectSession(token, nickname string) error {
	a.disconnectSession()
	ev := state.Events{
		OnPeerJoin: func(userID, nickname string, speaking, mute, deafen bool) {
			runtime.EventsEmit(a.ctx, "voice:peer-join", map[string]any{
				"user_id": userID, "nickname": nickname,
				"speaking": speaking, "mute": mute, "deafen": deafen,
			})
		},
		OnPeerLeave: func(userID string) {
			runtime.EventsEmit(a.ctx, "voice:peer-leave", userID)
		},
		OnPeerSpeaking: func(userID string, speaking, mute, deafen bool) {
			runtime.EventsEmit(a.ctx, "voice:speaking", map[string]any{
				"user_id": userID, "speaking": speaking, "mute": mute, "deafen": deafen,
			})
		},
		OnConnection: func(st string) {
			runtime.EventsEmit(a.ctx, "rtc:connection", st)
		},
		OnSelf: func(userID string) {
			runtime.EventsEmit(a.ctx, "voice:self", userID)
		},
	}
	server := os.Getenv("LITETALK_WS")
	if server == "" {
		server = "ws://127.0.0.1:8077/ws"
	}
	sess := state.NewSession(a.api, token, nickname, ev, log.Default())
	if err := sess.Connect(server); err != nil {
		return err
	}
	sess.ChatHandler(func(channelID, userID, nick, content, ts string) {
		runtime.EventsEmit(a.ctx, "chat:message", map[string]string{
			"channel_id": channelID, "user_id": userID,
			"nickname": nick, "content": content, "ts": ts,
		})
	})
	a.mu.Lock()
	a.sess = sess
	a.mu.Unlock()
	// 周期推送语音统计（延迟/成员数）。
	go func() {
		tick := time.NewTicker(2 * time.Second)
		defer tick.Stop()
		for range tick.C {
			a.mu.Lock()
			s := a.sess
			a.mu.Unlock()
			if s == nil {
				return
			}
			runtime.EventsEmit(a.ctx, "voice:stats", s.Stats())
		}
	}()
	return nil
}

// disconnectSession 关闭会话。
func (a *App) disconnectSession() {
	a.mu.Lock()
	s := a.sess
	a.sess = nil
	a.mu.Unlock()
	if s != nil {
		s.Close()
	}
}

// JoinChannel 加入语音频道（真实通话链路）。
func (a *App) JoinChannel(channelID string) error {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s == nil {
		return fmt.Errorf("not connected")
	}
	return s.JoinChannel(channelID)
}

// LeaveChannel 离开语音频道。
func (a *App) LeaveChannel() error {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s == nil {
		return nil
	}
	return s.LeaveChannel()
}

// SetMute 本地静音开关。
func (a *App) SetMute(muted bool) {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s != nil {
		s.SetMute(muted)
	}
}

// SetDeafen 耳聋开关。
func (a *App) SetDeafen(deafened bool) {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s != nil {
		s.SetDeafen(deafened)
	}
}

// SetPeerVolume 调整某远端成员音量（0.0–2.0）。
func (a *App) SetPeerVolume(userID string, gain float64) {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s != nil {
		s.SetPeerVolume(userID, gain)
	}
}

// SendChat 发送文字消息。
func (a *App) SendChat(channelID, content string) error {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s == nil {
		return fmt.Errorf("not connected")
	}
	return s.SendChat(channelID, content)
}

// GetVoiceMembers 房间成员快照。
func (a *App) GetVoiceMembers() []state.PeerView {
	a.mu.Lock()
	s := a.sess
	a.mu.Unlock()
	if s == nil {
		return nil
	}
	return s.Members()
}

// ---------- 生命周期 ----------

// shutdown 退出清理。
func (a *App) shutdown(ctx context.Context) {
	a.disconnectSession()
	a.mu.Lock()
	if a.audioCtx != nil {
		_ = a.audioCtx.Close()
		a.audioCtx = nil
	}
	a.mu.Unlock()
}
