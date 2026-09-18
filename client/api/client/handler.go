package client

import (
	"encoding/json"
	"fmt"
	"remote/global"
	"remote/model"

	"gitee.com/solidone/sutils/swebsocket"
)

// RemoteDataHandler 处理中转服务器下发的所有消息
func RemoteDataHandler(res []byte, conn *swebsocket.ServerConn) {
	var msg model.Signal
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join", "offer", "answer", "ice_candidate", "ice_server", "quality", "quality_ack":
		VideoHandler(msg)
	case "disconnected":
		if !msg.VideoSender {
			forwardTo(global.VideoConn, "disconnected", msg)
		} else {
			forwardTo(global.ClientConn, "video_disconnected", msg)
		}
	case "keyboard":
		HandlerKeyboard(msg)
	case "file_join":
		go HandlerFileShare(msg)
	default:
		fmt.Printf("参数错误%+v\n", msg)
	}
}
