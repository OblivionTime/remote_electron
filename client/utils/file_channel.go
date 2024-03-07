package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"remote/model"
	"time"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/pion/webrtc/v4"
)

var FileList = make(chan model.ClientFileReceive, 1024)

const BUF_WAITING_THRESHOLD uint64 = 1024 * 1000

var DataToBeSent = make(chan []byte)

func HandlerFile(FileConn *swebsocket.ServerConn, dataChannel *webrtc.DataChannel) {
	for res := range FileList {
		file, err := os.Open(res.FilePath)
		if err != nil {
			FileConn.Send <- model.ClientFileResponse{
				Op:     res.Op,
				ErrMsg: "文件出错" + err.Error(),
				Status: 3,
			}
			continue
		}

		ChannelSendData(model.ClientFileReceive{
			Op:         res.Op,
			FileOp:     "start",
			FilePath:   res.FilePath,
			FileSize:   res.FileSize,
			RemotePath: res.RemotePath,
		}, dataChannel)
		chunkSize := 10240
		fmt.Println("开始传输文件", res.FilePath, res.FileSize)
		// 定义一个缓冲区，用于存储每个分片的数据
		offset := 0
		pos := 0
		ExitChan := make(chan bool)
		buffer := make([]byte, chunkSize)
		FileName := filepath.Base(res.FilePath)
		go func() {
			for {
				// 读取分片数据到缓冲区
				readSize, err := file.Read(buffer)
				if err != nil {
					break
				}
				if dataChannel.BufferedAmount() >= BUF_WAITING_THRESHOLD {
					t, _ := json.Marshal(model.ClientFileReceive{
						Op:          res.Op,
						FileOp:      "progress",
						RemotePath:  res.RemotePath,
						FileSize:    res.FileSize,
						FileContent: buffer[:readSize],
						Offset:      offset,
					})
					DataToBeSent <- t

				} else {
					ChannelSendData(model.ClientFileReceive{
						Op:          res.Op,
						FileOp:      "progress",
						RemotePath:  res.RemotePath,
						FileSize:    res.FileSize,
						FileContent: buffer[:readSize],
						Offset:      offset,
					}, dataChannel)
				}

				offset += readSize
			}
			fmt.Println("ChannelSendData写完了")
			file.Close()
			ExitChan <- true

		}()
		flag := false
		for {
			if flag {
				break
			}

			select {
			case <-ExitChan:
				if res.Op == "upload" {
					FileConn.Send <- model.ClientFileResponse{
						Op:         res.Op,
						Speed:      "0 MB/s",
						FileName:   FileName,
						FilePath:   res.FilePath,
						FileSize:   FormatFileSize(uint64(res.FileSize)),
						RemotePath: res.RemotePath,
						Progress:   100,
						Status:     2,
					}
				} else {
					ChannelSendData(model.ClientFileReceive{
						Op:         res.Op,
						FileOp:     "status",
						RemotePath: res.RemotePath,
						FilePath:   res.FilePath,
						FileSize:   res.FileSize,
						Progress:   100,
						Speed:      "0 MB/s",
						Status:     2,
					}, dataChannel)
				}

				flag = true
			case <-time.After(time.Second):
				speed := FormatFileSize(uint64(offset - pos))
				pos = offset
				progress := int((float64(offset) / float64(res.FileSize)) * 100)
				if progress >= 98 {
					progress = 98
				}
				if res.Op == "upload" {
					FileConn.Send <- model.ClientFileResponse{
						Op:         res.Op,
						Speed:      speed + "/s",
						FileName:   FileName,
						FilePath:   res.FilePath,
						FileSize:   FormatFileSize(uint64(res.FileSize)),
						RemotePath: res.RemotePath,
						Progress:   progress,
						Status:     1,
					}
				} else {
					ChannelSendData(model.ClientFileReceive{
						Op:         res.Op,
						FileOp:     "status",
						RemotePath: res.RemotePath,
						FilePath:   res.FilePath,
						FileSize:   res.FileSize,
						Progress:   progress,
						Speed:      speed + "/s",
						Status:     1,
					}, dataChannel)
				}

			}
		}
	}

}
func SenderHandlerFile(msg model.ClientFileReceive, dataChannel *webrtc.DataChannel, FileConn *swebsocket.ServerConn) {
	switch msg.FileOp {
	case "start":
		file, _ := os.Create(msg.RemotePath)
		fmt.Println("接收到了", msg.FileOp, msg.RemotePath, msg.FileSize)
		file.Close()
	case "progress":
		file, _ := os.OpenFile(msg.RemotePath, os.O_RDWR|os.O_CREATE, 0644)
		file.WriteAt(msg.FileContent, int64(msg.Offset))
		file.Close()
		f, _ := os.Stat(msg.RemotePath)
		if f.Size() >= msg.FileSize {
			fmt.Println("写完了")

		}
	case "status":
		FileConn.Send <- model.ClientFileResponse{
			Op:         msg.Op,
			Speed:      msg.Speed,
			FileName:   filepath.Base(msg.FilePath),
			FilePath:   msg.FilePath,
			FileSize:   FormatFileSize(uint64(msg.FileSize)),
			RemotePath: msg.RemotePath,
			Progress:   msg.Progress,
			Status:     msg.Status,
		}
	}

}
func ChannelSendData(data interface{}, dataChannel *webrtc.DataChannel) {
	v, _ := json.Marshal(data)
	if dataChannel != nil {
		dataChannel.Send(v)
	}
}
