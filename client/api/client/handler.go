package client

import (
	"encoding/json"
	"fmt"
	"remote/global"
	"remote/model"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/pion/webrtc/v4"
)

type HandlerResult struct {
	Op          string            `json:"op"` //操作
	Device      string            `json:"device,omitempty"`
	Code        string            `json:"code,omitempty"`
	SendDevice  string            `json:"send_device,omitempty"`
	Data        interface{}       `json:"data,omitempty"`
	VideoSender bool              `json:"videoSender,omitempty"`
	ICEServers  []model.ICEServer `json:"iceservers,omitempty"`
	//键盘相关参数
	KeyboardOp   string                     `json:"keyboard_op,omitempty"`
	KeyboardData []byte                     `json:"keyboard_data,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`
}

// 数据处理
func RemoteDataHandler(res []byte, conn *swebsocket.ServerConn) {
	var msg HandlerResult
	json.Unmarshal(res, &msg)
	switch msg.Op {
	case "join", "offer", "answer", "ice_candidate", "ice_server":
		VideoHandler(msg)
	case "disconnected":
		if !msg.VideoSender {
			if global.VideoConn != nil {
				global.VideoConn.Send <- map[string]interface{}{
					"operation": "disconnected",
					"device":    msg.Device,
				}
			}
		} else {
			if global.ClientConn != nil {
				global.ClientConn.Send <- map[string]interface{}{
					"operation": "video_disconnected",
					"device":    msg.Device,
				}
			}
		}
	case "keyboard":
		HandlerKeyboard(msg)
	default:
		fmt.Printf("参数错误%+v\n", msg)
	}
}
