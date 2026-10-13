// wsprobe 最小 WS 拨测：连上 /ws → 收 hello → channel.join → 收 channel.joined → channel.leave。
// 用法：go run ./cmd/wsprobe -addr 127.0.0.1:18077 -token dev-alice -room demo
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/coder/websocket"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8077", "服务端地址")
	token := flag.String("token", "dev-alice", "访问令牌（M2: dev-<用户名>）")
	room := flag.String("room", "demo", "频道 ID")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/ws?token=%s", *addr, *token), nil)
	if err != nil {
		log.Fatalf("拨号失败: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "bye")

	read := func(step string) {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			log.Fatalf("%s 读失败: %v", step, err)
		}
		fmt.Printf("%s ← %s\n", step, msg)
	}
	read("hello") // 连接后服务端先下发 hello
	_ = conn.Write(ctx, websocket.MessageText,
		[]byte(`{"type":"channel.join","data":{"channel_id":"`+*room+`"},"ts":0}`))
	read("joined") // channel.joined（成员全量）
	_ = conn.Write(ctx, websocket.MessageText,
		[]byte(`{"type":"channel.leave","data":{"channel_id":"`+*room+`"},"ts":0}`))
	fmt.Println("leave 已发送，拨测通过")
}
