package utils

import (
	"bytes"
	"image/jpeg"

	"github.com/kbinani/screenshot"
)

func GetScreenshots() []byte {
	// 获取每一个屏幕的边界
	bounds := screenshot.GetDisplayBounds(0)
	// 捕获屏幕内容
	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil
	}
	var byteBuffer bytes.Buffer
	jpeg.Encode(&byteBuffer, img, nil)
	return byteBuffer.Bytes()
}
