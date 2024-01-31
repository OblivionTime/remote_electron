package utils

import (
	"encoding/json"
	"fmt"
	"remote/global"
	"remote/model"

	"github.com/pion/webrtc/v4"
)

// 初始化接收端键盘webrtc
func InitReceiveKeyboardP2P(ICEServers []model.ICEServer, device string) {
	if global.KeyboardP2PConn != nil {
		global.KeyboardP2PConn.Close()
		global.KeyboardP2PConn = nil
	}
	if global.KeyboardHandler != nil {
		global.KeyboardHandler.Close()
		global.KeyboardHandler = nil
	}

	iceServer := make([]webrtc.ICEServer, 0)
	for _, ice := range ICEServers {
		iceServer = append(iceServer, webrtc.ICEServer{
			URLs:       []string{ice.URL},
			Username:   ice.Username,
			Credential: ice.Credential,
		})

	}
	config := webrtc.Configuration{
		ICEServers: iceServer,
	}
	global.KeyboardP2PConn, _ = webrtc.NewPeerConnection(config)
	global.KeyboardP2PConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {

		if s >= 4 {
			fmt.Printf("Peer Connection State has changed: %s\n", s.String())
			fmt.Println("连接断开", s.String())
			if global.KeyboardHandler != nil {
				global.KeyboardP2PConn.Close()
				global.KeyboardP2PConn = nil
				global.KeyboardHandler.Close()
				global.KeyboardHandler = nil
			}
		}
	})
	global.KeyboardP2PConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		global.CandidatesMux.Lock()
		defer global.CandidatesMux.Unlock()
		global.Remote_serverConn.Send <- WebRTCSendData{
			Op:           "keyboard",
			KeyboardOp:   "candidate",
			SDP:          nil,
			ICECandidate: c.ToJSON(),
			Device:       device,
		}
	})
	// Register data channel creation handling
	global.KeyboardP2PConn.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())
		global.KeyboardHandler = d
		// Register channel opening handling
		global.KeyboardHandler.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. \n", d.Label(), d.ID())
		})

		// Register text message handling
		global.KeyboardHandler.OnMessage(func(msg webrtc.DataChannelMessage) {
			var data ReceiveJson
			json.Unmarshal(msg.Data, &data)
			err := Operation(data)
			if err != nil {
				fmt.Println("Operation(data)", err)
			}
		})
		global.KeyboardHandler.OnClose(func() {
			if global.KeyboardP2PConn != nil {
				global.KeyboardP2PConn.Close()
				global.KeyboardP2PConn = nil
			}
			fmt.Printf("close")
		})
	})
	fmt.Println("InitSenderKeyboardP2P successfully initialized")
}

// 初始化发送端键盘webrtc
func InitSenderKeyboardP2P(ICEServers []model.ICEServer, device string) {
	if global.KeyboardP2PConn != nil {
		global.KeyboardP2PConn.Close()
		global.KeyboardP2PConn = nil
	}
	if global.KeyboardHandler != nil {
		global.KeyboardHandler.Close()
		global.KeyboardHandler = nil
	}

	iceServer := make([]webrtc.ICEServer, 0)
	for _, ice := range ICEServers {
		iceServer = append(iceServer, webrtc.ICEServer{
			URLs:       []string{ice.URL},
			Username:   ice.Username,
			Credential: ice.Credential,
		})

	}
	config := webrtc.Configuration{
		ICEServers: iceServer,
	}
	global.KeyboardP2PConn, _ = webrtc.NewPeerConnection(config)
	global.KeyboardP2PConn.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {

		if s >= 4 {
			fmt.Printf("Peer Connection State has changed: %s\n", s.String())
			fmt.Println("连接断开", s.String())
			if global.KeyboardHandler != nil {
				global.KeyboardP2PConn.Close()
				global.KeyboardP2PConn = nil
				global.KeyboardHandler.Close()
				global.KeyboardHandler = nil
			}
		}
	})
	global.KeyboardP2PConn.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		global.CandidatesMux.Lock()
		defer global.CandidatesMux.Unlock()
		global.Remote_serverConn.Send <- WebRTCSendData{
			Op:         "keyboard",
			KeyboardOp: "candidate",

			ICECandidate: c.ToJSON(),
			Device:       device,
		}
	})
	global.KeyboardHandler, _ = global.KeyboardP2PConn.CreateDataChannel(global.DeviceInfo.IdentificationCode, nil)
	global.KeyboardHandler.OnOpen(func() {
		fmt.Println("控制端开启成功")
	})
	global.KeyboardHandler.OnClose(func() {
		if global.KeyboardP2PConn != nil {
			global.KeyboardP2PConn.Close()
			global.KeyboardP2PConn = nil
		}
		fmt.Printf("close")
	})
	//发送offer
	offer, _ := global.KeyboardP2PConn.CreateOffer(nil)
	global.KeyboardP2PConn.SetLocalDescription(offer)
	global.Remote_serverConn.Send <- map[string]interface{}{
		"op":          "keyboard",
		"keyboard_op": "offer",
		"SDP":         offer,
		"device":      device,
		"send_device": global.DeviceInfo.IdentificationCode,
	}
	fmt.Println("InitReceiveKeyboardP2P successfully initialized")
}
