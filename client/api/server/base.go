package server

import (
	"fmt"
	"os"
	"remote/api/client"
	"remote/global"
	"remote/model"
	"remote/model/common/response"
	"remote/utils"
	"strings"
	"time"

	"gitee.com/solidone/sutils/shttp"
	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type DeviceResponse struct {
	Code int    `json:"code"`
	Data []byte `json:"data"`
	Msg  string `json:"msg"`
}
type Remote struct {
	RemoteURL string `json:"remote_url"`
	HTTPS     bool   `json:"https"`
}
type RemoteServerResponse struct {
	Operation string `json:"operation"`
}

// 建立客户端的websocket
func Connect(ctx *gin.Context) {
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	if global.ClientConn != nil {
		global.ClientConn.CloseConn()
		global.ClientConn = nil
	}
	global.ClientConn, _ = swebsocket.CreateConn(wsConn, 1)
	global.ClientConn.WriteReadLoop()
	global.ClientConn = nil
}

// 连接服务器
func ConnectServer(ctx *gin.Context) {
	global.RemoteServerIP = ""
	var msg Remote
	if err := ctx.ShouldBindJSON(&msg); err != nil {
		response.FailWithMessage("参数错误", ctx)
		return
	}
	device, err := os.Hostname()
	if err != nil {
		response.FailWithMessage("无法获取设备信息", ctx)
		return
	}
	var protocol string
	if msg.HTTPS {
		protocol = "https:"
	} else {
		protocol = "http:"

	}
	remoteURL := fmt.Sprintf("%s//%s/v1/api/remote/code?device=%s", protocol, msg.RemoteURL, device)
	client := shttp.NewClient()
	err = client.GET(remoteURL, nil)
	if err != nil {
		response.FailWithMessage("服务器连接失败", ctx)
		return
	}
	var res DeviceResponse
	err = client.GetResponseData(&res)
	if err != nil {
		response.FailWithMessage("服务器连接失败", ctx)
		return
	}
	if res.Code != 0 {
		response.FailWithMessage(res.Msg, ctx)
		return
	}

	utils.Decrypt(res.Data, &global.DeviceInfo)
	global.RemoteServerIP = fmt.Sprintf("%s//%s", protocol, msg.RemoteURL)
	go connectServer()
	response.OkWithData(global.DeviceInfo, ctx)
}

type Device struct {
	DeviceID           string `json:"device_id,omitempty" `
	IdentificationCode string `json:"identificationCode" `
	VerificationCode   string `json:"verificationCode"`
	Connectioned       string `json:"connectioned,omitempty" ` //连接过的设备
}

// 远程控制
func CheckDeviceOnline(ctx *gin.Context) {
	if global.DeviceInfo.DeviceID == "" {
		response.FailWithMessage("请先与服务器建立连接后在发起请求", ctx)
		return
	}
	var msg Device
	if err := ctx.BindJSON(&msg); err != nil {
		response.Fail(ctx)
		return
	}
	remoteURL := fmt.Sprintf("%s/v1/api/remote/online?code=%s", global.RemoteServerIP, global.DeviceInfo.IdentificationCode)
	fmt.Println(remoteURL)
	client := shttp.NewClient()
	err := client.POST(remoteURL, msg)
	if err != nil {
		fmt.Println(err)
		response.FailWithMessage("服务器连接失败", ctx)
		return
	}
	var res response.Response
	err = client.GetResponseData(&res)
	if err != nil {
		fmt.Println(err)
		response.FailWithMessage("服务器连接失败", ctx)
		return
	}
	if res.Code != 0 {
		response.FailWithMessage(res.Msg, ctx)
		return
	}
	response.Ok(ctx)
}

// 连接服务器
func connectServer() {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	remoteURL := strings.Replace(global.RemoteServerIP, "http://", "ws://", -1)
	remoteURL = strings.Replace(remoteURL, "https://", "wss://", -1)
	remote_serverCn, _, err := dialer.Dial(remoteURL+"/v1/api/remote/connect?code="+global.DeviceInfo.IdentificationCode, nil)
	if err != nil {
		if global.ClientConn != nil {
			global.ClientConn.Send <- RemoteServerResponse{Operation: "disconnect"}
		}
		return
	}
	global.Remote_serverConn, err = swebsocket.CreateConn(remote_serverCn, 1)
	if err != nil {
		if global.ClientConn != nil {
			global.ClientConn.Send <- RemoteServerResponse{Operation: "disconnect"}
		}
		return
	}
	global.Remote_serverConn.Handle(client.RemoteDataHandler)
	global.Remote_serverConn.WriteReadLoop()
	if global.ClientConn != nil {
		global.ClientConn.Send <- RemoteServerResponse{Operation: "disconnect"}
	}
	global.DeviceInfo = &model.Device{}
}
