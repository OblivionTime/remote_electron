package model

type Device struct {
	DeviceID           string `json:"device_id" gorm:"primarykey"`
	IdentificationCode string `json:"identificationCode" gorm:"column:identificationCode"`
	VerificationCode   string `json:"verificationCode" gorm:"column:verificationCode"`
	Connectioned       string `json:"connectioned" gorm:"column:connectioned"` //连接过的设备
}

func (Device) TableName() string {
	return "device"
}
