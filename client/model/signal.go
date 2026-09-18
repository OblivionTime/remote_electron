package model

import "github.com/pion/webrtc/v4"

// Signal 是客户端与中转服务器之间收发的统一信令消息。
// 之前分散在 api/client.HandlerResult、utils.WebRTCSendData、
// api/server.FileRemoteReceive 里的重复定义统一到这里。
type Signal struct {
	Op          string `json:"op,omitempty"`
	Device      string `json:"device,omitempty"`
	SendDevice  string `json:"send_device,omitempty"`
	Code        string `json:"code,omitempty"`
	Data        any    `json:"data,omitempty"`
	VideoSender bool   `json:"videoSender,omitempty"`

	ICEServers   []ICEServer                `json:"iceservers,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`

	// 键盘信令（与 VideoSender 同层的 op=keyboard 分支）
	KeyboardOp string `json:"keyboard_op,omitempty"`
}

// InputOperation 是数据通道上收发的一次键鼠操作。
// 之前 SendData（被控端钩子→通道）和 ReceiveJson（通道→被控端）各定义了一份。
type InputOperation struct {
	Method string `json:"method,omitempty"`
	Arg1   int    `json:"arg1,omitempty"`
	Arg2   int    `json:"arg2,omitempty"`
	Key    string `json:"key,omitempty"`

	// 鼠标坐标映射用：控制端画面尺寸
	Width  float64 `json:"width,omitempty"`
	Height float64 `json:"height,omitempty"`

	// 剪贴板
	ClipboardType string   `json:"clipboard_type,omitempty"`
	ClipboardData []string `json:"clipboard_data,omitempty"`
}

// FileSignal 是文件传输数据通道 / 信令帧。
type FileSignal struct {
	Op string `json:"op,omitempty"`

	ICEServers   []ICEServer                `json:"iceservers,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`

	Device string `json:"device,omitempty"`
	Code   string `json:"code,omitempty"`
	ErrMsg string `json:"errmsg,omitempty"`
}
