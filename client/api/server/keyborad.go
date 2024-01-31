package server

import (
	"encoding/json"
	"fmt"
	"remote/ckeyboard"
	"remote/global"
	"remote/utils"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
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

// 建立键盘的websocket
func ConnectKeyboard(ctx *gin.Context) {
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
			if global.KeyboardConn == nil || global.Remote_serverConn == nil {
				return
			} else if (global.Remote_serverConn != nil && global.Remote_serverConn.Flag) || (global.KeyboardConn != nil && global.KeyboardConn.Flag) {
				return
			} else if global.KeyboardHandler == nil {
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
			if global.Remote_serverConn != nil && global.KeyboardHandler != nil {
				global.KeyboardHandler.Send(succesInfo)
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
				fmt.Println("关闭监听模式")
				global.OpenWatch = false
			}
		case "send":
			succesInfo, _ := json.Marshal(msg.Data)
			if global.Remote_serverConn != nil && global.KeyboardHandler != nil {
				global.KeyboardHandler.Send(succesInfo)
			}
		}
	})

	global.KeyboardConn.WriteReadLoop()
	global.KeyboardConn = nil
}
