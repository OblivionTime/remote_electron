package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"remote/model"

	"gitee.com/solidone/sutils/swebsocket"
	"github.com/pion/webrtc/v4"
)

const (
	BUF_WAITING_THRESHOLD uint64 = 1024 * 1000
	chunkSize                    = 10240
	// 等接收端回执（续传偏移）的最长时间；超时就当对方没这文件，从 0 开始
	startAckTimeout = 5 * time.Second
	// 等接收端整文件校验结果
	verifyTimeout = 30 * time.Second
	// 背压：缓冲区满时轮询等待的间隔与总时长上限
	backpressureInterval = 20 * time.Millisecond
	maxBackpressureWait  = 30 * time.Second
)

// ErrStopped 传输被用户暂停 / 取消
var ErrStopped = errors.New("传输已停止")

var FileList = make(chan model.ClientFileReceive, 1024)

// CRC32C（Castagnoli），amd64 上有 SSE4.2 硬件加速
var crcTable = crc32.MakeTable(crc32.Castagnoli)

// #region 停止标记（暂停 / 取消）------------------------------------------------------------
var (
	stopMu  sync.Mutex
	stopped = make(map[string]bool) // key: 目标路径
)

// Stop 请求停止某个目标的传输，暂停和取消都走这里
func Stop(dest string) {
	stopMu.Lock()
	stopped[dest] = true
	stopMu.Unlock()
}

// ClearStop 清除停止标记（继续 / 重试前调用）
func ClearStop(dest string) {
	stopMu.Lock()
	delete(stopped, dest)
	stopMu.Unlock()
}

func isStopped(dest string) bool {
	stopMu.Lock()
	defer stopMu.Unlock()
	return stopped[dest]
}

//#endregion -----------------------------------------------------------

// #region 握手回执 ------------------------------------------------------------
// ackCh 接收端回执通道。HandlerFile 是串行的，同一时刻只有一个文件在传，单通道够用。
var ackCh = make(chan model.ClientFileReceive, 128)

// 握手类消息（不是业务数据，收到要直接交给等待方，不能转发给 UI / 对端）
var ackOps = map[string]bool{"start_ack": true, "verify": true}

func IsAck(fileOp string) bool { return ackOps[fileOp] }

// Ack 接收端回执：start_ack 带已有内容的偏移与 CRC，verify 带整文件校验结果
func Ack(msg model.ClientFileReceive) {
	select {
	case ackCh <- msg:
	default:
	}
}

func drainAcks() {
	for {
		select {
		case <-ackCh:
		default:
			return
		}
	}
}

func waitAck(remotePath, fileOp string, timeout time.Duration) (model.ClientFileReceive, bool) {
	deadline := time.After(timeout)
	for {
		select {
		case msg := <-ackCh:
			if msg.RemotePath == remotePath && msg.FileOp == fileOp {
				return msg, true
			}
		case <-deadline:
			return model.ClientFileReceive{}, false
		}
	}
}

//#endregion -----------------------------------------------------------

