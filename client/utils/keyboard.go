package utils

import (
	"fmt"

	"github.com/go-vgo/robotgo"
	"github.com/pion/webrtc/v4"
)

type WebRTCSendData struct {
	Op           string                     `json:"op"` //操作
	Device       string                     `json:"device,omitempty"`
	SendDevice   string                     `json:"send_device,omitempty"`
	KeyboardOp   string                     `json:"keyboard_op,omitempty"`
	ICECandidate webrtc.ICECandidateInit    `json:"icecandidate,omitempty"`
	SDP          *webrtc.SessionDescription `json:"sdp,omitempty"`
}
type ReceiveJson struct {
	Method        string   `json:"method,omitempty"`
	Arg1          int      `json:"arg1,omitempty"`
	Arg2          int      `json:"arg2,omitempty"`
	Key           string   `json:"key,omitempty"`
	Width         float64  `json:"width,omitempty"`
	Height        float64  `json:"height,omitempty"`
	ClipboardType string   `json:"clipboard_type,omitempty"`
	ClipboardData []string `json:"clipboard_data,omitempty"`
}

// 分辨率映射
func mapValues(value, inMin, inMax, outMin, outMax float64) int {
	return int((value-inMin)*(outMax-outMin)/(inMax-inMin) + outMin)
}
func Operation(oper ReceiveJson) error {
	method := oper.Method
	switch method {
	case "MouseMove":
		width, height := robotgo.GetScreenSize()
		arg1 := mapValues(float64(oper.Arg1), 0, oper.Width, 0, float64(width))
		arg2 := mapValues(float64(oper.Arg2), 0, oper.Height, 0, float64(height))
		robotgo.Move(arg1, arg2)
	case "MouseDown":
		if oper.Arg1 == 1 {
			robotgo.Toggle("left")
		} else if oper.Arg1 == 2 {
			robotgo.Toggle("center")
		} else {
			robotgo.Toggle("right")
		}
	case "MouseUp":
		if oper.Arg1 == 1 {
			robotgo.Toggle("left", "up")
		} else if oper.Arg1 == 2 {
			robotgo.Toggle("center", "up")
		} else {
			robotgo.Toggle("right", "up")
		}
	case "ScrollMouse":
		if oper.Arg1 == -1 {
			robotgo.ScrollDir(1, "down")
		} else {
			robotgo.ScrollDir(1, "up")
		}
	case "KeyDown":
		return robotgo.KeyDown(oper.Key)
	case "KeyRelease":
		return robotgo.KeyUp(oper.Key)
	case "clipboard":
		fmt.Println("监听到剪贴板了")
		for _, c := range oper.ClipboardData {
			robotgo.WriteAll(c)
		}
		return nil
	default:
		return fmt.Errorf("参数不正确")
	}
	return nil
}
