package client

import (
	"fmt"
	"remote/global"
)

func HandlerKeyboard(msg HandlerResult) {

	switch msg.KeyboardOp {
	case "offer":
		if global.KeyboardP2PConn != nil {
			global.KeyboardP2PConn.SetRemoteDescription(*msg.SDP)
			//发送offer
			answer, _ := global.KeyboardP2PConn.CreateAnswer(nil)
			global.KeyboardP2PConn.SetLocalDescription(answer)
			global.Remote_serverConn.Send <- map[string]interface{}{
				"op":          "keyboard",
				"keyboard_op": "answer",
				"SDP":         answer,
				"device":      msg.SendDevice,
			}
		}

	case "answer":
		global.KeyboardP2PConn.SetRemoteDescription(*msg.SDP)
	case "candidate":
		if err := global.KeyboardP2PConn.AddICECandidate(msg.ICECandidate); err != nil {
			fmt.Println(err)
		}
	}

}
