package server

import (
	"encoding/json"
	"fmt"
	"remote/ckeyboard"
	"remote/global"
	"remote/input"
	"remote/model"
	"remote/p2p"
	"remote/utils"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
)

// Request 是 Electron UI 通过 /keyboard_connect 发来的指令
type Request struct {
	Operation string               `json:"operation"`
	RemoteURL string               `json:"remote_url,omitempty"`
	Status    bool                 `json:"status,omitempty"`
	Mode      string               `json:"mode,omitempty"`
	Data      model.InputOperation `json:"data,omitempty"`
}

// ConnectKeyboard 建立键盘通道：本机物理按键经钩子捕获后，
// 与 UI 发来的键鼠指令一起走数据通道发往被控端。
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
	global.AddKeyboardSession(user)
	downCh, upCh, _ := global.KeyboardChannels(user)

	go func() {
		defer func() {
			if global.KeyboardConn != nil {
				global.KeyboardConn.CloseConn()
				global.KeyboardConn = nil
			}
			utils.RecoverPanic()
		}()
		var data model.InputOperation
		for {
			if global.KeyboardConn == nil || global.Remote_serverConn == nil || global.KeyboardHandler == nil {
				return
			}
			select {
			case key := <-downCh:
				data = model.InputOperation{Method: "KeyDown", Key: key}
			case key := <-upCh:
				data = model.InputOperation{Method: "KeyRelease", Key: key}
			}
			if global.Remote_serverConn != nil && global.KeyboardHandler != nil {
				p2p.SendJSON(global.KeyboardHandler, data)
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
			if global.Remote_serverConn != nil && global.KeyboardHandler != nil {
				p2p.SendJSON(global.KeyboardHandler, msg.Data)
			}
		}
	})

	global.KeyboardConn.WriteReadLoop()
	releasePressedKeys()
	global.KeyboardConn = nil
}

// releasePressedKeys 会话结束时松开可能仍被按住的修饰键
func releasePressedKeys() {
	global.OpenWatch = false
	for _, k := range []string{"lalt", "ralt", "lctrl", "rctrl"} {
		input.KeyUp(k)
	}
}
