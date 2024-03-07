package server

import (
	"encoding/json"
	"fmt"
	"remote/global"
	"remote/model"
	"remote/model/common/response"
	"remote/utils"
	"strings"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

// 远程服务器发送的数据
type FileRemoteReceive struct {
	Op           string                     `json:"op,omitempty"`
	ICEServers   []model.ICEServer          `json:"iceservers,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`
	Device       string                     `json:"device,omitempty"`
	Code         string                     `json:"code,omitempty"`
	ErrMsg       string                     `json:"errmsg,omitempty"`
}

const BUFFERED_AMOUNT_LOW_THRESHOLD uint64 = 256 * 1024

func FileConnect(ctx *gin.Context) {
	wsConn, err := global.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	FileConn, _ := swebsocket.CreateConn(wsConn, 1)
	var FileRemoteConn *swebsocket.ServerConn
	var dataChannel *webrtc.DataChannel
	var webrtcConn *webrtc.PeerConnection

	FileConn.Handle(func(msg []byte, conn *swebsocket.ServerConn) {
		var res model.ClientFileReceive
		json.Unmarshal(msg, &res)
		switch res.Op {
		case "connection":
			dialer := websocket.DefaultDialer
			dialer.HandshakeTimeout = 30 * time.Second
			remoteURL := strings.Replace(global.RemoteServerIP, "http://", "ws://", -1)
			remoteURL = strings.Replace(remoteURL, "https://", "wss://", -1)
			remote_serverCn, _, err := dialer.Dial(remoteURL+"/v1/api/remote/file_connect?device_id="+global.DeviceInfo.IdentificationCode+"&room="+global.DeviceInfo.IdentificationCode, nil)
			if err != nil {
				FileConn.Send <- response.ErrRequestUSBExit("连接远程服务器失败")
				return
			}
			FileRemoteConn, err = swebsocket.CreateConn(remote_serverCn, 1)
			if err != nil {
				FileConn.Send <- response.ErrRequestUSBExit("连接远程服务器失败")
				return
			}
			FileRemoteConn.Send <- FileRemoteReceive{
				Op:     "join",
				Device: res.Device,
				Code:   res.Code,
			}
			FileRemoteConn.Handle(func(remoteMsg []byte, conn *swebsocket.ServerConn) {
				var remoteRes FileRemoteReceive
				json.Unmarshal(remoteMsg, &remoteRes)
				switch remoteRes.Op {
				case "disconnected":
					FileConn.Send <- model.ClientFileResponse{
						Op:     "disconnected",
						ErrMsg: remoteRes.ErrMsg,
					}
					FileRemoteConn.CloseConn()
				case "ice_server":
					webrtcConn, dataChannel = utils.InitSendFileP2P(remoteRes.ICEServers)
					webrtcConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
						global.CandidatesMux.Lock()
						defer global.CandidatesMux.Unlock()
						if s >= 4 {
							fmt.Printf("Peer Connection State has changed: %s\n", s.String())
							fmt.Println("连接断开", s.String())
							if dataChannel != nil {
								fmt.Printf("即将断开'%s'-'%d' \n", dataChannel.Label(), dataChannel.ID())
							}

						}
					})
					webrtcConn.OnICECandidate(func(c *webrtc.ICECandidate) {
						if c == nil {
							return
						}
						global.CandidatesMux.Lock()
						defer global.CandidatesMux.Unlock()
						FileRemoteConn.Send <- FileRemoteReceive{
							Op:           "candidate",
							SDP:          nil,
							ICECandidate: c.ToJSON(),
						}
					})
					dataChannel.OnOpen(func() {
						fmt.Println("控制端开启成功", dataChannel.Label(), dataChannel.ID())

						FileConn.Send <- model.ClientFileResponse{
							Op: "connection",
						}
					})
					dataChannel.OnBufferedAmountLow(func() {
						select {
						case data := <-utils.DataToBeSent:
							dataChannel.Send(data)
						default:
							fmt.Println("No data available to send")
						}
					})
					dataChannel.OnMessage(func(dataMsg webrtc.DataChannelMessage) {
						var msg model.ClientFileReceive
						json.Unmarshal(dataMsg.Data, &msg)
						if msg.Op == "download" {
							fmt.Println("接收到数据", msg.Op)
							utils.SenderHandlerFile(msg, dataChannel, FileConn)

						} else {
							FileConn.Send <- dataMsg.Data
						}
					})
					dataChannel.OnClose(func() {
						dataChannel = nil
					})
					dataChannel.SetBufferedAmountLowThreshold(BUFFERED_AMOUNT_LOW_THRESHOLD)
					go utils.HandlerFile(FileConn, dataChannel)
				case "new_peer":
					//发送offer
					offerData, _ := webrtcConn.CreateOffer(nil)
					webrtcConn.SetLocalDescription(offerData)
					FileRemoteConn.Send <- FileRemoteReceive{
						Op:  "offer",
						SDP: &offerData,
					}
				case "answer":
					if webrtcConn != nil {
						webrtcConn.SetRemoteDescription(*remoteRes.SDP)
					}
				case "candidate":
					if webrtcConn != nil {
						if err := webrtcConn.AddICECandidate(remoteRes.ICECandidate); err != nil {
							fmt.Println("接收candidate错误", err)
						}
					}
				}
			})
			go FileRemoteConn.WriteReadLoop()
		case "local_ls":
			FileConn.Send <- model.ClientFileResponse{
				Op:            res.Op,
				LocalFileList: utils.GetAllFiles(res.LocalPath),
			}
		case "local_basic":
			FileConn.Send <- model.ClientFileResponse{
				Op:             res.Op,
				LocalBasicList: utils.GetBasicDir(),
			}
		case "remote_ls":
			if dataChannel != nil {
				dataChannel.Send(msg)
			}
		case "remote_basic":
			if dataChannel != nil {
				dataChannel.Send(msg)
			}
		case "upload":
			if dataChannel != nil {
				utils.FileList <- res
			}
		case "download":
			if dataChannel != nil {
				dataChannel.Send(msg)
			}
		}
	})

	FileConn.WriteReadLoop()
	FileConn = nil
	if FileRemoteConn != nil {
		FileRemoteConn.CloseConn()
	}
	if webrtcConn != nil {
		webrtcConn.Close()
		webrtcConn = nil
	}
	if dataChannel != nil {
		dataChannel.Close()
		dataChannel = nil
	}
	if utils.FileList != nil {
		close(utils.FileList)
		utils.FileList = nil
	}
	utils.FileList = make(chan model.ClientFileReceive, 1024)
	fmt.Println("FileConnect结束")
}
