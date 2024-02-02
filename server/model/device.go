package model

type Device struct {
	DeviceID           string `json:"device_id" gorm:"primarykey"`
	IdentificationCode string `json:"identificationCode" gorm:"column:identificationCode"`
	VerificationCode   string `json:"verificationCode" gorm:"column:verificationCode"`
}

func (Device) TableName() string {
	return "device"
}

type Connectioned struct {
	ID                 int    `gorm:"primary_key;type:INTEGER;auto_increment;not null;" json:"id"`
	IdentificationCode string `json:"identificationCode" gorm:"column:identificationCode"`
	ConnectedId        string `json:"connected_id" gorm:"column:connected_id"`
	Note               string `json:"note" gorm:"column:note;default:''"`
}

func (Connectioned) TableName() string {
	return "connectioned"
}
