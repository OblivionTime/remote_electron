package client

import (
	"encoding/json"
	"fmt"
	"remote/file"
	"remote/global"
	"remote/model"
	"remote/p2p"
	"strings"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

// HandlerFileShare 被控端收到控制端的文件共享请求：连上中转服务器，
// 等控制端建好数据通道后接管文件收发。
func HandlerFileShare(msg model.Signal) {
	fmt.Println("接收到远程文件共享")
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 30 * time.Second
	remoteURL := strings.Replace(global.RemoteServerIP, "http://", "ws://", -1)
	remoteURL = strings.Replace(remoteURL, "https://", "wss://", -1)
	remoteConn, _, err := dialer.Dial(remoteURL+"/v1/api/remote/file_connect?device_id="+global.DeviceInfo.IdentificationCode+"&room="+msg.SendDevice, nil)
	if err != nil {
		return
	}
	FileRemoteConn, err := swebsocket.CreateConn(remoteConn, 1)
	if err != nil {
		return
	}

	var dataChannel *webrtc.DataChannel
	webrtcConn, err := p2p.NewFileReceiver(msg.ICEServers, func(d *webrtc.DataChannel) {
		dataChannel = d
		d.OnOpen(func() {
			fmt.Println("被控端开启成功", d.Label(), d.ID())
			// 队列消费者进程内只起一个，这里只把通道换成当前会话的
			file.SetSender(d, FileRemoteConn)
			file.StartWorker()
		})
		d.OnMessage(func(dataMsg webrtc.DataChannelMessage) {
			var res model.ClientFileReceive
			json.Unmarshal(dataMsg.Data, &res)
			// 续传 / 校验握手回执，不能当普通业务消息处理
			if file.IsAck(res.FileOp) {
				file.Ack(res)
				return
			}
			switch res.Op {
			case "remote_ls":
				p2p.SendJSON(d, model.ClientFileResponse{
					Op:             res.Op,
					RemoteFileList: file.GetAllFiles(res.RemotePath),
				})
			case "remote_basic":
				p2p.SendJSON(d, model.ClientFileResponse{
					Op:              res.Op,
					RemoteBasicList: file.GetBasicDir(),
				})
			case "upload":
				file.SenderHandlerFile(res, d, FileRemoteConn)
			case "download":
				HandlerDownload(res, d)
			case "pause", "cancel":
				// 下载的发送端在本机，收到控制端的暂停/取消就停
				file.Stop(res.RemotePath)
			case "fs_mkdir", "fs_delete":
				// 控制端代它操作本机文件
				if err := file.Apply(res); err != nil {
					p2p.SendJSON(d, model.ClientFileResponse{Op: "error", ErrMsg: err.Error()})
					return
				}
				p2p.SendJSON(d, model.ClientFileResponse{
					Op:             "remote_ls",
					RemoteFileList: file.GetAllFiles(res.RemotePath),
				})
			}
		})
		d.OnClose(func() {
			file.ClearSender(d)
			dataChannel = nil
		})
	})
	if err != nil {
		fmt.Println("初始化文件 P2P 失败", err)
		return
	}

	p2p.WatchState(webrtcConn, "file", func() {
		if dataChannel != nil {
			dataChannel.Close()
		}
	})
	p2p.OnCandidate(webrtcConn, func(c webrtc.ICECandidateInit) {
		FileRemoteConn.Send <- model.FileSignal{Op: "candidate", ICECandidate: c}
	})

	FileRemoteConn.Send <- model.FileSignal{Op: "new_peer"}
	FileRemoteConn.Handle(func(remoteMsg []byte, conn *swebsocket.ServerConn) {
		var remoteRes model.FileSignal
		json.Unmarshal(remoteMsg, &remoteRes)
		switch remoteRes.Op {
		case "disconnect":
			FileRemoteConn.CloseConn()
		case "offer":
			if webrtcConn != nil {
				answer, err := p2p.NewAnswer(webrtcConn, remoteRes.SDP)
				if err != nil {
					fmt.Println("创建文件 answer 失败", err)
					return
				}
				FileRemoteConn.Send <- model.FileSignal{Op: "answer", SDP: answer}
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
	if webrtcConn != nil {
		webrtcConn.Close()
	}
	if dataChannel != nil {
		file.ClearSender(dataChannel)
		dataChannel.Close()
	}
	// FileList 与消费者是进程级常驻的，这里不能关掉/重建：
	// 换掉队列会让正在处理旧队列的消费者永久丢任务，新任务没人接。
}

func HandlerDownload(msg model.ClientFileReceive, dataChannel *webrtc.DataChannel) {
	fmt.Println("接收到数据", msg.Op, msg.FilePath, msg.RemotePath, msg.FileSize)
	if dataChannel != nil {
		// 重试 / 继续时清掉上次的暂停标记
		file.ClearStop(msg.RemotePath)
		file.FileList <- msg
	}
}
