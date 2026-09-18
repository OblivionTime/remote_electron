//go:build windows

package input

import "sync"

// KeyboardMap 虚拟键码 -> 按键名称。
// 控制端的键盘钩子用它把 VK 转成名字发出去，被控端反查回 VK 注入，
// 收发两端共用同一张表，命名不会错位。
var KeyboardMap = map[uint32]string{
	8:   "backspace",
	9:   "tab",
	12:  "num_clear",
	13:  "enter",
	19:  "audio_pause",
	20:  "capslock",
	27:  "escape",
	32:  "space",
	33:  "pageup",
	34:  "pagedown",
	35:  "end",
	36:  "home",
	37:  "leftarrow",
	38:  "uparrow",
	39:  "rightarrow",
	40:  "downarrow",
	44:  "printscreen",
	45:  "insert",
	46:  "delete",
	48:  "0",
	49:  "1",
	50:  "2",
	51:  "3",
	52:  "4",
	53:  "5",
	54:  "6",
	55:  "7",
	56:  "8",
	57:  "9",
	65:  "a",
	66:  "b",
	67:  "c",
	68:  "d",
	69:  "e",
	70:  "f",
	71:  "g",
	72:  "h",
	73:  "i",
	74:  "j",
	75:  "k",
	76:  "l",
	77:  "m",
	78:  "n",
	79:  "o",
	80:  "p",
	81:  "q",
	82:  "r",
	83:  "s",
	84:  "t",
	85:  "u",
	86:  "v",
	87:  "w",
	88:  "x",
	89:  "y",
	90:  "z",
	91:  "lcmd",
	92:  "rcmd",
	93:  "contextmenu",
	96:  "numpad_0",
	97:  "numpad_1",
	98:  "numpad_2",
	99:  "numpad_3",
	100: "numpad_4",
	101: "numpad_5",
	102: "numpad_6",
	103: "numpad_7",
	104: "numpad_8",
	105: "numpad_9",
	106: "num*",
	107: "num+",
	109: "num-",
	110: "num.",
	111: "num/",
	112: "f1",
	113: "f2",
	114: "f3",
	115: "f4",
	116: "f5",
	117: "f6",
	118: "f7",
	119: "f8",
	120: "f9",
	121: "f10",
	122: "f11",
	123: "f12",
	144: "num_lock",
	145: "scroll_lock",

	160: "lshift",
	161: "rshift",
	162: "lctrl",
	163: "rctrl",
	164: "lalt",
	165: "ralt",
	186: ";",
	187: "=",
	188: ",",
	189: "-",
	190: ".",
	191: "/",
	192: "`",
	219: "[",
	221: "]",
	220: `\`,
	222: "'",
}

var (
	vkByName map[string]uint16
	vkOnce   sync.Once
)

// VKCode 按名称反查虚拟键码，未知按键返回 0
func VKCode(name string) uint16 {
	if name == "" {
		return 0
	}
	vkOnce.Do(func() {
		vkByName = make(map[string]uint16, len(KeyboardMap))
		for code, n := range KeyboardMap {
			vkByName[n] = uint16(code)
		}
	})
	return vkByName[name]
}
