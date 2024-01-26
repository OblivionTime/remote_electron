package utils

import (
	"encoding/json"

	"gitee.com/solidone/sutils/sencryption"
)

var X = string([]byte{224, 130, 44, 118, 90, 64, 194, 254, 132, 203, 16, 162, 60, 98, 138, 210, 133, 127, 81, 251, 99, 33, 185, 109, 228, 56, 236, 246, 188, 40, 236, 45})
var Y = string([]byte{35, 94, 180, 165, 65, 11, 166, 41, 215, 187, 165, 69, 202, 219, 158, 136})

func Decrypt(text []byte, data interface{}) {
	res, _ := sencryption.AesDecrypt(text, X, Y)
	json.Unmarshal(res, data)
}
