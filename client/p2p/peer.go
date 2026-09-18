// Package p2p 收拢所有 WebRTC 相关逻辑：ICE 配置转换、PeerConnection 构造、
// 候选收发、连接状态监听。之前这段脚手架在 keyboard / file / video 三处各抄了一遍。
package p2p

import (
	"encoding/json"
	"fmt"

	"remote/model"

	"github.com/pion/webrtc/v4"
)

// SendFunc 用于把本地 ICE 候选回传出去（具体走哪条 websocket 由调用方决定）
type SendFunc func(webrtc.ICECandidateInit)

// NewPeer 用统一的 ICE 配置创建 PeerConnection
func NewPeer(ices []model.ICEServer) (*webrtc.PeerConnection, error) {
	return webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: toICEServers(ices),
	})
}

func toICEServers(ices []model.ICEServer) []webrtc.ICEServer {
	out := make([]webrtc.ICEServer, 0, len(ices))
	for _, ice := range ices {
		out = append(out, webrtc.ICEServer{
			URLs:       []string{ice.URL},
			Username:   ice.Username,
			Credential: ice.Credential,
		})
	}
	return out
}

// OnCandidate 注册候选回调；候选为 nil 表示收集结束，不发送
func OnCandidate(pc *webrtc.PeerConnection, send SendFunc) {
	if send == nil {
		return
	}
	pc.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		send(c.ToJSON())
	})
}

// WatchState 监听连接状态，断开（Failed/Closed）时回调 onClosed
func WatchState(pc *webrtc.PeerConnection, tag string, onClosed func()) {
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s >= webrtc.PeerConnectionStateFailed {
			fmt.Printf("[%s] 连接断开: %s\n", tag, s.String())
			if onClosed != nil {
				onClosed()
			}
		}
	})
}

// NewOffer 创建并设置本地 offer
func NewOffer(pc *webrtc.PeerConnection) (*webrtc.SessionDescription, error) {
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return nil, err
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		return nil, err
	}
	return &offer, nil
}

// NewAnswer 设置远端 offer 并创建本地 answer
func NewAnswer(pc *webrtc.PeerConnection, remote *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	if err := pc.SetRemoteDescription(*remote); err != nil {
		return nil, err
	}
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	if err := pc.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	return &answer, nil
}

// SendJSON 用 JSON 往数据通道发一条消息
func SendJSON(dc *webrtc.DataChannel, v any) error {
	if dc == nil {
		return fmt.Errorf("data channel 为空")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return dc.Send(b)
}
