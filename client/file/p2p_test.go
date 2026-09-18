package file

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"remote/model"

	"github.com/pion/webrtc/v4"
)

// loopbackChannel 建一对进程内直连的 PeerConnection / 数据通道（走本机 host candidate）
func loopbackChannel(t *testing.T) (sender, receiver *webrtc.DataChannel, cleanup func()) {
	t.Helper()

	a, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("建发送端 PeerConnection 失败: %v", err)
	}
	b, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatalf("建接收端 PeerConnection 失败: %v", err)
	}

	dcA, err := a.CreateDataChannel("file", nil)
	if err != nil {
		t.Fatalf("创建数据通道失败: %v", err)
	}

	incoming := make(chan *webrtc.DataChannel, 1)
	b.OnDataChannel(func(d *webrtc.DataChannel) { incoming <- d })

	a.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			b.AddICECandidate(c.ToJSON())
		}
	})
	b.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			a.AddICECandidate(c.ToJSON())
		}
	})

	offer, err := a.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	if err := b.SetRemoteDescription(offer); err != nil {
		t.Fatal(err)
	}
	answer, err := b.CreateAnswer(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := b.SetLocalDescription(answer); err != nil {
		t.Fatal(err)
	}
	if err := a.SetRemoteDescription(answer); err != nil {
		t.Fatal(err)
	}

	var dcB *webrtc.DataChannel
	select {
	case dcB = <-incoming:
	case <-time.After(15 * time.Second):
		t.Fatal("接收端数据通道未建立")
	}

	openA := make(chan struct{})
	openB := make(chan struct{})
	dcA.OnOpen(func() { close(openA) })
	dcB.OnOpen(func() { close(openB) })

	// 发送端回执路由（start_ack / verify 必须交给等待方）
	dcA.OnMessage(func(m webrtc.DataChannelMessage) {
		var msg model.ClientFileReceive
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			return
		}
		if IsAck(msg.FileOp) {
			Ack(msg)
		}
	})

	for name, ch := range map[string]chan struct{}{"发送端": openA, "接收端": openB} {
		select {
		case <-ch:
		case <-time.After(15 * time.Second):
			t.Fatalf("%s数据通道未打开", name)
		}
	}

	return dcA, dcB, func() {
		a.Close()
		b.Close()
	}
}

// countChunks 把接收端挂到 dcB 上，返回已收分片数的读取器
func countChunks(dcB *webrtc.DataChannel) (func() int, <-chan struct{}) {
	var (
		mu     sync.Mutex
		chunks int
	)
	done := make(chan struct{})
	var once sync.Once

	dcB.OnMessage(func(m webrtc.DataChannelMessage) {
		var msg model.ClientFileReceive
		if err := json.Unmarshal(m.Data, &msg); err != nil {
			return
		}
		if IsAck(msg.FileOp) {
			Ack(msg)
			return
		}
		if msg.FileOp == "progress" {
			mu.Lock()
			chunks++
			mu.Unlock()
		}
		SenderHandlerFile(msg, dcB, nil)
		if msg.FileOp == "done" {
			once.Do(func() { close(done) })
		}
	})

	return func() int {
		mu.Lock()
		defer mu.Unlock()
		return chunks
	}, done
}

// runTransfer 走一遍完整的 P2P 传输，返回接收端实际收到的分片数
func runTransfer(t *testing.T, src, dst string, want []byte) int {
	t.Helper()

	dcA, dcB, cleanup := loopbackChannel(t)
	defer cleanup()

	SetSender(dcA, nil)
	defer ClearSender(dcA)
	StartWorker()

	getChunks, done := countChunks(dcB)

	FileList <- model.ClientFileReceive{
		Op: "download", FilePath: src, RemotePath: dst, FileSize: int64(len(want)),
	}

	select {
	case <-done:
	case <-time.After(60 * time.Second):
		t.Fatal("传输超时")
	}
	// 发送端还在等 verify 回执；留点时间让对端校验并回执
	time.Sleep(300 * time.Millisecond)
	return getChunks()
}

