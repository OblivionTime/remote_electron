//go:build windows

package input

import (
	"fmt"

	"remote/model"
)

// mapValues 把控制端画面坐标映射到本机屏幕坐标
func mapValues(value, inMin, inMax, outMin, outMax float64) int {
	return int((value-inMin)*(outMax-outMin)/(inMax-inMin) + outMin)
}

// Operation 执行一条来自数据通道的键鼠 / 剪贴板指令
func Operation(op model.InputOperation) error {
	switch op.Method {
	case "MouseMove":
		width, height := GetScreenSize()
		x, y := op.Arg1, op.Arg2
		if op.Width > 0 && op.Height > 0 {
			x = mapValues(float64(op.Arg1), 0, op.Width, 0, float64(width))
			y = mapValues(float64(op.Arg2), 0, op.Height, 0, float64(height))
		}
		Move(x, y)
	case "MouseDown":
		switch op.Arg1 {
		case 1:
			return Toggle("left")
		case 2:
			return Toggle("center")
		default:
			return Toggle("right")
		}
	case "MouseUp":
		switch op.Arg1 {
		case 1:
			return Toggle("left", "up")
		case 2:
			return Toggle("center", "up")
		default:
			return Toggle("right", "up")
		}
	case "ScrollMouse":
		if op.Arg1 == -1 {
			return ScrollDir(1, "down")
		}
		return ScrollDir(1, "up")
	case "KeyDown":
		return KeyDown(op.Key)
	case "KeyRelease":
		return KeyUp(op.Key)
	case "clipboard":
		for _, s := range op.ClipboardData {
			WriteAll(s)
		}
		return nil
	default:
		return fmt.Errorf("未知输入指令: %s", op.Method)
	}
	return nil
}
