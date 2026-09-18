package client

import (
	"remote/global"
	"remote/model"

	"gitee.com/solidone/sutils/swebsocket"
)

// VideoHandler 处理来自中转服务器的音视频信令。
// 视频的 PeerConnection 实际跑在 Electron 渲染进程里，这里只做信令转发；
// 键盘 P2P 则跑在本进程内。
func VideoHandler(msg model.Signal) {
	switch msg.Op {
	case "join":
		//共享屏幕给连接的设备
		if global.ClientConn != nil {
			global.ClientConn.Send <- map[string]interface{}{
				"operation":  "video",
				"device":     msg.Device,
				"iceservers": msg.ICEServers,
			}
			startKeyboardReceiver(msg.ICEServers, msg.Device)
		}
	case "ice_server":
		if global.VideoConn != nil {
			global.VideoConn.Send <- map[string]interface{}{
				"operation":  "ice_server",
				"iceservers": msg.ICEServers,
				"device":     msg.Device,
			}
			startKeyboardSender(msg.ICEServers, msg.Device)
		}
	case "offer":
		forwardTo(global.VideoConn, "offer", msg)
	case "answer":
		forwardTo(global.ClientConn, "answer", msg)
	case "quality":
		// 控制端下发的画质预设，转给被控端 UI 去调编码参数
		forwardTo(global.ClientConn, "quality", msg)
	case "quality_ack":
		// 被控端应用结果，回给控制端确认
		forwardTo(global.VideoConn, "quality_ack", msg)
	case "ice_candidate":
		// VideoSender=false 表示来自被控端，回给控制端；反之回给被控端
		conn := global.VideoConn
		if msg.VideoSender {
			conn = global.ClientConn
		}
		forwardTo(conn, "ice_candidate", msg)
	}
}

func forwardTo(conn *swebsocket.ServerConn, operation string, msg model.Signal) {
	if conn == nil {
		return
	}
	conn.Send <- map[string]interface{}{
		"operation": operation,
		"data":      msg.Data,
		"device":    msg.Device,
	}
}
