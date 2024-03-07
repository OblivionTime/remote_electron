package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"remote/global"
	"remote/model"

	"github.com/pion/webrtc/v4"
	"github.com/shirou/gopsutil/disk"
)

// formatFileSize 将文件大小转换为易读的格式
func FormatFileSize(size uint64) string {
	// 定义文件大小单位
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}

	// 处理文件大小为0的情况
	if size == 0 {
		return "0 B"
	}

	// 计算文件大小所在单位的索引
	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	// 格式化文件大小
	return fmt.Sprintf("%d %s", size, units[unitIndex])
}

// 初始化接收端文件管理
func InitReceiveFileP2P(ICEServers []model.ICEServer, channelHandler func(d *webrtc.DataChannel)) *webrtc.PeerConnection {
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
	Conn, _ := webrtc.NewPeerConnection(config)

	// Register data channel creation handling
	Conn.OnDataChannel(channelHandler)
	fmt.Println("InitReceiveFileP2P successfully initialized")
	return Conn
}

// 初始化发送端
func InitSendFileP2P(ICEServers []model.ICEServer) (*webrtc.PeerConnection, *webrtc.DataChannel) {
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
	Conn, _ := webrtc.NewPeerConnection(config)
	var dataChannel *webrtc.DataChannel

	dataChannel, _ = Conn.CreateDataChannel(global.DeviceInfo.IdentificationCode, nil)
	fmt.Println("InitSendFileP2P successfully initialized")
	return Conn, dataChannel
}

// 获取目录下所有文件和目录
func GetAllFiles(dir string) []model.FileInfo {
	var dirList, fileList []model.FileInfo
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		info := &model.FileInfo{
			FileName:   entry.Name(),
			FilePath:   filepath.Join(dir, entry.Name()),
			IsDir:      false,
			FileSize:   0,
			UpdateTime: "",
		}
		if entry.IsDir() {
			info.IsDir = true
			dirList = append(dirList, *info)
		} else {
			t, _ := os.Stat(info.FilePath)
			info.FileSize = int64(t.Size())
			info.UpdateTime = t.ModTime().Format("2006-01-02 15:04:05")
			fileList = append(fileList, *info)
		}
	}
	return append(dirList, fileList...)
}

// 获取本地的下载目录,桌面,文档,图库,用户目录
func GetBasicDir() map[string]string {
	result := make(map[string]string)
	// 获取当前用户
	currentUser, _ := user.Current()
	homeDir := currentUser.HomeDir
	partitions, _ := disk.Partitions(false)
	for _, item := range partitions {
		result[item.Device] = item.Mountpoint + "/"
	}
	result["用户"] = homeDir
	result["下载"] = filepath.Join(homeDir, "Downloads")
	result["桌面"] = filepath.Join(homeDir, "Desktop")
	result["文档"] = filepath.Join(homeDir, "Documents")
	result["图库"] = filepath.Join(homeDir, "Pictures")
	return result

}