func TestP2PTransferAndResume(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "big.bin")
	dst := filepath.Join(dstDir, "big.bin")

	// 200KB → 20 个分片
	data := make([]byte, 200*1024)
	for i := range data {
		data[i] = byte(i % 251)
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatal(err)
	}

	// 第一次：完整传一遍
	full := runTransfer(t, src, dst, data)
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("首次传输内容不一致（收到 %d 字节，期望 %d）", len(got), len(data))
	}

	// 模拟中断：把对端文件截断到一半
	half := len(data) / 2
	if err := os.Truncate(dst, int64(half)); err != nil {
		t.Fatal(err)
	}

	// 第二次：应当从断点续传，而不是从头再来
	resumed := runTransfer(t, src, dst, data)
	got, err = os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("续传后内容不一致（收到 %d 字节，期望 %d）", len(got), len(data))
	}

	fullChunks := len(data)/chunkSize + 1
	if full <= 0 {
		t.Fatal("首次传输没有统计到分片")
	}
	// 续传只需发后半段，分片数应明显少于完整传输
	if resumed >= full*3/4 {
		t.Fatalf("续传发送了 %d 个分片，完整传输是 %d 个 —— 看起来是从头重传而不是续传", resumed, full)
	}
	t.Logf("完整传输 %d 个分片，续传只发了 %d 个（预期约 %d）", full, resumed, fullChunks/2)
}

// 回归：会话切换时任务必须走"当前"通道。
// 之前每个会话各起一个消费者并按会话重建 FileList，旧消费者会抢走新任务、
// 发往已死的通道，UI 侧状态永远停在"等待传输"。
func TestWorkerFollowsCurrentSession(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "a.bin")
	dst := filepath.Join(dstDir, "a.bin")

	data := make([]byte, 50*1024)
	for i := range data {
		data[i] = byte(i % 89)
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatal(err)
	}

	// 旧会话：故意不给它挂接收端。任务若被它抢走就会石沉大海。
	oldA, _, closeOld := loopbackChannel(t)
	SetSender(oldA, nil)
	StartWorker()

	// 新会话建立（此时旧会话可能还没收尾）
	newA, newB, closeNew := loopbackChannel(t)
	defer closeNew()
	SetSender(newA, nil)

	// 旧会话收尾：不能把新会话的通道清掉
	ClearSender(oldA)
	if dc, _ := sender(); dc != newA {
		t.Fatal("旧会话收尾把新会话的通道清掉了")
	}
	closeOld()

	// 任务必须走新通道
	getChunks, done := countChunks(newB)
	FileList <- model.ClientFileReceive{
		Op: "download", FilePath: src, RemotePath: dst, FileSize: int64(len(data)),
	}

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		t.Fatal("任务没走当前会话的通道（很可能被旧会话残留的消费者抢走了）")
	}
	if getChunks() == 0 {
		t.Fatal("接收端没有收到任何分片")
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Fatal("内容不一致")
	}
	ClearSender(newA)
}

func TestP2PTransferDetectsChangedSource(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	src := filepath.Join(srcDir, "a.bin")
	dst := filepath.Join(dstDir, "a.bin")

	data := make([]byte, 100*1024)
	for i := range data {
		data[i] = byte(i % 97)
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(dst, data[:50*1024], 0644); err != nil {
		t.Fatal(err)
	}

	// 源文件前一半被改过（总大小不变）——续传必须检测出并从头重传
	altered := append([]byte(nil), data...)
	altered[100] ^= 0xFF
	altered[1024] ^= 0xFF
	if err := os.WriteFile(src, altered, 0644); err != nil {
		t.Fatal(err)
	}

	runTransfer(t, src, dst, altered)

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, altered) {
		t.Fatal("源文件被改过时应从头重传，但结果与源文件不一致（说明错误地复用了旧前缀）")
	}
}
