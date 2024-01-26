package utils

import (
	"fmt"
	"math/rand"
	"remote_server/global"
	"remote_server/model"
	"time"
)

// 随机生成识别码
func GenerateRandomNumber() string {
	// 设置随机种子
	rand.NewSource(time.Now().UnixNano())
	// 生成随机的9位数字
	return fmt.Sprintf("%v", rand.Intn(1e9))
}
func RandomlyGeneratedIdentificationCodes() string {
	var deviceInfo model.Device
	for {
		s := GenerateRandomNumber()
		res := global.DB.Model(model.Device{}).Where("identificationCode = ?", s).Scan(&deviceInfo)
		if res.RowsAffected == 0 {
			return s
		}

	}
}

// 随机生成验证码
func RandomlyGenerateVerificationCodes() string {
	length := 6
	// 字符集合
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	rand.NewSource(time.Now().UnixNano())

	// 生成随机的字符串
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
