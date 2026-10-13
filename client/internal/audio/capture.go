package audio

import (
	"fmt"
	"unsafe"

	"github.com/gen2brain/malgo"
)

// Capture 一路麦克风采集流：48k / mono / int16，回调内零分配零锁。
type Capture struct {
	dev  *malgo.Device
	wake chan struct{} // 容量 1，回调非阻塞投递，唤醒处理 goroutine
}

// StartCapture 打开采集设备并启动流。selector 为空用默认设备。
// 采集到的采样直接写入 ring；ring 满时丢弃并计数。
func StartCapture(ctx *Context, selector string, ring *RingBuffer, stats *Stats) (*Capture, error) {
	devID, err := ctx.resolveDeviceID(malgo.Capture, selector)
	if err != nil {
		return nil, err
	}

	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = SampleRate
	cfg.PeriodSizeInMilliseconds = 10
	if devID != nil {
		// malgo DeviceID 与 C ma_device_id 内存布局一致，直接取 Go 侧数组指针；
		// miniaudio 仅在 InitDevice 期间读取，调用返回后不再持有。
		cfg.Capture.DeviceID = unsafe.Pointer(devID)
	}

	c := &Capture{wake: make(chan struct{}, 1)}
	callbacks := malgo.DeviceCallbacks{
		Data: func(_, inputSamples []byte, _ uint32) {
			stats.CaptureCallbacks.Add(1)
			if n := len(inputSamples) / 2; n > 0 {
				// bytes → int16：unsafe.Slice 只构造切片头（栈上），不拷贝、
				// 不持有 C 内存——RingBuffer.Write 立即把数据拷进预分配缓冲。
				samples := unsafe.Slice((*int16)(unsafe.Pointer(unsafe.SliceData(inputSamples))), n)
				w := ring.Write(samples)
				stats.CaptureSamples.Add(uint64(w))
				if w < n {
					stats.CaptureDropped.Add(uint64(n - w))
				}
			}
			select {
			case c.wake <- struct{}{}:
			default: // 处理 goroutine 尚未消费上次的唤醒，无妨
			}
		},
	}

	dev, err := malgo.InitDevice(ctx.ctx.Context, cfg, callbacks)
	if err != nil {
		return nil, fmt.Errorf("init capture device: %w", err)
	}
	if err := dev.Start(); err != nil {
		dev.Uninit()
		return nil, fmt.Errorf("start capture device: %w", err)
	}
	c.dev = dev
	return c, nil
}

// Wake 返回唤醒通道：每次采集回调投递一次（可合并）。
func (c *Capture) Wake() <-chan struct{} { return c.wake }

// Stop 停流并释放设备。
func (c *Capture) Stop() {
	if c.dev != nil {
		c.dev.Uninit()
		c.dev = nil
	}
}
