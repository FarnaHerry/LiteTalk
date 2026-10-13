# 轻音 LiteTalk

桌面语音联机应用。客户端为 Go + Wails v2 + Vue 3，服务端为 Go 单体（REST API + WebSocket 信令 + 自研 Pion SFU）。完整设计文档仅保存在团队本地工作区，不入库。

## 目录结构

- `client/` Wails 桌面客户端（`frontend/` 为 Vue 3 + Vite + TS 前端，`build/nfpm.yaml` 为 Linux 打包配置）
- `server/` Go 服务端单体（部署配置、SQL 迁移等后续置于 `server/deploy/`、`server/migrations/`）

## 本地开发

前置：Go 1.27+、Node 22+、pnpm 9。国内开发先配置镜像：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
# 前端镜像已内置在项目级 client/frontend/.npmrc（npmmirror），无需全局配置
```

### 服务端

```bash
cd server
go run ./cmd/server   # 监听 :8077，可用环境变量 LITETALK_ADDR 覆盖
go test ./...
```

健康检查：`GET http://localhost:8077/api/v1/healthz`

### 客户端

需要安装 wails CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）：

```bash
cd client/frontend && pnpm install   # 首次
cd client
wails dev                            # 开发模式（前端热更新）
wails build                          # 出包到 client/build/bin/
```

不装 wails CLI 也可以只验证编译：先 `cd client/frontend && pnpm install && pnpm run build`（`go:embed` 依赖 dist 产物），再 `cd client && go vet ./... && go test ./... && go build ./...`。

## CI/CD

`.github/workflows/ci.yml` 为单共享工作流：

- push 任意分支或 `workflow_dispatch`：服务端/客户端 vet + test + build 检查，不打包
- 推送 `v*` 标签：在检查之上追加全平台打包（Windows 安装器/绿色版、Linux deb/rpm/tar.gz、macOS zip[非阻塞]、服务端 tar.gz），并汇总发布 GitHub Release
