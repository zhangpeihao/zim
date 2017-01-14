// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"github.com/golang/glog"
	"runtime"
	"runtime/debug"
)

// RecoverFromPanic 捕获异常，输出日志
func RecoverFromPanic() {
	if r := recover(); r != nil {
		glog.Errorf("!!!Panic!!!\n%s", debug.Stack())
	}
}

// SetCPU 设置CPU运行数
func SetCPU(cpu int) {
	numcpu := runtime.NumCPU()
	currentcpu := runtime.GOMAXPROCS(0)
	if cpu <= 0 || cpu > numcpu-1 {
		cpu = numcpu - 1
	}
	if cpu > 1 && currentcpu != cpu {
		runtime.GOMAXPROCS(cpu)
	}
}