// fileCrc 计算文件内容的 CRC32C；n > 0 时只算前 n 字节
func fileCrc(path string, n int64) (uint32, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := crc32.New(crcTable)
	var r io.Reader = f
	if n > 0 {
		r = io.LimitReader(f, n)
	}
	if _, err := io.Copy(h, r); err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

// resumeOffset 决定续传起点：只有对端已有内容的 CRC 和源文件前 offset 字节的 CRC
// 完全一致才续传，否则返回 0 从头重传。只比大小是不够的——源文件在中断期间被改过、
// 大小又恰好没变的话，续传会拼出一个错误的文件。
func resumeOffset(offset int64, ackedCrc, prefixCrc uint32) int64 {
	if offset <= 0 || ackedCrc != prefixCrc {
		return 0
	}
	return offset
}

// #region 会话通道 / 队列消费者 ------------------------------------------------------------
// 消费者在进程内只启动一次，会话切换只换它手里的数据通道。
// 之前是每个会话各起一个 HandlerFile、并把 FileList 整个换掉：旧会话的消费者卡在
// streamFile 里不会退出，收尾又可能卡住导致 close(FileList) 不执行，于是新旧消费者
// 抢同一条队列——任务被旧消费者抢走发往已死的通道，UI 永远收不到状态（卡在"等待"）。
var (
	hubMu      sync.RWMutex
	hubDC      *webrtc.DataChannel
	hubConn    *swebsocket.ServerConn
	workerOnce sync.Once
)

// SetSender 注册当前会话的数据通道（新会话直接覆盖）
func SetSender(dc *webrtc.DataChannel, conn *swebsocket.ServerConn) {
	hubMu.Lock()
	hubDC, hubConn = dc, conn
	hubMu.Unlock()
}

// ClearSender 只清掉属于自己那条通道的注册，避免旧会话收尾时把新会话的通道清掉
func ClearSender(dc *webrtc.DataChannel) {
	if dc == nil {
		return
	}
	hubMu.Lock()
	if hubDC == dc {
		hubDC, hubConn = nil, nil
	}
	hubMu.Unlock()
}

func sender() (*webrtc.DataChannel, *swebsocket.ServerConn) {
	hubMu.RLock()
	defer hubMu.RUnlock()
	return hubDC, hubConn
}

// waitSender 等一条可用的数据通道。会话正在切换时任务会在这里等新通道，
// 而不是被旧消费者抢走后发进黑洞。
func waitSender(timeout time.Duration) (*webrtc.DataChannel, *swebsocket.ServerConn, bool) {
	deadline := time.Now().Add(timeout)
	for {
		dc, conn := sender()
		if dc != nil && dc.ReadyState() == webrtc.DataChannelStateOpen {
			return dc, conn, true
		}
		if time.Now().After(deadline) {
			return dc, conn, false
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// StartWorker 启动唯一的队列消费者（幂等，可在每次会话建立时调用）
func StartWorker() {
	workerOnce.Do(func() {
		go func() {
			for res := range FileList {
				dc, conn, ok := waitSender(15 * time.Second)
				if !ok {
					fmt.Println("数据通道未就绪，跳过:", res.FilePath)
					reportStatus(res, 0, 3, "", "数据通道未就绪，请重试", conn, dc)
					continue
				}
				if res.FileOp == "dir" {
					sendDir(res, dc, conn)
				} else {
					sendFile(res, dc, conn)
				}
			}
		}()
	})
}

//#endregion -----------------------------------------------------------

func sendFile(res model.ClientFileReceive, dc *webrtc.DataChannel, FileConn *swebsocket.ServerConn) {
	var sent atomic.Int64
	pos := int64(0)
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				cur := sent.Load()
				speed := FormatFileSize(uint64(cur - pos))
				pos = cur
				reportStatus(res, percent(cur, res.FileSize), 1, speed+"/s", "", FileConn, dc)
			}
		}
	}()

	err := streamFile(res.Op, res.FilePath, res.RemotePath, res.FileSize, dc, func(n int) {
		sent.Add(int64(n))
	})
	close(done)

	// 用户主动暂停 / 取消：状态由 UI 决定，这里不报错
	if errors.Is(err, ErrStopped) {
		return
	}
	if err != nil {
		reportStatus(res, percent(sent.Load(), res.FileSize), 3, "0 MB/s", err.Error(), FileConn, dc)
		return
	}
	reportStatus(res, 100, 2, "0 MB/s", "", FileConn, dc)
}

// sendDir 递归发送整个目录：先在对端建好目录结构，再逐个文件续传，
// 进度按整个目录的字节数聚合上报，UI 上只显示一行。
func sendDir(res model.ClientFileReceive, dc *webrtc.DataChannel, FileConn *swebsocket.ServerConn) {
	dirs, files, total := collectDir(res.FilePath, res.RemotePath)
	res.FileSize = total

	if len(files) == 0 {
		reportStatus(res, 100, 2, "0 MB/s", "", FileConn, dc)
		return
	}

	// 目录结构先建出来（父目录在前）
	for _, d := range dirs {
		if err := ChannelSendData(model.ClientFileReceive{Op: res.Op, FileOp: "mkdir", RemotePath: d}, dc); err != nil {
			reportStatus(res, 0, 3, "0 MB/s", "建目录失败: "+err.Error(), FileConn, dc)
			return
		}
	}

	var sent atomic.Int64
	pos := int64(0)
	done := make(chan struct{})
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				cur := sent.Load()
				speed := FormatFileSize(uint64(cur - pos))
				pos = cur
				reportStatus(res, percent(cur, total), 1, speed+"/s", "", FileConn, dc)
			}
		}
	}()

	var failed int
	var stopped bool
	var firstErr string
	for _, f := range files {
		err := streamFile(res.Op, f.path, f.dest, f.size, dc, func(n int) {
			sent.Add(int64(n))
		})
		if errors.Is(err, ErrStopped) {
			stopped = true
			break
		}
		if err != nil {
			failed++
			if firstErr == "" {
				firstErr = filepath.Base(f.path) + ": " + err.Error()
			}
		}
	}
	close(done)

	if stopped {
		return
	}
	if failed > 0 {
		reportStatus(res, percent(sent.Load(), total), 3, "0 MB/s",
			fmt.Sprintf("%d/%d 个文件失败（%s）", failed, len(files), firstErr), FileConn, dc)
		return
	}
	reportStatus(res, 100, 2, "0 MB/s", "", FileConn, dc)
}

