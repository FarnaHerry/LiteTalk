// 轻音 LiteTalk 服务端入口
// M2 阶段：healthz + WS 信令（GET /ws）+ 自研 Pion SFU + TURN 短期凭证（POST /api/v1/rtc/credentials）。
// 后续里程碑扩展：REST API（chi）、JWT 鉴权、MySQL/Redis。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	ossignal "os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/OpenALC/LiteTalk/server/internal/rtcapi"
	"github.com/OpenALC/LiteTalk/server/internal/sfu"
	"github.com/OpenALC/LiteTalk/server/internal/signal"
)

// version 由构建时 -X main.version= 注入，默认 dev
var version = "dev"

// 默认监听地址，可用环境变量 LITETALK_ADDR 覆盖
const defaultAddr = ":8077"

func main() {
	addr := envOr("LITETALK_ADDR", defaultAddr)

	// SFU：ICE-Lite + 单 UDP mux（§5.2）
	sfuInst, err := sfu.New(sfu.Config{
		UDPPort:    envInt("LITETALK_SFU_UDP", 7845),
		NAT1To1IP:  envOr("LITETALK_SFU_IP", "127.0.0.1"),
		MaxMembers: envInt("LITETALK_SFU_MAX", 10),
		Logger:     log.Default(),
	})
	if err != nil {
		log.Fatalf("SFU 初始化失败: %v", err)
	}
	defer sfuInst.Close()

	// WS 信令 hub（M2 开发令牌 dev-<username>，M3 换 JWT）
	hub := signal.NewHub(sfuInst, signal.DevAuthenticator, log.Default(), 0)

	// TURN 短期凭证签发（coturn use-auth-secret，§5.4）
	rtcCreds := rtcapi.NewHandler(rtcapi.Config{
		Secret:   os.Getenv("LITETALK_TURN_SECRET"),
		TurnURLs: splitList(envOr("LITETALK_TURN_URLS", "turn:127.0.0.1:3478")),
		StunURLs: splitList(envOr("LITETALK_STUN_URLS", "stun:127.0.0.1:3478")),
		TTL:      12 * time.Hour,
	}, devUserID, log.Default())

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", healthzHandler)
	mux.Handle("POST /api/v1/rtc/credentials", rtcCreds)
	mux.HandleFunc("GET /ws", hub.ServeWS)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 捕获 SIGINT / SIGTERM，触发优雅关闭
	ctx, stop := ossignal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("litetalk server %s 监听 %s（SFU UDP %d）", version, addr, sfuInst.Port())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务异常退出: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("收到退出信号，开始优雅关闭")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("优雅关闭失败: %v", err)
	}
	log.Println("服务已停止")
}

// healthzHandler 健康检查：GET /api/v1/healthz
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"version": version,
	})
}

// devUserID 把 M2 开发令牌映射为用户 ID（M3 换 JWT 时同步替换）。
func devUserID(token string) (string, error) {
	id, _, err := signal.DevAuthenticator(token)
	return id, err
}

// ---- 环境变量小工具 ----

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("%s 不是合法整数: %q", key, v)
	}
	return n
}

func splitList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
