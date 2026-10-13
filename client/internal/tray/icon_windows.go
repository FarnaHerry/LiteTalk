//go:build windows

package tray

import "image"

// iconICO 生成 32x32 托盘图标（品牌蓝声波条，ICO 格式字节）。
// 仅在内存绘制，不依赖外部资源文件。
func iconICO() []byte {
	const size = 32
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// 品牌蓝 #4FA4F3 声波条：5 根竖条，中间高两边低。
	barColor := [4]uint8{0x4F, 0xA4, 0xF3, 0xFF}
	bars := []struct{ x, h int }{
		{7, 10}, {12, 18}, {17, 24}, {22, 18}, {27, 10},
	}
	for _, b := range bars {
		top := (size - b.h) / 2
		for y := top; y < top+b.h; y++ {
			for x := b.x - 1; x <= b.x+1; x++ {
				i := img.PixOffset(x, y)
				img.Pix[i+0] = barColor[0]
				img.Pix[i+1] = barColor[1]
				img.Pix[i+2] = barColor[2]
				img.Pix[i+3] = barColor[3]
			}
		}
	}

	return encodeICO(img)
}

// encodeICO 把 NRGBA 图像编码为单帧 32x32 32bpp ICO 字节。
func encodeICO(img *image.NRGBA) []byte {
	const size = 32
	xorSize := size * size * 4
	andRow := (size + 31) / 32 * 4 // AND 掩码行按 32 位对齐
	andSize := andRow * size
	dataSize := 40 + xorSize + andSize

	ico := make([]byte, 6+16+dataSize)
	// ICONDIR
	ico[2], ico[3] = 1, 0 // type = icon
	ico[4], ico[5] = 1, 0 // count = 1
	// ICONDIRENTRY
	ico[6] = size
	ico[7] = size
	ico[11] = 32 // bitcount
	put32(ico[14:], uint32(dataSize))
	put32(ico[18:], 22) // 图像数据偏移

	// BITMAPINFOHEADER（height 含 AND 掩码翻倍，bottom-up）
	d := ico[22:]
	put32(d[0:], 40)
	put32(d[4:], size)
	put32(d[8:], size*2)
	d[12], d[13] = 1, 0 // planes
	d[14], d[15] = 32, 0

	// XOR 位图：BGRA 自底向上
	xor := d[40:]
	for y := 0; y < size; y++ {
		src := img.PixOffset(0, size-1-y)
		dst := y * size * 4
		for x := 0; x < size; x++ {
			xor[dst+x*4+0] = img.Pix[src+x*4+2] // B
			xor[dst+x*4+1] = img.Pix[src+x*4+1] // G
			xor[dst+x*4+2] = img.Pix[src+x*4+0] // R
			xor[dst+x*4+3] = img.Pix[src+x*4+3] // A
		}
	}
	// AND 掩码全 0：透明度由 alpha 通道决定
	return ico
}

func put32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
