package p2p

import (
	"encoding/json"
	"fmt"

	"remote/input"
	"remote/model"

	"github.com/pion/webrtc/v4"
)

// KeyboardPeer 键盘数据通道的收发端。
// 之前 InitReceiveKeyboardP2P / InitSenderKeyboardP2P 两个 76 行的函数几乎完全重复，
// 这里把公共部分收成 NewKeyboardPeer，只保留"谁建通道、谁收通道"的差异。
type KeyboardPeer struct {
	PC       *webrtc.PeerConnection
	DC       *webrtc.DataChannel
	onClosed func()
}

func NewKeyboardPeer(ices []model.ICEServer, send SendFunc, onClosed func()) (*KeyboardPeer, error) {
	pc, err := NewPeer(ices)
	if err != nil {
		return nil, err
	}
	p := &KeyboardPeer{PC: pc, onClosed: onClosed}
	WatchState(pc, "keyboard", func() { p.Close() })
	OnCandidate(pc, send)
	return p, nil
}

// Receiver 被控端：等待控制端建好的数据通道，收到键鼠操作后注入本机。
// onChannel 用于让调用方拿到数据通道（例如挂到 global.KeyboardHandler）。
func (p *KeyboardPeer) Receiver(onChannel func(*webrtc.DataChannel)) {
	p.PC.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())
		p.DC = d
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open\n", d.Label(), d.ID())
		})
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			var op model.InputOperation
			if err := json.Unmarshal(msg.Data, &op); err != nil {
				fmt.Println("解析输入指令失败", err)
				return
			}
			if err := input.Operation(op); err != nil {
				fmt.Println("执行输入指令失败", err)
			}
		})
		d.OnClose(func() { fmt.Println("keyboard data channel closed") })
		if onChannel != nil {
			onChannel(d)
		}
	})
}

// Sender 控制端：创建数据通道并返回 offer，由调用方通过信令发出去
func (p *KeyboardPeer) Sender(label string, onClipboard func([]string)) (*webrtc.SessionDescription, error) {
	d, err := p.PC.CreateDataChannel(label, nil)
	if err != nil {
		return nil, err
	}
	p.DC = d
	d.OnOpen(func() { fmt.Println("控制端数据通道开启") })
	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		var c model.Clipboard
		if err := json.Unmarshal(msg.Data, &c); err != nil {
			return
		}
		if onClipboard != nil {
			onClipboard(c.ClipboardData)
			return
		}
		for _, s := range c.ClipboardData {
			input.WriteAll(s)
		}
	})
	d.OnClose(func() { fmt.Println("keyboard data channel closed") })

	return NewOffer(p.PC)
}

func (p *KeyboardPeer) Close() {
	if p.DC != nil {
		p.DC.Close()
		p.DC = nil
	}
	if p.PC != nil {
		p.PC.Close()
		p.PC = nil
	}
	if p.onClosed != nil {
		p.onClosed()
	}
}
