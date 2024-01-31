package api

import (
	"encoding/json"
	"fmt"
	"remote_server/config"
	"remote_server/global"
	"remote_server/model"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
)

type ICEServer struct {
	URL        string `json:"url"`
	Credential string `json:"credential"`
	Username   string `json:"username"`
}
type HandlerResult struct {
	Op           string      `json:"op"` //操作
	Device       string      `json:"device,omitempty"`
	Code         string      `json:"code,omitempty"`
	SendDevice   string      `json:"send_device,omitempty"`
	KeyboardData []byte      `json:"keyboard_data,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	VideoSender  bool        `json:"videoSender,omitempty"`
	ICEServers   []ICEServer `json:"iceservers,omitempty"`
}

func HandlerData(res []byte, conn *swebsocket.ServerConn) {
	var msg HandlerResult
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join":
		if _, ok := global.DeviceList[msg.SendDevice]; ok {
			if _, ok2 := global.DeviceList[msg.Device]; !ok2 {
				global.DeviceList[msg.SendDevice].Send <- HandlerResult{
					Op:     "disconnected",
					Device: msg.Device,
				}
				break
			}
			//判断验证码是否正确
			var deviceInfo model.Device
			global.DB.Model(model.Device{}).Where("identificationCode = ?", msg.Device).First(&deviceInfo)
			if deviceInfo.VerificationCode != msg.Code {
				global.DeviceList[msg.SendDevice].Send <- HandlerResult{
					Op:     "disconnected",
					Device: msg.Device,
				}
				break
			}
			global.VideoRooms[msg.Device] = append(global.VideoRooms[msg.Device], msg.SendDevice)
			Username, Credential := global.Turnserver.Credentials(msg.SendDevice)
			iceClient := []ICEServer{{
				URL:        fmt.Sprintf("turn:%v:%d?transport=udp", config.Config.Turn.PublicIP, config.Config.Turn.Port),
				Credential: Credential,
				Username:   Username,
			}, {
				URL:        fmt.Sprintf("turn:%v:%d?transport=tcp", config.Config.Turn.PublicIP, config.Config.Turn.Port),
				Credential: Credential,
				Username:   Username,
			}}
			fmt.Printf("iceClient ,%+v\n", iceClient)
			global.DeviceList[msg.SendDevice].Send <- HandlerResult{
				Op:         "ice_server",
				ICEServers: iceClient,
			}
			global.DeviceList[msg.Device].Send <- HandlerResult{
				Op:         "join",
				Device:     msg.SendDevice,
				ICEServers: iceClient,
			}

		}
	case "disconnected":
		global.Turnserver.Disallow(msg.SendDevice)
		global.Turnserver.Disallow(msg.Device)
		if _, ok := global.DeviceList[msg.Device]; !ok {
			global.DeviceList[msg.SendDevice].Send <- HandlerResult{
				Op:     "disconnected",
				Device: msg.Device,
			}
			break
		}
		global.DeviceList[msg.Device].Send <- msg
	default:
		if _, ok := global.DeviceList[msg.Device]; !ok {
			global.DeviceList[msg.SendDevice].Send <- HandlerResult{
				Op:     "disconnected",
				Device: msg.Device,
			}
			break
		}
		global.DeviceList[msg.Device].Send <- msg
	}
}
func Connect(ctx *gin.Context) {
	DeviceID := ctx.Query("code")
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	conn, _ := swebsocket.CreateConn(wsConn, 2)
	fmt.Println(DeviceID, "连接到服务器")
	global.DeviceList[DeviceID] = conn
	conn.Handle(HandlerData)
	conn.WriteReadLoop()
	fmt.Println(DeviceID, "已退出")
	delete(global.DeviceList, DeviceID)
}
