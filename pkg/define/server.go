// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

import (
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
)

// Server 服务接口
type Server interface {
	// SafeCloseServer 安全关闭接口
	util.SafeCloseServer
}

// SubServer 子服务接口
type SubServer interface {
	// SafeCloseServer 安全关闭接口
	util.SafeCloseServer
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
