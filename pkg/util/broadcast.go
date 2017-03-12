// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"sync"

	"github.com/golang/glog"
)

// BroadcastSignal 广播信令
type BroadcastSignal struct {
	sync.Mutex
	signals []chan struct{}
	gate    sync.WaitGroup
}

// NewBroadcastSignal 新建
func NewBroadcastSignal() *BroadcastSignal {
	return &BroadcastSignal{
		signals: make([]chan struct{}, 0),
	}
}

// Wait 等待信号
func (b *BroadcastSignal) Wait() chan struct{} {
	glog.Infoln("util::BroadcastSignal::Wait()")
	defer glog.Infoln("util::BroadcastSignal::Wait() done")
	signal := make(chan struct{}, 0)
	b.Lock()
	b.signals = append(b.signals, signal)
	b.Unlock()
	return signal
}

// RemoveWait 取消等待
func (b *BroadcastSignal) RemoveWait(signal chan struct{}) {
	glog.Infoln("util::BroadcastSignal::RemoveWait()")
	defer glog.Infoln("util::BroadcastSignal::RemoveWait() done")
	b.Lock()
	for index, s := range b.signals {
		if s == signal {
			b.signals = append(b.signals[:index], b.signals[index+1:]...)
			break
		}
	}
	b.Unlock()
}

// Broadcast 广播
func (b *BroadcastSignal) Broadcast() {
	glog.Infoln("util::BroadcastSignal::Broadcast()")
	defer glog.Infoln("util::BroadcastSignal::Broadcast() done")
	b.gate.Add(1)
	b.Lock()
	for _, signal := range b.signals {
		b.gate.Add(1)
		go func(ch chan struct{}) {
			glog.Infoln("util::BroadcastSignal::Broadcast() send signal")
			defer glog.Infoln("util::BroadcastSignal::Broadcast() send signal done")
			ch <- struct{}{}
			b.gate.Done()
		}(signal)
	}
	b.Unlock()
	b.gate.Done()
}

// Close 关闭
func (b *BroadcastSignal) Close() {
	glog.Infoln("util::BroadcastSignal::Close()")
	defer glog.Infoln("util::BroadcastSignal::Close() done")
	b.gate.Wait()
	b.Lock()
	for _, signal := range b.signals {
		close(signal)
	}
	b.Unlock()
	//	time.Sleep(time.Second)
}

// Empty 是否为空
func (b *BroadcastSignal) Empty() bool {
	b.Lock()
	defer b.Unlock()
	return len(b.signals) == 0
}
