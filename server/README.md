# LiteTalk Server（M2）

轻音 LiteTalk 服务端：WS 信令 hub + 自研 Pion SFU + TURN 短期凭证签发。
协议契约见 `docs/落地方案.md` §5.2（SFU）/ §5.3（信令）/ §5.5（通话时序）。

## 本地启动

```bash
cd server
go run ./cmd/server                 # 默认监听 :8077
LITETALK_ADDR=:18077 go run ./cmd/server   # 自定义端口
```

| 环境变量 | 默认 | 说明 |
|---|---|---|
| `LITETALK_ADDR` | `:8077` | HTTP/WS 监听地址 |
| `LITETALK_SFU_UDP` | `7845` | SFU ICE-Lite UDP mux 端口（0 = 随机，测试用） |
| `LITETALK_SFU_IP` | `127.0.0.1` | SFU 对外宣告 IP（部署填公网 IP） |
| `LITETALK_SFU_MAX` | `10` | 语音房间人数上限 |
| `LITETALK_TURN_SECRET` | （空） | coturn `static-auth-secret`；空则凭证接口返回 503 |
| `LITETALK_TURN_URLS` | `turn:127.0.0.1:3478` | TURN URL，逗号分隔 |
| `LITETALK_STUN_URLS` | `stun:127.0.0.1:3478` | STUN URL，逗号分隔 |

健康检查：`curl http://127.0.0.1:8077/api/v1/healthz`

## wscat 拨测 join

```bash
npm i -g wscat
wscat -c "ws://127.0.0.1:8077/ws?token=dev-alice"
```

连接后服务端先下发 `hello`，随后逐条粘贴：

```json
{"type":"channel.join","data":{"channel_id":"demo"},"ts":0}
{"type":"voice.state","data":{"speaking":true},"ts":0}
{"type":"channel.leave","data":{"channel_id":"demo"},"ts":0}
```

预期回包：`channel.joined`（成员全量，含 `self`）→ 其他成员收到 `voice.state` → 离开后其余成员收到 `channel.peer_left`。
M2 开发令牌格式 `dev-<用户名>`，无签名校验（M3 换 JWT）。

也可以用仓库自带的 Go 拨测（join + 收 hello + leave 一气呵成）：

```bash
go run ./cmd/wsprobe -addr 127.0.0.1:8077 -token dev-alice -room demo
```

## TURN 短期凭证

```bash
curl -X POST "http://127.0.0.1:8077/api/v1/rtc/credentials?token=dev-alice"
```

返回（coturn `use-auth-secret` / TURN REST API 格式，12h 有效）：

```json
{
  "iceServers": [
    {"urls": ["stun:127.0.0.1:3478"]},
    {"urls": ["turn:127.0.0.1:3478"], "username": "1734543200:alice", "credential": "<base64(HMAC-SHA1(username, TURN_SECRET))>"}
  ],
  "expiresAt": 1734543200
}
```

鉴权：`Authorization: Bearer <token>` 或 `?token=`。需先配置 `LITETALK_TURN_SECRET`（与 coturn 的 `static-auth-secret` 一致）。

## 自验

```bash
go vet ./... && go test -count=1 ./... && go build ./...
```

- `internal/sfu`：双客户端 RTP 逐字节透传、第三人加入重协商、断连停止转发、房间满员（真实 Pion ICE/DTLS/SRTP）
- `internal/rtcapi`：HMAC-SHA1 签名（RFC 2202 向量）、凭证格式、HTTP 鉴权/错误分支
- `internal/signal`：WS 信令 hub（经真实服务端冒烟覆盖）
