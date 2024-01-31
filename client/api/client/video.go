package client

import (
	"remote/global"
	"remote/utils"
)

var ConnectDevice = make(map[interface{}]bool)

func VideoHandler(msg HandlerResult) {
	switch msg.Op {
	case "join":
		//共享屏幕给连接的设备
		if global.ClientConn != nil {
			global.ClientConn.Send <- map[string]interface{}{
				"operation":  "video",
				"device":     msg.Device,
				"iceservers": msg.ICEServers,
			}
			utils.InitReceiveKeyboardP2P(msg.ICEServers, msg.Device)
		}
	case "ice_server":
		if global.VideoConn != nil {
			global.VideoConn.Send <- map[string]interface{}{
				"operation":  "ice_server",
				"iceservers": msg.ICEServers,
				"device":     msg.Device,
			}
			//初始化键盘
			utils.InitSenderKeyboardP2P(msg.ICEServers, msg.Device)
		}
	case "offer":
		if global.VideoConn != nil {
			global.VideoConn.Send <- map[string]interface{}{
				"operation": "offer",
				"data":      msg.Data,
				"device":    msg.Device,
			}
		}
	case "answer":
		if global.ClientConn != nil {
			global.ClientConn.Send <- map[string]interface{}{
				"operation": "answer",
				"data":      msg.Data,
				"device":    msg.Device,
			}
		}
	case "ice_candidate":
		if !msg.VideoSender {
			if global.VideoConn != nil {
				global.VideoConn.Send <- map[string]interface{}{
					"operation": "ice_candidate",
					"data":      msg.Data,
					"device":    msg.Device,
				}
			}
		} else {
			if global.ClientConn != nil {
				global.ClientConn.Send <- map[string]interface{}{
					"operation": "ice_candidate",
					"data":      msg.Data,
					"device":    msg.Device,
				}
			}

		}

	}
}
