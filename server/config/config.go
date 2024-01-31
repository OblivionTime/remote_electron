package config

import (
	"fmt"
	"os"
	"remote_server/utils"

	"gopkg.in/yaml.v2"
)

var Config *RemoteConfig

type RemoteConfig struct {
	//服务器地址
	ServerAddr string `yaml:"serveraddr" json:"serveraddr"` //server address
	//Turn服务器地址
	Turn   TurnConfig `yaml:"turn" json:"turn"`
	DBPath string     `yaml:"db_path" json:"db_path"` //数据库相关配置
}

// Turn服务器相关配置
type TurnConfig struct {
	//公网地址
	PublicIP string `yaml:"public_ip" json:"public_ip"`
	//端口
	Port int `yaml:"port" json:"port"`
	//协程数
	ThreadNum uint `yaml:"thread_num" json:"thread_num"`
}

func init() {
	yamlFile, err := os.ReadFile(`config.yaml`)
	if err != nil {
		fmt.Printf("yamlFile.Get err   #%v \n", err)
	}
	Config = &RemoteConfig{}
	err = yaml.Unmarshal(yamlFile, Config)
	if err != nil {
		fmt.Printf("Unmarshal   #%v \n", err)
	}
	if Config.Turn.PublicIP == "" {
		Config.Turn.PublicIP = utils.GetHostIp()
	}
	if Config.Turn.ThreadNum == 0 {
		Config.Turn.ThreadNum = 1
	}
}
