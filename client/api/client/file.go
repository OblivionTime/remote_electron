package client

import (
	"encoding/json"
	"fmt"
	"remote/global"
	"remote/model"
	"remote/utils"
	"strings"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type FileRemoteReceive struct {
	Op           string                     `json:"op,omitempty"`
	ICEServers   []model.ICEServer          `json:"iceservers,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`
	Device       string                     `json:"device,omitempty"`
	Code         string                     `json:"code,omitempty"`
	ErrMsg       string                     `json:"errmsg,omitempty"`
}

func HandlerFileShare(msg HandlerResult) {
	fmt.Println("接收到远程文件共享")
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	remoteURL := strings.Replace(global.RemoteServerIP, "http://", "ws://", -1)
	remoteURL = strings.Replace(remoteURL, "https://", "wss://", -1)
	remote_serverCn, _, err := dialer.Dial(remoteURL+"/v1/api/remote/file_connect?device_id="+global.DeviceInfo.IdentificationCode+"&room="+msg.SendDevice, nil)
	if err != nil {
		return
	}
	FileRemoteConn, err := swebsocket.CreateConn(remote_serverCn, 1)
	if err != nil {
		return
	}
	var dataChannel *webrtc.DataChannel
	webrtcConn := utils.InitReceiveFileP2P(msg.ICEServers, func(d *webrtc.DataChannel) {
		dataChannel = d
		dataChannel.OnOpen(func() {
			fmt.Println("被控端开启成功", dataChannel.Label(), dataChannel.ID())
			go utils.HandlerFile(FileRemoteConn, dataChannel)
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
			switch msg.Op {
			case "remote_ls":
				utils.SendMessage(model.ClientFileResponse{
					Op:             msg.Op,
					RemoteFileList: utils.GetAllFiles(msg.RemotePath),
				}, dataChannel)
			case "remote_basic":
				utils.SendMessage(model.ClientFileResponse{
					Op:              msg.Op,
					RemoteBasicList: utils.GetBasicDir(),
				}, dataChannel)
			case "upload":
				utils.SenderHandlerFile(msg, dataChannel, FileRemoteConn)
			case "download":
				HandlerDownload(msg, dataChannel)
			}
		})
		dataChannel.OnClose(func() {
			dataChannel = nil
		})
	})
	webrtcConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s >= 4 {
			fmt.Printf("Peer Connection State has changed: %s\n", s.String())
			if dataChannel != nil {
				fmt.Printf("即将断开'%s'-'%d' \n", dataChannel.Label(), dataChannel.ID())
			}
			fmt.Println("连接断开", s.String())
			if dataChannel != nil {
				dataChannel.Close()
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
	FileRemoteConn.Send <- map[string]string{
		"op": "new_peer",
	}
	FileRemoteConn.Handle(func(remoteMsg []byte, conn *swebsocket.ServerConn) {
		var remoteRes FileRemoteReceive
		json.Unmarshal(remoteMsg, &remoteRes)
		switch remoteRes.Op {
		case "disconnect":
			FileRemoteConn.CloseConn()
		case "offer":
			if webrtcConn != nil {
				webrtcConn.SetRemoteDescription(*remoteRes.SDP)
				answerData, _ := webrtcConn.CreateAnswer(nil)
				webrtcConn.SetLocalDescription(answerData)
				FileRemoteConn.Send <- FileRemoteReceive{
					Op:  "answer",
					SDP: &answerData,
				}
			}
		case "candidate":
			if webrtcConn != nil {
				webrtcConn.AddICECandidate(remoteRes.ICECandidate)
			}
		}
	})
	FileRemoteConn.WriteReadLoop()
	fmt.Println("被控端连接结束")
	FileRemoteConn = nil
	if FileRemoteConn != nil {
		FileRemoteConn.CloseConn()
	}
	if webrtcConn != nil {
		webrtcConn = nil
	}
	if dataChannel != nil {
		dataChannel.Close()
	}
	if utils.FileList != nil {
		close(utils.FileList)
		utils.FileList = nil
	}
	utils.FileList = make(chan model.ClientFileReceive, 1024)
}

func HandlerDownload(msg model.ClientFileReceive, dataChannel *webrtc.DataChannel) {
	fmt.Println("接收到数据", msg.Op, msg.FilePath, msg.RemotePath, msg.FileSize)
	if dataChannel != nil {
		utils.FileList <- msg
	}
}