// streamFile 把 src 分片写到对端 dst。
// 续传前用 CRC 校验对端已有内容确实是源文件的真实前缀；传完后把整文件 CRC 发给对端复核。
func streamFile(op, src, dst string, size int64, dc *webrtc.DataChannel, onBytes func(int)) error {
	if isStopped(dst) {
		return ErrStopped
	}
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	drainAcks()
	if err := ChannelSendData(model.ClientFileReceive{
		Op: op, FileOp: "start", FilePath: src, FileSize: size, RemotePath: dst,
	}, dc); err != nil {
		return fmt.Errorf("发送失败，连接可能已断开: %w", err)
	}

	ack, acked := waitAck(dst, "start_ack", startAckTimeout)
	var offset int64
	if acked && ack.Offset > 0 && int64(ack.Offset) <= size {
		offset = int64(ack.Offset)
	}

	h := crc32.New(crcTable)

	if offset > 0 {
		// 读源文件的前 offset 字节，一边算前缀 CRC（hasher 继续复用，滚出整文件 CRC）
		_, readErr := io.CopyN(h, f, offset)
		if readErr != nil || resumeOffset(offset, ack.Crc, h.Sum32()) == 0 {
			// 对端内容不是源文件的真实前缀（文件被改过 / 记录不全），清掉重传
			ChannelSendData(model.ClientFileReceive{Op: op, FileOp: "reset", RemotePath: dst}, dc)
			offset = 0
			h.Reset()
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return err
			}
		}
	}

	if offset >= size {
		// 对端内容与源文件完全一致，直接走校验确认
		if err := ChannelSendData(model.ClientFileReceive{
			Op: op, FileOp: "done", FilePath: src, FileSize: size, RemotePath: dst, Crc: h.Sum32(),
		}, dc); err != nil {
			return err
		}
		return waitVerify(dst)
	}

	buf := make([]byte, chunkSize)
	for {
		if isStopped(dst) {
			return ErrStopped
		}
		n, err := f.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
			chunk := model.ClientFileReceive{
				Op: op, FileOp: "progress", FilePath: src,
				FileSize: size, RemotePath: dst,
				FileContent: buf[:n], Offset: int(offset),
			}
			b, mErr := json.Marshal(chunk)
			if mErr != nil {
				return mErr
			}
			if err := sendChunk(dc, b); err != nil {
				return err
			}
			offset += int64(n)
			if onBytes != nil {
				onBytes(n)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	if err := ChannelSendData(model.ClientFileReceive{
		Op: op, FileOp: "done", FilePath: src, FileSize: size, RemotePath: dst, Crc: h.Sum32(),
	}, dc); err != nil {
		return err
	}
	return waitVerify(dst)
}

// sendChunk 发送一个分片。SCTP 缓冲区满时短暂等待重试；通道断开立即失败——
// 以前忽略 Send 的错误，连接断了还会把整个文件"读完"并报成功，导致看起来传完了其实没传。
func sendChunk(dc *webrtc.DataChannel, b []byte) error {
	waited := time.Duration(0)
	for {
		if dc.ReadyState() != webrtc.DataChannelStateOpen {
			return errors.New("数据通道已断开")
		}
		if dc.BufferedAmount() > BUF_WAITING_THRESHOLD {
			if waited > maxBackpressureWait {
				return errors.New("对端接收过慢，发送超时")
			}
			time.Sleep(backpressureInterval)
			waited += backpressureInterval
			continue
		}
		if err := dc.Send(b); err == nil {
			return nil
		} else if waited > maxBackpressureWait {
			return fmt.Errorf("发送超时: %w", err)
		}
		time.Sleep(backpressureInterval)
		waited += backpressureInterval
	}
}

func waitVerify(dst string) error {
	msg, ok := waitAck(dst, "verify", verifyTimeout)
	if !ok {
		// 对端没回执（超时/旧版本），不做最终校验
		fmt.Println("未收到校验回执:", dst)
		return nil
	}
	if msg.Status != 0 {
		return errors.New("内容校验不通过")
	}
	return nil
}

// reportStatus upload 的进度回给控制端 UI；download 的进度要走数据通道回对端
func reportStatus(res model.ClientFileReceive, progress, status int, speed, errMsg string, FileConn *swebsocket.ServerConn, dc *webrtc.DataChannel) {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	if res.Op == "upload" {
		if FileConn != nil {
			FileConn.Send <- model.ClientFileResponse{
				Op:         res.Op,
				Speed:      speed,
				FileName:   filepath.Base(res.FilePath),
				FilePath:   res.FilePath,
				FileSize:   FormatFileSize(uint64(res.FileSize)),
				RemotePath: res.RemotePath,
				Progress:   progress,
				Status:     status,
				ErrMsg:     errMsg,
			}
		}
		return
	}
	ChannelSendData(model.ClientFileReceive{
		Op: res.Op, FileOp: "status", FilePath: res.FilePath,
		FileSize: res.FileSize, RemotePath: res.RemotePath,
		Progress: progress, Speed: speed, Status: status, ErrMsg: errMsg,
	}, dc)
}

func percent(done, total int64) int {
	if total <= 0 {
		return 0
	}
	p := int(float64(done) / float64(total) * 100)
	if p > 99 {
		p = 99
	}
	return p
}

type dirFile struct {
	path string
	dest string
	size int64
}

// collectDir 递归收集目录下的子目录与文件（相对路径拼接目标路径）
func collectDir(root, dest string) (dirs []string, files []dirFile, total int64) {
	filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil || info == nil || p == root {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return nil
		}
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			dirs = append(dirs, target)
			return nil
		}
		files = append(files, dirFile{path: p, dest: target, size: info.Size()})
		total += info.Size()
		return nil
	})
	return
}

