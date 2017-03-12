// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

import (
	"context"
	"time"

	"github.com/zhangpeihao/zim/pkg/protocol"
)

// Server 服务接口
type Server interface {
	// Run 运行
	Run(context.Context) error
	// Close 关闭
	Close(timeout time.Duration) (err error)
}

// ServerHandler 服务回调接口
type ServerHandler interface {
	// OnNewConnection 当有新连接建立
	OnNewConnection(conn Connection)
	// OnCloseConnection 当有连接关闭
	OnCloseConnection(conn Connection)
	// OnReceivedCommand 当收到命令
	OnReceivedCommand(conn Connection, command *protocol.Command) error
}
