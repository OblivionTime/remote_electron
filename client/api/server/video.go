package server

import (
	"encoding/json"
	"remote/api/client"
	"remote/global"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
)

func HandlerVideo(msg []byte, conn *swebsocket.ServerConn) {
	var res WebRtcResponse
	json.Unmarshal(msg, &res)
	if global.Remote_serverConn != nil {
		global.Remote_serverConn.Send <- client.HandlerResult{
			Op:          res.Name,
			Data:        res.Data,
			Device:      res.Device,
			SendDevice:  global.DeviceInfo.IdentificationCode,
			VideoSender: true,
		}
	}

}

// 建立音视频的websocket
func ConnectVideo(ctx *gin.Context) {
	DeviceID := ctx.Query("device")
	code := ctx.Query("code")
	if DeviceID == "" || code == "" {
		return
	}
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	if global.VideoConn != nil {
		global.VideoConn.CloseConn()
		global.VideoConn = nil
	}
	global.VideoConn, _ = swebsocket.CreateConn(wsConn, 1)
	global.Remote_serverConn.Send <- client.HandlerResult{
		Op:          "join",
		Device:      DeviceID,
		Code:        code,
		SendDevice:  global.DeviceInfo.IdentificationCode,
		VideoSender: true,
	}
	global.VideoConn.Handle(HandlerVideo)
	global.VideoConn.WriteReadLoop()
	global.Remote_serverConn.Send <- client.HandlerResult{
		Op:         "disconnected",
		Device:     DeviceID,
		SendDevice: global.DeviceInfo.IdentificationCode,
	}
	global.VideoConn = nil
}
