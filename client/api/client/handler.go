package client

import (
	"encoding/json"
	"fmt"

	"gitee.com/solidone/sutils/swebsocket"
)

type HandlerResult struct {
	Op           string      `json:"op"` //操作
	Device       string      `json:"device,omitempty"`
	Code         string      `json:"code,omitempty"`
	SendDevice   string      `json:"send_device,omitempty"`
	KeyboardData []byte      `json:"keyboard_data,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	VideoSender  bool        `json:"videoSender,omitempty"`
}

// 数据处理
func RemoteDataHandler(res []byte, conn *swebsocket.ServerConn) {
	var msg HandlerResult
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join", "offer", "answer", "ice_candidate":
		VideoHandler(msg)
	case "disconnected":
		ConnectDevice[msg.SendDevice] = true
	case "keyboard":
		HandlerKeyboard(msg.KeyboardData)
	default:
		fmt.Printf("参数错误%+v\n", msg)
	}
}
