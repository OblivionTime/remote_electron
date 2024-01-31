package model

type ICEServer struct {
	URL        string `json:"url"`
	Credential string `json:"credential"`
	Username   string `json:"username"`
}
type Device struct {
	DeviceID           string `json:"device_id" `
	IdentificationCode string `json:"identificationCode" `
	VerificationCode   string `json:"verificationCode"`
	Connectioned       string `json:"connectioned" ` //连接过的设备
}
