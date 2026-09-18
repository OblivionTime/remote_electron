//go:build windows

// Package ckeyboard 通过 Windows 底层键盘钩子（WH_KEYBOARD_LL）捕获控制端的按键，
// 交给 api/server 的转发 goroutine 发往被控端。
package ckeyboard

import (
	"fmt"
	"remote/global"
	"remote/input"
	"sync"
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	setWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	callNextHookEx      = user32.NewProc("CallNextHookEx")
	unhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	getMessage          = user32.NewProc("GetMessageW")
	translateMessage    = user32.NewProc("TranslateMessage")
	dispatchMessage     = user32.NewProc("DispatchMessageW")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
)

type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}
type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var (
	g_hook uintptr

	userMu      sync.RWMutex
	currentUser string
)

func keyboardProc(nCode int32, wParam uintptr, lParam unsafe.Pointer) uintptr {
	if nCode >= 0 && global.OpenWatch {
		user, ok := sessionUser()
		if ok {
			downCh, upCh, ok := global.KeyboardChannels(user)
			if ok {
				vk := uint32((*KBDLLHOOKSTRUCT)(lParam).VkCode)
				name := input.KeyboardMap[vk]
				switch wParam {
				case WM_KEYDOWN, WM_SYSKEYDOWN:
					trySend(downCh, name)
				case WM_KEYUP, WM_SYSKEYUP:
					trySend(upCh, name)
				}
				// 拦截本地按键，避免同时在本机生效
				return 1
			}
		}
	}
	ret, _, _ := callNextHookEx.Call(g_hook, uintptr(nCode), wParam, uintptr(lParam))
	return ret
}

// trySend 非阻塞投递：消费端已退出时直接丢弃，
// 避免往无人接收（或已关闭）的 channel 发送把钩子线程永久卡死。
func trySend(ch chan string, name string) {
	select {
	case ch <- name:
	default:
	}
}

func sessionUser() (string, bool) {
	userMu.RLock()
	defer userMu.RUnlock()
	return currentUser, currentUser != ""
}

func ListenKeyboard() {
	if g_hook != 0 {
		return
	}
	// WH_KEYBOARD_LL；第二个参数必须是我们自己的回调地址，
	// 传 0 会装不上钩子（原来 cgo 版本用的是 C.HOOKPROC(C.KeyboardProc)）
	hook, _, err := setWindowsHookEx.Call(WH_KEYBOARD_LL, syscall.NewCallback(keyboardProc), 0, 0)
	if hook == 0 {
		fmt.Println("安装键盘钩子失败:", err)
		return
	}
	g_hook = hook

	var msg MSG
	for {
		ret, _, _ := getMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 {
			break
		}
		translateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		dispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func UnhookKeyboard() {
	if g_hook != 0 {
		unhookWindowsHookEx.Call(g_hook)
		g_hook = 0
	}
}

func ChangeUser(user string) {
	userMu.Lock()
	currentUser = user
	userMu.Unlock()
}
