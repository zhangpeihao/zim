// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"github.com/golang/glog"
	"runtime/debug"
)

// RecoverFromPanic 捕获异常，输出日志
func RecoverFromPanic() {
	if r := recover(); r != nil {
		glog.Errorf("!!!Panic!!!\n%s", debug.Stack())
	}
}
