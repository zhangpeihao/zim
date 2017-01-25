// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package cmd

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Root 参数
	cfgFile string

	// stress client 参数
	cfgWebSocketURL     string
	cfgAppID            string
	cfgKey              string
	cfgNumber           uint
	cfgBase             uint
	cfgInterval         uint
	gInterval           time.Duration
	cfgInfluxdbAddress  string
	cfgInfluxdbDB       string
	cfgInfluxdbUser     string
	cfgInfluxdbPassword string
	cfgInfluxdbInterval uint

	// stub 参数
	cfgStubBindAddress string

	// 计数器
	gErrorCounter      int32
	gReceiveCounter    int32
	gSendCounter       int32
	gCheckErrorCounter int32
	gPreTime           int64

	// 程序控制变量
	gCloseGate = new(sync.WaitGroup)
	gExit      int32
)

// SetExitFlag 设置退出标志
func SetExitFlag() {
	atomic.StoreInt32(&gExit, 1)
}

// IsExit 判断是否程序是否已经处于退出状态
func IsExit() bool {
	return atomic.LoadInt32(&gExit) != 0
}

// CountError 错误计数加1
func CountError() {
	atomic.AddInt32(&gErrorCounter, 1)
}

// CountReceive 接收计数加1
func CountReceive() {
	atomic.AddInt32(&gReceiveCounter, 1)
}

// CountSend 发送计数加1
func CountSend() {
	atomic.AddInt32(&gSendCounter, 1)
}

// CountCheckError 消息校验错误计数加1
func CountCheckError() {
	atomic.AddInt32(&gCheckErrorCounter, 1)
}
