//go:build windows

package main

import "time"

// cpuTime Windows 开发机近似值（Getrusage 不可用；压测部署在 Linux，见 cpu_unix.go）。
// 用 WallTime×0 占位保持 API 统一——Windows 上不做 CPU 精确统计。
func cpuTime() time.Duration { return 0 }
