//go:build windows

// Package input 负责把远端的键鼠操作注入本机，并持有虚拟键码表。
package input

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	setCursorPos     = user32.NewProc("SetCursorPos")
	getSystemMetrics = user32.NewProc("GetSystemMetrics")
	sendInput        = user32.NewProc("SendInput")
)

const (
	SM_CXSCREEN            = 0
	SM_CYSCREEN            = 1
	INPUT_MOUSE            = 0
	INPUT_KEYBOARD         = 1
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0080
	MOUSEEVENTF_WHEEL      = 0x0800
	KEYEVENTF_KEYUP        = 0x0002
	WHEEL_DELTA            = 120
)

type MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// Windows 的 INPUT 是 union：MOUSEINPUT(32B) 与 KEYBDINPUT(24B) 共用同一块内存，整体 40 字节。
// 若按字段顺序平铺定义，unsafe.Sizeof 会变成 96，SendInput 会因 cbSize 不合法而直接失败。
// 这里用最大的 MOUSEINPUT 占位，键盘数据按 KEYBDINPUT 覆盖写入同一内存。
type INPUT struct {
	Type uint32
	_    uint32
	Mi   MOUSEINPUT
}

func GetScreenSize() (int, int) {
	ret, _, _ := getSystemMetrics.Call(uintptr(SM_CXSCREEN))
	cx := int(ret)
	ret, _, _ = getSystemMetrics.Call(uintptr(SM_CYSCREEN))
	cy := int(ret)
	return cx, cy
}

func Move(x, y int) {
	setCursorPos.Call(uintptr(x), uintptr(y))
}

func sendInputStruct(in INPUT) error {
	ret, _, err := sendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if ret == 0 {
		return fmt.Errorf("SendInput 失败: %v", err)
	}
	return nil
}

func Toggle(button string, up ...string) error {
	var flags uint32
	switch button {
	case "left":
		if len(up) > 0 && up[0] == "up" {
			flags = MOUSEEVENTF_LEFTUP
		} else {
			flags = MOUSEEVENTF_LEFTDOWN
		}
	case "right":
		if len(up) > 0 && up[0] == "up" {
			flags = MOUSEEVENTF_RIGHTUP
		} else {
			flags = MOUSEEVENTF_RIGHTDOWN
		}
	case "center":
		if len(up) > 0 && up[0] == "up" {
			flags = MOUSEEVENTF_MIDDLEUP
		} else {
			flags = MOUSEEVENTF_MIDDLEDOWN
		}
	default:
		return nil
	}
	return sendInputStruct(INPUT{Type: INPUT_MOUSE, Mi: MOUSEINPUT{DwFlags: flags}})
}

func ScrollDir(n int, dir string) error {
	delta := int32(WHEEL_DELTA)
	if dir == "down" {
		delta = -WHEEL_DELTA
	}
	return sendInputStruct(INPUT{Type: INPUT_MOUSE, Mi: MOUSEINPUT{MouseData: uint32(delta), DwFlags: MOUSEEVENTF_WHEEL}})
}

func sendKey(vk uint16, up bool) error {
	ki := KEYBDINPUT{WVk: vk}
	if up {
		ki.DwFlags = KEYEVENTF_KEYUP
	}
	in := INPUT{Type: INPUT_KEYBOARD}
	*(*KEYBDINPUT)(unsafe.Pointer(&in.Mi)) = ki
	return sendInputStruct(in)
}

func KeyDown(key string) error {
	vk := VKCode(key)
	if vk == 0 {
		return fmt.Errorf("未知按键: %s", key)
	}
	return sendKey(vk, false)
}

func KeyUp(key string) error {
	vk := VKCode(key)
	if vk == 0 {
		return fmt.Errorf("未知按键: %s", key)
	}
	return sendKey(vk, true)
}

// WriteAll 逐字符模拟键盘输入（剪贴板同步用）
func WriteAll(s string) {
	for _, r := range s {
		sendKey(uint16(r), false)
		sendKey(uint16(r), true)
	}
}
