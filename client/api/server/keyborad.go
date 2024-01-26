package server

import (
	"encoding/json"
	"fmt"
	"remote/api/client"
	"remote/ckeyboard"
	"remote/global"
	"remote/model/common/response"
	"remote/utils"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Request struct {
	Operation string   `json:"operation"`
	RemoteURL string   `json:"remote_url,omitempty"`
	Status    bool     `json:"status,omitempty"`
	Mode      string   `json:"mode,omitempty"`
	Data      SendData `json:"data,omitempty"`
}

type SendData struct {
	Method string  `json:"method"`
	Arg1   int     `json:"arg1,omitempty"`
	Arg2   int     `json:"arg2,omitempty"`
	Key    string  `json:"key,omitempty"`
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`
}

func HandlerKeyboard(msg []byte, conn *swebsocket.ServerConn) {

}

// 建立键盘的websocket
func ConnectKeyboard(ctx *gin.Context) {
	DeviceID := ctx.Query("device")
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	if global.KeyboardConn != nil {
		global.KeyboardConn.CloseConn()
		global.KeyboardConn = nil
	}
	global.KeyboardConn, _ = swebsocket.CreateConn(wsConn, 1)
	user := utils.GenerateRandomString()
	global.KeyDownList[user] = make(chan string, 1024)
	global.KeyRelseList[user] = make(chan string, 1024)
	go func() {
		defer func() {
			if global.KeyboardConn != nil {
				global.KeyboardConn.CloseConn()
				global.KeyboardConn = nil
			}
			utils.RecoverPanic()
		}()
		var data SendData
		for {
			if (global.Remote_serverConn != nil && global.Remote_serverConn.Flag) || global.KeyboardConn.Flag {
				return
			}
			select {
			case key := <-global.KeyDownList[user]:
				data = SendData{
					Method: "KeyDown",
					Key:    key,
				}

			case key := <-global.KeyRelseList[user]:
				data = SendData{
					Method: "KeyRelease",
					Key:    key,
				}
			}
			succesInfo, _ := json.Marshal(data)
			if global.Remote_serverConn != nil {
				global.Remote_serverConn.Send <- client.HandlerResult{
					Op:           "keyboard",
					Device:       DeviceID,
					SendDevice:   global.DeviceInfo.IdentificationCode,
					KeyboardData: succesInfo,
				}
			}
		}
	}()
	ckeyboard.ChangeUser(user)
	go ckeyboard.ListenKeyboard()
	global.KeyboardConn.Handle(func(res []byte, conn *swebsocket.ServerConn) {
		var msg Request
		json.Unmarshal(res, &msg)
		switch msg.Operation {
		case "change":
			if msg.Status {
				global.OpenWatch = true
				ckeyboard.ChangeUser(user)
			} else {
				global.OpenWatch = false
			}
		case "send":
			if global.Remote_serverConn == nil {
				errinfo, _ := json.Marshal(response.ErrRequestClient("远程服务器连接失败!!"))
				conn.Send <- errinfo
				break
			}
			succesInfo, _ := json.Marshal(msg.Data)
			if global.Remote_serverConn != nil {
				global.Remote_serverConn.Send <- client.HandlerResult{
					Op:           "keyboard",
					Device:       DeviceID,
					SendDevice:   global.DeviceInfo.IdentificationCode,
					KeyboardData: succesInfo,
				}
			}
		}
	})

	global.KeyboardConn.WriteReadLoop()
	global.KeyboardConn = nil
}
func Bluetooth(ctx *gin.Context) {
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	conn, _ := swebsocket.CreateConn(wsConn, 1)
	user := utils.GenerateRandomString()
	var remote_serverConn *swebsocket.ServerConn
	global.KeyDownList[user] = make(chan string, 1024)
	global.KeyRelseList[user] = make(chan string, 1024)
	go func() {
		defer func() {
			if conn != nil {
				conn.CloseConn()

			}
			if remote_serverConn != nil {
				remote_serverConn.CloseConn()
			}
			utils.RecoverPanic()
		}()
		for {
			if (remote_serverConn != nil && remote_serverConn.Flag) || conn.Flag {
				return
			}
			select {
			case key := <-global.KeyDownList[user]:
				data := SendData{
					Method: "KeyDown",
					Key:    key,
				}
				succesInfo, _ := json.Marshal(data)
				if remote_serverConn != nil {
					remote_serverConn.Send <- succesInfo
				}
			case key := <-global.KeyRelseList[user]:
				data := SendData{
					Method: "KeyRelease",
					Key:    key,
				}
				succesInfo, _ := json.Marshal(data)
				if remote_serverConn != nil {
					remote_serverConn.Send <- succesInfo
				}

			}
		}
	}()

	conn.Handle(func(res []byte, conn *swebsocket.ServerConn) {
		var msg Request
		json.Unmarshal(res, &msg)
		switch msg.Operation {
		case "connect":
			dialer := websocket.DefaultDialer
			dialer.HandshakeTimeout = 30 * time.Second
			remote_serverCn, _, err := dialer.Dial(msg.RemoteURL+"/v1/api/remote/client/keyboard_ws", nil)
			if err != nil {
				errinfo, _ := json.Marshal(response.ErrRequestClient("远程服务器连接失败!!"))
				conn.Send <- errinfo
				break
			}
			remote_serverConn, _ = swebsocket.CreateConn(remote_serverCn, 1)
			fmt.Println("键盘连接成功")
			go remote_serverConn.WriteReadLoop()
			ckeyboard.ChangeUser(user)
			go ckeyboard.ListenKeyboard()

		case "change":
			if msg.Status {
				global.OpenWatch = true
				ckeyboard.ChangeUser(user)
			} else {
				global.OpenWatch = false
			}
		case "send":
			if remote_serverConn == nil {
				errinfo, _ := json.Marshal(response.ErrRequestClient("远程服务器连接失败!!"))
				conn.Send <- errinfo
				break
			}
			succesInfo, _ := json.Marshal(msg.Data)
			if remote_serverConn != nil {
				remote_serverConn.Send <- succesInfo
			}
		}
	})
	conn.WriteReadLoop()

}
