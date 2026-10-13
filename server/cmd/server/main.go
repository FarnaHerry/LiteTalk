// 轻音 LiteTalk 服务端入口
// M0 阶段：最小可运行 HTTP 服务，仅提供健康检查。
// 后续里程碑扩展：REST API（chi）、WebSocket 信令、自研 Pion SFU。
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// version 由构建时 -X main.version= 注入，默认 dev
var version = "dev"

// 默认监听地址，可用环境变量 LITETALK_ADDR 覆盖
const defaultAddr = ":8077"

func main() {
	addr := os.Getenv("LITETALK_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", healthzHandler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 捕获 SIGINT / SIGTERM，触发优雅关闭
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("litetalk server %s 监听 %s", version, addr)
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
