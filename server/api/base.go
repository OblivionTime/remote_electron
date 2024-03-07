package api

import (
	"fmt"
	"remote_server/global"
	"remote_server/model"
	"remote_server/model/common/response"
	"remote_server/utils"

	"github.com/gin-gonic/gin"
)

type ConnectDevice struct {
	DeviceID           string `json:"device_id"`
	IdentificationCode string `json:"identificationCode" gorm:"column:identificationCode"`
	VerificationCode   string `json:"verificationCode" gorm:"column:verificationCode"`
	Note               string `json:"note" gorm:"column:note"`
}

// 获取自己的识别码和验证码
func GetCode(ctx *gin.Context) {
	DeviceID := ctx.Query("device")
	if DeviceID == "" {
		response.FailWithMessage("未获取到设备信息", ctx)
		return
	}
	//判断当前设备是否已经存在
	var deviceInfo model.Device
	res := global.DB.Model(model.Device{}).Where("device_id = ?", DeviceID).First(&deviceInfo)
	if res.Error != nil || res.RowsAffected == 0 {
		//不存在
		deviceInfo.DeviceID = DeviceID
		deviceInfo.IdentificationCode = utils.RandomlyGeneratedIdentificationCodes()
		deviceInfo.VerificationCode = utils.RandomlyGenerateVerificationCodes()

		//判断识别码是否重复
		global.DB.Model(model.Device{}).Create(deviceInfo)
	}
	fmt.Printf("%s的识别码为%s,验证码为%s\n", deviceInfo.DeviceID, deviceInfo.IdentificationCode, deviceInfo.VerificationCode)

	response.OkWithData(deviceInfo, ctx)
}

// 判断设备是否连接
func DeviceOnlineStatus(ctx *gin.Context) {
	DeviceID := ctx.Query("code")
	var msg model.Device
	if err := ctx.BindJSON(&msg); err != nil {
		response.Fail(ctx)
		return
	}
	if _, ok := global.DeviceList[msg.IdentificationCode]; !ok {
		response.FailWithMessage("对方不在线!!", ctx)
		return
	}
	//判断验证码是否正确
	var deviceInfo model.Device
	global.DB.Model(model.Device{}).Where("identificationCode = ?", msg.IdentificationCode).First(&deviceInfo)
	if deviceInfo.VerificationCode != msg.VerificationCode {
		response.FailWithMessage("验证码错误!!", ctx)
		return
	}
	//判断当前设备是否连接过
	var connected model.Connectioned
	res := global.DB.Model(model.Connectioned{}).Where("identificationCode = ? and connected_id=?", DeviceID, msg.IdentificationCode).First(&connected)
	if res.RowsAffected == 0 || res.Error != nil {
		res = global.DB.Model(model.Connectioned{}).Create(&model.Connectioned{
			IdentificationCode: DeviceID,  
			ConnectedId:        msg.IdentificationCode,
		})
		fmt.Println(res.Error)
	}
	response.Ok(ctx)
}

// 获取连接过的设备列表
func GetConnectDeviceList(ctx *gin.Context) {
	DeviceID := ctx.Query("device")
	if DeviceID == "" {
		response.FailWithMessage("未获取到设备信息", ctx)
		return
	}
	var deviceList []ConnectDevice
	global.DB.Raw("select d.*,c.note from connectioned as c,device as d where c.identificationCode=? and d.identificationCode=c.connected_id", DeviceID).Scan(&deviceList)
	fmt.Println(deviceList)
	response.OkWithData(deviceList, ctx)
}
