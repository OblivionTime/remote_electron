package api

import (
	"encoding/json"
	"fmt"
	"remote_server/config"
	"remote_server/global"
	"remote_server/model"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"
)

type ICEServer struct {
	URL        string `json:"url"`
	Credential string `json:"credential"`
	Username   string `json:"username"`
}
type HandlerResult struct {
	Op          string      `json:"op"` //操作
	Device      string      `json:"device,omitempty"`
	Code        string      `json:"code,omitempty"`
	SendDevice  string      `json:"send_device,omitempty"`
	Data        interface{} `json:"data,omitempty"`
	VideoSender bool        `json:"videoSender,omitempty"`
	ICEServers  []ICEServer `json:"iceservers,omitempty"`
	//键盘相关参数
	KeyboardOp   string                     `json:"keyboard_op,omitempty"`
	KeyboardData []byte                     `json:"keyboard_data,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`
}

func HandlerData(res []byte, conn *swebsocket.ServerConn) {
	var msg HandlerResult
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join":
		if _, ok := global.DeviceList[msg.SendDevice]; ok {
			if flag := SendDisconnected(msg.Device, msg.SendDevice); !flag {
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
			global.DeviceList[msg.SendDevice].Send <- HandlerResult{
				Op:         "ice_server",
				ICEServers: iceClient,
				Device:     msg.Device,
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
		if flag := SendDisconnected(msg.Device, msg.SendDevice); !flag {
			break
		}
		global.DeviceList[msg.Device].Send <- msg
	default:
		if flag := SendDisconnected(msg.Device, msg.SendDevice); !flag {
			break
		}
		if msg.Op == "keyboard" && msg.KeyboardOp == "candidate" {
			global.DeviceList[msg.Device].Send <- HandlerResult{
				Op:           msg.Op,
				KeyboardOp:   msg.KeyboardOp,
				Device:       msg.Device,
				ICECandidate: msg.ICECandidate,
			}
		} else {
			global.DeviceList[msg.Device].Send <- msg
		}
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

func SendDisconnected(device, sendDevice string) bool {
	if _, ok := global.DeviceList[device]; !ok {
		if _, ok2 := global.DeviceList[sendDevice]; ok2 {
			global.DeviceList[sendDevice].Send <- HandlerResult{
				Op:     "disconnected",
				Device: device,
			}
		}
		return false
	}
	return true
}

type Device struct {
	DeviceID string                 `json:"device_id"`
	Conn     *swebsocket.ServerConn `json:"connection"`
}

var DeviceList = make(map[string][]Device, 0)

// 远程服务器发送的数据
type FileRemoteReceive struct {
	Op     string `json:"op,omitempty"`
	Device string `json:"device,omitempty"`
	Code   string `json:"code,omitempty"`
}

func FileConnect(ctx *gin.Context) {
	deviceId := ctx.Query("device_id")
	room := ctx.Query("room")
	if deviceId == "" || room == "" {
		return
	}
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	ClientConn, _ := swebsocket.CreateConn(wsConn, 1)
	if _, ok := DeviceList[room]; !ok {
		DeviceList[room] = make([]Device, 0)
	}
	if len(DeviceList[room]) == 2 {
		DeviceList[room][1] = Device{
			DeviceID: deviceId,
			Conn:     ClientConn,
		}
	} else {
		DeviceList[room] = append(DeviceList[room], Device{
			DeviceID: deviceId,
			Conn:     ClientConn,
		})
	}

	ClientConn.Handle(func(res []byte, conn *swebsocket.ServerConn) {
		var msg FileRemoteReceive
		json.Unmarshal(res, &msg)
		if msg.Op == "join" {
			if _, ok := global.DeviceList[msg.Device]; !ok {
				conn.Send <- map[string]string{
					"op":     "disconnected",
					"errmsg": "对方不在线",
				}
			}

			//判断验证码是否正确
			var deviceInfo model.Device
			global.DB.Model(model.Device{}).Where("identificationCode = ?", msg.Device).First(&deviceInfo)
			if deviceInfo.VerificationCode != msg.Code {
				conn.Send <- map[string]string{
					"op":     "disconnected",
					"errmsg": "验证码不正确",
				}
				return
			}
			Username, Credential := global.Turnserver.Credentials(fmt.Sprintf("%s_remote_%v", room, time.Now().UnixNano()))
			iceClient := []ICEServer{{
				URL:        fmt.Sprintf("turn:%v:%d?transport=udp", config.Config.Turn.PublicIP, config.Config.Turn.Port),
				Credential: Credential,
				Username:   Username,
			}, {
				URL:        fmt.Sprintf("turn:%v:%d?transport=tcp", config.Config.Turn.PublicIP, config.Config.Turn.Port),
				Credential: Credential,
				Username:   Username,
			}}
			conn.Send <- map[string]interface{}{
				"op":         "ice_server",
				"iceservers": iceClient,
			}
			if _, ok := global.DeviceList[msg.Device]; ok {
				global.DeviceList[msg.Device].Send <- HandlerResult{
					Op:         "file_join",
					ICEServers: iceClient,
					SendDevice: deviceId,
				}
			}

			return
		}
		for _, device := range DeviceList[room] {
			if device.DeviceID == deviceId {
				continue
			}
			device.Conn.Send <- res
		}
	})
	ClientConn.WriteReadLoop()
	for _, device := range DeviceList[room] {
		if device.DeviceID == deviceId {
			continue
		}
		device.Conn.CloseConn()
	}
	delete(DeviceList, room)
	global.Turnserver.Disallow(room + "_fileshare")
	fmt.Println("清空所有", DeviceList[room])
}
