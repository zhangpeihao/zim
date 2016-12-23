// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	// ErrCloseTimeout 关闭超时
	ErrCloseTimeout = errors.New("close timeout")
	// terminationSignals 退出信号量.
	terminationSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}
)

// CloseFunc 关闭函数
type CloseFunc func()

// SafeCloseServer 安全退出服务接口
type SafeCloseServer interface {
	// Run 运行
	Run(closer *SafeCloser) error
	// Close 关闭
	Close(timeout time.Duration) error
}

// SafeCloser 安全退出控制开关
type SafeCloser struct {
	sync.Mutex
	closeFlag            int32
	gate                 *sync.WaitGroup
	signals              map[string]chan struct{}
	terminationSignalsCh chan os.Signal
}

// NewSafeCloser 新建安全退出控制开关
func NewSafeCloser() *SafeCloser {
	return &SafeCloser{
		gate:    new(sync.WaitGroup),
		signals: make(map[string]chan struct{}),
	}
}

// Add 添加控制项，返回关闭信号
func (sc *SafeCloser) Add(name string, closeFunc func()) error {
	glog.Infof("util::SafeClose::Add(%s)\n", name)
	sc.Lock()
	defer sc.Unlock()
	if _, found := sc.signals[name]; found {
		return fmt.Errorf("safecloser %s existed", name)
	}
	sc.gate.Add(1)
	ch := make(chan struct{})
	sc.signals[name] = ch

	// 侦听退出信号
	go func() {
		<-ch
		glog.Warningln("server \"" + name + "\" to close")
		closeFunc()
	}()

	return nil
}

// Done 控制项已安全退出
func (sc *SafeCloser) Done(name string) error {
	glog.Infof("util::SafeClose::Done(%s)\n", name)
	sc.Lock()
	defer sc.Unlock()
	ch, found := sc.signals[name]
	if !found {
		return fmt.Errorf("safecloser %s not existed", name)
	}
	close(ch)
	delete(sc.signals, name)
	sc.gate.Done()
	glog.Infof("util::SafeClose::Done() %s Done\n", name)
	return nil
}

// WaitAndClose 开始安全退出
func (sc *SafeCloser) WaitAndClose(timeout time.Duration, fn CloseFunc) error {
	atomic.StoreInt32(&sc.closeFlag, 1)
	sc.terminationSignalsCh = make(chan os.Signal, 1)
	signal.Notify(sc.terminationSignalsCh, terminationSignals...)
	defer func() {
		signal.Stop(sc.terminationSignalsCh)
		close(sc.terminationSignalsCh)
	}()
	<-sc.terminationSignalsCh
	fn()

	return sc.Close(timeout)
}

// Close 开始安全退出
func (sc *SafeCloser) Close(timeout time.Duration) (err error) {
	atomic.StoreInt32(&sc.closeFlag, 1)
	// 异步向所有控制项发送退出指令
	go func() {
		sc.Lock()
		chs := make([]chan<- struct{}, len(sc.signals))
		i := 0
		for _, ch := range sc.signals {
			chs[i] = ch
			i++
		}
		sc.Unlock()
		for _, ch := range chs {
			ch <- struct{}{}
		}
	}()

	if timeout > 0 {
		// 异步等待所有控制项退出完成
		closedCh := make(chan struct{})
		go func() {
			sc.gate.Wait()
			closedCh <- struct{}{}
		}()

		// 处理超时
		select {
		case <-closedCh:
			glog.Infoln("util::SafeClose::Close() safe closed")
		case <-time.After(timeout):
			glog.Errorln("util::SafeClose::Close() timeout")
			err = ErrCloseTimeout
		}
	} else {
		// 等待所有控制项退出完成
		sc.gate.Wait()
	}
	return err
}

// IsClose 是否关闭
func (sc *SafeCloser) IsClose() bool {
	closeFlag := atomic.LoadInt32(&sc.closeFlag)
	return closeFlag != 0
}
