// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

import "github.com/zhangpeihao/zim/pkg/protocol"

// Connection 连接接口
type Connection interface {
	LoginSuccess()
	IsLogin() bool
	Close(force bool) error
	String() string
	Send(cmd *protocol.Command) error
}
