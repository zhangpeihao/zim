// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package invoker

import "github.com/zhangpeihao/zim/pkg/protocol"

// CallbackFunc 回调
type CallbackFunc func(req *protocol.Command, resp *protocol.Command)

// Invoker 调用接口
type Invoker interface {
	// Invoke 同步调用
	Invoke(string, *protocol.Command) (*protocol.Command, error)
	// String 输出
	String() string
}
