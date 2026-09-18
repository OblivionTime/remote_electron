package p2p

import (
	"fmt"

	"remote/model"

	"github.com/pion/webrtc/v4"
)

// NewFileReceiver 文件接收端（被控端）：等待控制端建好的数据通道
func NewFileReceiver(ices []model.ICEServer, onChannel func(*webrtc.DataChannel)) (*webrtc.PeerConnection, error) {
	pc, err := NewPeer(ices)
	if err != nil {
		return nil, err
	}
	pc.OnDataChannel(onChannel)
	fmt.Println("文件 P2P（接收端）初始化完成")
	return pc, nil
}

// NewFileSender 文件发送端（控制端）：创建数据通道，offer 由调用方发送
func NewFileSender(ices []model.ICEServer, label string) (*webrtc.PeerConnection, *webrtc.DataChannel, error) {
	pc, err := NewPeer(ices)
	if err != nil {
		return nil, nil, err
	}
	dc, err := pc.CreateDataChannel(label, nil)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("文件 P2P（发送端）初始化完成")
	return pc, dc, nil
}
