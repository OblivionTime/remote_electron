package utils

import (
	"encoding/json"

	"github.com/pion/webrtc/v4"
)

func SendMessage(data interface{}, dataChannel *webrtc.DataChannel) {
	successInfo, _ := json.Marshal(data)
	dataChannel.Send(successInfo)
}
