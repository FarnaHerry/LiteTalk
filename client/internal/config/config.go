// Package config 管理客户端本地配置（config.json 持久化）。
//
// 服务器地址优先级：环境变量 LITETALK_API/LITETALK_WS > 配置文件 > 默认本机地址。
// 环境变量优先级最高（运维/调试覆盖用）。
package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Defaults 默认服务器地址（本机开发）。
const (
	DefaultBaseURL = "http://127.0.0.1:8077"
	DefaultWSURL   = "ws://127.0.0.1:8077/ws"
)

// Env 环境变量名（覆盖配置文件用）。
const (
	EnvAPI = "LITETALK_API"
	EnvWS  = "LITETALK_WS"
)

// Source 配置来源标识（env / file / default）。
type Source string

const (
	SourceEnv     Source = "env"
	SourceFile    Source = "file"
	SourceDefault Source = "default"
)

// Config config.json 的持久化结构。
type Config struct {
	ServerBase        string `json:"server_base,omitempty"`         // 服务端基址，如 https://example.com
	HibernateOnMinimize *bool `json:"hibernate_on_minimize,omitempty"` // 最小化时休眠界面
}

// Manager 读写配置文件（路径：os.UserConfigDir()/LiteTalk/config.json）。
type Manager struct {
	path string

	mu sync.Mutex
	cfg Config
}

// New 在默认路径创建 Manager 并加载（文件损坏/缺失时静默回默认值）。
func New() *Manager {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return NewWithPath(filepath.Join(dir, "LiteTalk", "config.json"))
}

// NewWithPath 指定路径创建 Manager（测试用）。
func NewWithPath(path string) *Manager {
	m := &Manager{path: path}
	m.cfg = m.read()
	return m
}

// Path 返回配置文件路径。
func (m *Manager) Path() string { return m.path }

// read 读盘；任何失败（缺失/损坏）都返回零值配置，不报错。
func (m *Manager) read() Config {
	data, err := os.ReadFile(m.path)
	if err != nil {
		return Config{}
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return Config{} // 损坏：回默认值不报错
	}
	return c
}

// Save 原子写盘（临时文件 + rename）。
func (m *Manager) Save(c Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, m.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	m.cfg = c
	return nil
}

// Get 返回当前配置快照。
func (m *Manager) Get() Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

// Update 读-改-写配置（f 返回 false 则放弃保存）。
func (m *Manager) Update(f func(*Config) error) error {
	m.mu.Lock()
	c := m.cfg
	m.mu.Unlock()
	if err := f(&c); err != nil {
		return err
	}
	return m.Save(c)
}

// Resolved 服务器地址解析结果。
type Resolved struct {
	BaseURL string `json:"base_url"`
	WSURL   string `json:"ws_url"`
	Source  Source `json:"source"` // env / file / default
}

// Resolve 按优先级解析服务器地址：env > file > default。
// WS 地址：env LITETALK_WS 优先；否则由 base 推导（DeriveWS）。
func (m *Manager) Resolve() Resolved {
	envBase := strings.TrimRight(strings.TrimSpace(os.Getenv(EnvAPI)), "/")
	envWS := strings.TrimSpace(os.Getenv(EnvWS))

	fileBase := strings.TrimRight(strings.TrimSpace(m.Get().ServerBase), "/")

	base, source := DefaultBaseURL, SourceDefault
	if fileBase != "" {
		base, source = fileBase, SourceFile
	}
	if envBase != "" {
		base, source = envBase, SourceEnv
	}

	ws := ""
	switch {
	case envWS != "":
		ws = envWS
		source = SourceEnv
	default:
		ws, _ = DeriveWS(base) // base 已保证合法时推导必然成功；异常时留空由调用方兜底
		if ws == "" {
			ws = DefaultWSURL
		}
	}
	return Resolved{BaseURL: base, WSURL: ws, Source: source}
}

// DeriveWS 由 http(s) 基址推导 ws(s) 信令地址。
// 规则：https→wss、http→ws；末尾追加 /ws（已带 /ws 结尾则不重复拼）。
func DeriveWS(base string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" {
		return "", errors.New("服务器地址无效")
	}
	var wsScheme string
	switch u.Scheme {
	case "https":
		wsScheme = "wss"
	case "http":
		wsScheme = "ws"
	default:
		return "", errors.New("服务器地址需以 http:// 或 https:// 开头")
	}
	path := strings.TrimRight(u.Path, "/")
	if path == "" || path == "/ws" {
		path = "/ws"
	} else {
		path = path + "/ws"
	}
	return wsScheme + "://" + u.Host + path, nil
}

// ValidateBase 校验服务器基址合法性（http/https 前缀 + 可解析）。
func ValidateBase(base string) error {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Host == "" {
		return errors.New("服务器地址无效")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("服务器地址需以 http:// 或 https:// 开头")
	}
	return nil
}
