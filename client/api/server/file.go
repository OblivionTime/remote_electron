package server

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
	logger "github.com/OblivionTime/simple-logger"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

const BUFFERED_AMOUNT_LOW_THRESHOLD uint64 = 256 * 1024

// FileConnect 控制端的文件窗口接入点：与对端建 P2P 数据通道收发文件
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
			remoteConn, _, err := dialer.Dial(remoteURL+"/v1/api/remote/file_connect?device_id="+global.DeviceInfo.IdentificationCode+"&room="+global.DeviceInfo.IdentificationCode, nil)
			if err != nil {
				FileConn.Send <- model.ClientFileResponse{Op: "disconnected", ErrMsg: "连接远程服务器失败"}
				return
			}
			FileRemoteConn, err = swebsocket.CreateConn(remoteConn, 1)
			if err != nil {
				FileConn.Send <- model.ClientFileResponse{Op: "disconnected", ErrMsg: "连接远程服务器失败"}
				return
			}
			FileRemoteConn.Send <- model.FileSignal{
				Op:     "join",
				Device: res.Device,
				Code:   res.Code,
			}
			FileRemoteConn.Handle(func(remoteMsg []byte, conn *swebsocket.ServerConn) {
				var remoteRes model.FileSignal
				json.Unmarshal(remoteMsg, &remoteRes)
				switch remoteRes.Op {
				case "disconnected":
					FileConn.Send <- model.ClientFileResponse{
						Op:     "disconnected",
						ErrMsg: remoteRes.ErrMsg,
					}
					FileRemoteConn.CloseConn()
				case "ice_server":
					// 用局部变量捕获本次会话的通道，闭包里读全局 dataChannel 会拿到已置空的值
					var dc *webrtc.DataChannel
					webrtcConn, dc, err = p2p.NewFileSender(remoteRes.ICEServers, global.DeviceInfo.IdentificationCode)
					dataChannel = dc
					if err != nil {
						fmt.Println("初始化文件 P2P 失败", err)
						return
					}
					p2p.WatchState(webrtcConn, "file", func() {
						if dc != nil {
							dc.Close()
						}
					})
					p2p.OnCandidate(webrtcConn, func(c webrtc.ICECandidateInit) {
						FileRemoteConn.Send <- model.FileSignal{Op: "candidate", ICECandidate: c}
					})
					dc.OnOpen(func() {
						fmt.Println("控制端开启成功", dc.Label(), dc.ID())
						FileConn.Send <- model.ClientFileResponse{Op: "connection"}
					})
					dc.OnMessage(func(dataMsg webrtc.DataChannelMessage) {
						var res model.ClientFileReceive
						json.Unmarshal(dataMsg.Data, &res)
						// 续传 / 校验握手回执，不能当普通业务消息处理
						if file.IsAck(res.FileOp) {
							file.Ack(res)
							return
						}
						if res.Op == "download" {
							file.SenderHandlerFile(res, dc, FileConn)
						} else {
							FileConn.Send <- dataMsg.Data
						}
					})
					dc.OnClose(func() {
						file.ClearSender(dc)
						dataChannel = nil
					})
					// 队列消费者进程内只起一个，这里只把通道换成当前会话的
					file.SetSender(dc, FileConn)
					file.StartWorker()
				case "new_peer":
					offer, err := p2p.NewOffer(webrtcConn)
					if err != nil {
						fmt.Println("创建文件 offer 失败", err)
						return
					}
					FileRemoteConn.Send <- model.FileSignal{Op: "offer", SDP: offer}
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
				LocalFileList: file.GetAllFiles(res.LocalPath),
			}
		case "local_basic":
			FileConn.Send <- model.ClientFileResponse{
				Op:             res.Op,
				LocalBasicList: file.GetBasicDir(),
			}
		case "remote_ls", "remote_basic", "download":
			if dataChannel != nil {
				dataChannel.Send(msg)
			}
		case "upload":
			// 重试 / 继续时清掉上次的暂停标记
			file.ClearStop(res.RemotePath)
			if dataChannel != nil {
				file.FileList <- res
			}
		case "pause", "cancel":
			// 上传时本机就是发送端；下载时发送端在对端，转发过去
			file.Stop(res.RemotePath)
			if dataChannel != nil {
				dataChannel.Send(msg)
			}
		case "fs_mkdir", "fs_delete":
			if res.Side == "local" {
				if err := file.Apply(res); err != nil {
					FileConn.Send <- model.ClientFileResponse{Op: "error", ErrMsg: err.Error()}
					return
				}
				// 操作完直接把最新的列表回给 UI
				FileConn.Send <- model.ClientFileResponse{
					Op:            "local_ls",
					LocalFileList: file.GetAllFiles(res.LocalPath),
				}
				return
			}
			// 远端文件和本地走同一套 op，这里原样转发，由对端执行后回 remote_ls
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
		file.ClearSender(dataChannel)
		dataChannel.Close()
		dataChannel = nil
	}
	// FileList 与消费者是进程级常驻的，这里不能关掉/重建：
	// 换掉队列会让正在处理旧队列的消费者永久丢任务，新任务没人接。
	fmt.Println("FileConnect结束")
}
