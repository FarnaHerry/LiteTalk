//go:build windows

package tray

import _ "embed"

// iconICO 静态托盘图标（32x32 品牌蓝声波条，gen_icon.go 生成的合法 BMP+AND ICO）。
// go:embed 编译期嵌入，避免运行时内存编码在真机上渲染不稳的问题。
//
//go:embed icon.ico
var iconICOBytes []byte

// iconICO 返回嵌入的 ICO 字节。
func iconICO() []byte { return iconICOBytes }