// prepareFile 接收前的准备：文件已存在且不完整就返回已有字节数（续传），
// 比预期大就清空重来，不存在就建好。
func prepareFile(path string, size int64) int {
	if info, err := os.Stat(path); err == nil {
		switch {
		case info.Size() == size:
			return int(size)
		case info.Size() < size:
			return int(info.Size())
		default:
			os.Truncate(path, 0)
			return 0
		}
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	if f, err := os.Create(path); err == nil {
		f.Close()
	}
	return 0
}

// SenderHandlerFile 接收端：建目录 / 回执续传偏移与已有内容 CRC / 按偏移落盘 / 整文件校验
func SenderHandlerFile(msg model.ClientFileReceive, dataChannel *webrtc.DataChannel, FileConn *swebsocket.ServerConn) {
	switch msg.FileOp {
	case "mkdir":
		if err := os.MkdirAll(msg.RemotePath, 0755); err != nil {
			fmt.Println("建目录失败", msg.RemotePath, err)
		}
	case "start":
		offset := prepareFile(msg.RemotePath, msg.FileSize)
		crc, _ := fileCrc(msg.RemotePath, 0)
		fmt.Println("接收", msg.RemotePath, "从偏移", offset, "开始")
		ChannelSendData(model.ClientFileReceive{
			Op: msg.Op, FileOp: "start_ack",
			RemotePath: msg.RemotePath, Offset: offset, FileSize: msg.FileSize, Crc: crc,
		}, dataChannel)
	case "reset":
		os.MkdirAll(filepath.Dir(msg.RemotePath), 0755)
		if f, err := os.Create(msg.RemotePath); err == nil {
			f.Close()
		}
	case "progress":
		f, err := os.OpenFile(msg.RemotePath, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			fmt.Println("写入失败", msg.RemotePath, err)
			return
		}
		f.WriteAt(msg.FileContent, int64(msg.Offset))
		f.Close()
	case "done":
		crc, err := fileCrc(msg.RemotePath, 0)
		info, statErr := os.Stat(msg.RemotePath)
		status := 0
		if err != nil || statErr != nil || info.Size() != msg.FileSize || crc != msg.Crc {
			status = 3
			fmt.Println("校验不通过", msg.RemotePath, "本地", crc, "期望", msg.Crc, msg.FileSize)
		}
		ChannelSendData(model.ClientFileReceive{
			Op: msg.Op, FileOp: "verify", RemotePath: msg.RemotePath,
			FileSize: msg.FileSize, Crc: crc, Status: status,
		}, dataChannel)
	case "status":
		if FileConn != nil {
			FileConn.Send <- model.ClientFileResponse{
				Op:         msg.Op,
				Speed:      msg.Speed,
				FileName:   filepath.Base(msg.FilePath),
				FilePath:   msg.FilePath,
				FileSize:   FormatFileSize(uint64(msg.FileSize)),
				RemotePath: msg.RemotePath,
				Progress:   msg.Progress,
				Status:     msg.Status,
				ErrMsg:     msg.ErrMsg,
			}
		}
	}
}

func ChannelSendData(data any, dataChannel *webrtc.DataChannel) error {
	if dataChannel == nil {
		return errors.New("数据通道未就绪")
	}
	v, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return dataChannel.Send(v)
}
