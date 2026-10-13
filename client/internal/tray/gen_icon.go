//go:build ignore

// 一次性生成器：产出 client/internal/tray/icon.ico（32x32 品牌蓝声波条，单帧 32bpp BMP+AND）。
// go run gen_icon.go
package main

import (
	"image"
	"log"
	"os"
)

func main() {
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

	ico := encodeICO(img)
	if err := os.WriteFile("icon.ico", ico, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("icon.ico written: %d bytes", len(ico))
}

// encodeICO 编码单帧 32x32 32bpp BMP+AND ICO（与 tray 包同款逻辑）。
func encodeICO(img *image.NRGBA) []byte {
	const size = 32
	xorSize := size * size * 4
	andRow := (size + 31) / 32 * 4 // AND 掩码行按 32 位对齐
	andSize := andRow * size
	dataSize := 40 + xorSize + andSize

	out := make([]byte, 0, 6+16+dataSize)
	// ICONDIR
	out = append(out, 0, 0, 1, 0, 1, 0)
	// ICONDIRENTRY（32bpp BMP 行高需 ×2）
	out = append(out,
		byte(size), byte(size), 0, 0,
		1, 0, 32, 0,
		byte(dataSize), byte(dataSize>>8), byte(dataSize>>16), byte(dataSize>>24),
		22, 0, 0, 0, // 数据偏移 = 6 + 16
	)
	// BITMAPINFOHEADER
	bi := make([]byte, 40)
	bi[0] = 40
	bi[4] = byte(size)
	bi[8] = byte(size * 2) // 高度 = XOR + AND
	bi[12] = 1             // planes
	bi[14] = 32            // bpp
	out = append(out, bi...)
	// XOR：自底向上 BGRA
	for y := size - 1; y >= 0; y-- {
		for x := 0; x < size; x++ {
			i := img.PixOffset(x, y)
			out = append(out, img.Pix[i+2], img.Pix[i+1], img.Pix[i+0], img.Pix[i+3])
		}
	}
	// AND：alpha 即掩码，透明处置 1
	for y := size - 1; y >= 0; y-- {
		row := make([]byte, andRow)
		for x := 0; x < size; x++ {
			if img.Pix[img.PixOffset(x, y)+3] < 128 {
				row[x/8] |= 1 << (7 - x%8)
			}
		}
		out = append(out, row...)
	}
	return out
}
