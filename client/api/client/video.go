package client

import (
	"remote/global"
	"remote/utils"

	"gitee.com/solidone/sutils/swebsocket"
)

var ConnectDevice = make(map[interface{}]bool)

func VideoHandler(device string, conn *swebsocket.ServerConn) {
	var flag bool
	ConnectDevice[device] = false
	for {
		flag = ConnectDevice[device]
		if flag {
			return
		}
		data := utils.GetScreenshots()
		if data != nil {
			conn.Send <- HandlerResult{
				Op:         "video",
				Device:     device,
				SendDevice: global.DeviceInfo.IdentificationCode,
				VideoData:  data,
			}
		}
	}

}
