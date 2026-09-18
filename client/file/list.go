// Package file 负责本地文件列举与基于数据通道的文件传输。
package file

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"remote/model"

	"github.com/shirou/gopsutil/disk"
)

// FormatFileSize 把字节数转成易读格式，保留两位小数
func FormatFileSize(size uint64) string {
	const k = 1024
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	v := float64(size)
	i := 0
	for v >= k && i < len(units)-1 {
		v /= k
		i++
	}
	return fmt.Sprintf("%.2f %s", v, units[i])
}

// GetAllFiles 列出目录下的所有文件与子目录，目录在前
func GetAllFiles(dir string) []model.FileInfo {
	var dirList, fileList []model.FileInfo
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		info := &model.FileInfo{
			FileName: entry.Name(),
			FilePath: filepath.Join(dir, entry.Name()),
		}
		if entry.IsDir() {
			info.IsDir = true
			dirList = append(dirList, *info)
		} else {
			t, _ := os.Stat(info.FilePath)
			info.FileSize = t.Size()
			info.UpdateTime = t.ModTime().Format("2006-01-02 15:04:05")
			fileList = append(fileList, *info)
		}
	}
	return append(dirList, fileList...)
}

// GetBasicDir 返回常用目录与所有分区根目录
func GetBasicDir() map[string]string {
	result := make(map[string]string)
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
