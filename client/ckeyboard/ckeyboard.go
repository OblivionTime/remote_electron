package ckeyboard

import (
	"fmt"
	"remote/global"
	"unsafe"
)

/*
#include <windows.h>

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam);

void ListenKeyboard();
void UnhookKeyboard();
*/
import "C"

var g_hook C.HHOOK
var currentUser string

//export KeyboardProc
func KeyboardProc(nCode C.int, wParam C.WPARAM, lParam C.LPARAM) C.LRESULT {
	if nCode >= 0 && global.OpenWatch {
		kbdStruct := (*C.KBDLLHOOKSTRUCT)(unsafe.Pointer(uintptr(lParam)))
		if wParam == C.WM_KEYDOWN || wParam == C.WM_SYSKEYDOWN {
			// 键下事件
			if _, ok := global.KeyDownList[currentUser]; ok {
				vkCode := uint32(kbdStruct.vkCode)
				global.KeyDownList[currentUser] <- global.KeyboardMap[vkCode]
			}
		} else if wParam == C.WM_KEYUP || wParam == C.WM_SYSKEYUP {
			vkCode := uint32(kbdStruct.vkCode)
			// 键释放事件
			global.KeyRelseList[currentUser] <- global.KeyboardMap[vkCode]
		}
		return 1
	}

	// 其他情况，继续正常处理
	return C.CallNextHookEx(g_hook, nCode, wParam, lParam)
}

//export ListenKeyboard
func ListenKeyboard() {
	if g_hook != nil {
		return
	}
	global.OpenWatch = true
	g_hook = C.SetWindowsHookEx(C.WH_KEYBOARD_LL, C.HOOKPROC(C.KeyboardProc), C.GetModuleHandle(nil), 0)
	if g_hook == nil {
		// 处理钩子设置失败的情况
		return
	}
	fmt.Println("开启监听模式")
	// 进入消息循环
	var msg C.MSG
	for C.GetMessage(&msg, nil, 0, 0) != 0 {
		C.TranslateMessage(&msg)
		C.DispatchMessage(&msg)
	}

}

//export UnhookKeyboard
func UnhookKeyboard() {

	if g_hook != nil {
		C.UnhookWindowsHookEx(g_hook)
		g_hook = nil
		fmt.Println("关闭监听模式")
	}
}
func ChangeUser(user string) {
	currentUser = user
}
