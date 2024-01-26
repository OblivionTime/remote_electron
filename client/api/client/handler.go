package client

import (
	"encoding/json"
	"remote/global"

	"gitee.com/solidone/sutils/swebsocket"
)

type HandlerResult struct {
	Op           string `json:"op"` //操作
	Device       string `json:"device,omitempty"`
	Code         string `json:"code,omitempty"`
	SendDevice   string `json:"send_device,omitempty"`
	KeyboardData []byte `json:"keyboard_data,omitempty"`
	VideoData    []byte `json:"videoData,omitempty"`
}

// 数据处理
func RemoteDataHandler(res []byte, conn *swebsocket.ServerConn) {
	var msg HandlerResult
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join":
		//共享屏幕给连接的设备
		go VideoHandler(msg.Device, conn)
	case "video":
		if global.VideoConn != nil {
			global.VideoConn.Send <- msg.VideoData
		}
	case "disconnected":
		ConnectDevice[msg.SendDevice] = true
	case "keyboard":
		HandlerKeyboard(msg.KeyboardData)
	}
}
