package utils

import (
	"encoding/json"
	"remote/global"

	"gitee.com/solidone/sutils/sencryption"
)

var X = string([]byte{224, 130, 44, 118, 90, 64, 194, 254, 132, 203, 16, 162, 60, 98, 138, 210, 133, 127, 81, 251, 99, 33, 185, 109, 228, 56, 236, 246, 188, 40, 236, 45})
var Y = string([]byte{35, 94, 180, 165, 65, 11, 166, 41, 215, 187, 165, 69, 202, 219, 158, 136})

func Decrypt(text []byte, data interface{}) {
	res, _ := sencryption.AesDecrypt(text, X, Y)
	json.Unmarshal(res, data)
}

func CloseAllConnect() {
	if global.ClientConn != nil {
		global.ClientConn.Send <- map[string]string{"operation": "disconnect"}
		global.ClientConn.CloseConn()
		global.ClientConn = nil
	}
	if global.VideoConn != nil {
		global.VideoConn.CloseConn()
		global.VideoConn = nil
	}
	if global.KeyboardConn != nil {
		global.KeyboardConn.CloseConn()
		global.KeyboardConn = nil
	}
	if global.KeyboardP2PConn != nil {
		global.KeyboardP2PConn.Close()
		global.KeyboardP2PConn = nil
	}
	if global.KeyboardHandler != nil {
		global.KeyboardHandler.Close()
		global.KeyboardHandler = nil
	}
}
