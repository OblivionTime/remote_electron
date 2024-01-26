package server

import (
	"remote/api/client"
	"remote/global"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
)

func HandlerVideo(msg []byte, conn *swebsocket.ServerConn) {
	conn.Send <- msg
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
	global.VideoConn, _ = swebsocket.CreateConn(wsConn, 2)
	global.Remote_serverConn.Send <- client.HandlerResult{
		Op:         "join",
		Device:     DeviceID,
		Code:       code,
		SendDevice: global.DeviceInfo.IdentificationCode,
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
