package client

import (
	"fmt"
	"remote/global"
	"remote/model"
	"remote/p2p"

	"github.com/pion/webrtc/v4"
)

// sendKeyboardCandidate 把本地 ICE 候选经中转服务器发回对端
func sendKeyboardCandidate(device string) p2p.SendFunc {
	return func(c webrtc.ICECandidateInit) {
		if global.Remote_serverConn == nil {
			return
		}
		global.Remote_serverConn.Send <- model.Signal{
			Op:           "keyboard",
			KeyboardOp:   "candidate",
			Device:       device,
			ICECandidate: c,
		}
	}
}

// closeKeyboardPeer 释放键盘 P2P 资源
func closeKeyboardPeer() {
	if global.KeyboardP2PConn != nil {
		global.KeyboardP2PConn.Close()
		global.KeyboardP2PConn = nil
	}
	if global.KeyboardHandler != nil {
		global.KeyboardHandler.Close()
		global.KeyboardHandler = nil
	}
}

// startKeyboardReceiver 被控端：等待控制端建好的数据通道
func startKeyboardReceiver(ices []model.ICEServer, device string) {
	closeKeyboardPeer()
	peer, err := p2p.NewKeyboardPeer(ices, sendKeyboardCandidate(device), closeKeyboardPeer)
	if err != nil {
		fmt.Println("初始化键盘 P2P 失败", err)
		return
	}
	global.KeyboardP2PConn = peer.PC
	peer.Receiver(func(dc *webrtc.DataChannel) {
		global.KeyboardHandler = dc
	})
}

// startKeyboardSender 控制端：创建数据通道并把 offer 发给对端
func startKeyboardSender(ices []model.ICEServer, device string) {
	closeKeyboardPeer()
	peer, err := p2p.NewKeyboardPeer(ices, sendKeyboardCandidate(device), closeKeyboardPeer)
	if err != nil {
		fmt.Println("初始化键盘 P2P 失败", err)
		return
	}
	offer, err := peer.Sender(global.DeviceInfo.IdentificationCode, nil)
	if err != nil {
		fmt.Println("创建键盘 offer 失败", err)
		return
	}
	global.KeyboardP2PConn = peer.PC
	global.KeyboardHandler = peer.DC
	if global.Remote_serverConn != nil {
		global.Remote_serverConn.Send <- model.Signal{
			Op:         "keyboard",
			KeyboardOp: "offer",
			SDP:        offer,
			Device:     device,
			SendDevice: global.DeviceInfo.IdentificationCode,
		}
	}
}

// HandlerKeyboard 处理键盘 P2P 的 offer / answer / candidate
func HandlerKeyboard(msg model.Signal) {
	if global.KeyboardP2PConn == nil {
		return
	}
	switch msg.KeyboardOp {
	case "offer":
		answer, err := p2p.NewAnswer(global.KeyboardP2PConn, msg.SDP)
		if err != nil {
			fmt.Println("创建键盘 answer 失败", err)
			return
		}
		if global.Remote_serverConn != nil {
			global.Remote_serverConn.Send <- model.Signal{
				Op:         "keyboard",
				KeyboardOp: "answer",
				SDP:        answer,
				Device:     msg.SendDevice,
			}
		}
	case "answer":
		if err := global.KeyboardP2PConn.SetRemoteDescription(*msg.SDP); err != nil {
			fmt.Println("设置键盘 answer 失败", err)
		}
	case "candidate":
		if err := global.KeyboardP2PConn.AddICECandidate(msg.ICECandidate); err != nil {
			fmt.Println("添加键盘候选失败", err)
		}
	}
}
