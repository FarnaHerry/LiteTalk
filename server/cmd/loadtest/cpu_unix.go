//go:build !windows

package main

import (
	"syscall"
	"time"
)

// cpuTime 进程累计 CPU（user+sys）。部署目标是 Linux，Windows 用 cpu_windows.go。
func cpuTime() time.Duration {
	var ru syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}
