// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package gateway

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"github.com/zhangpeihao/zim/pkg/websocket"
	"sync"
	"time"
)

const (
	// ServerName 服务名
	ServerName = "gateway"
)

// ServerParameter 网关服务参数
type ServerParameter struct {
	websocket.ServerParameter
	// Key 验证密钥
	Key protocol.Key
}

// Server 网关服务
type Server struct {
	// ServerParameter 服务参数
	ServerParameter
	closer *util.SafeCloser
	sync.Mutex
	wsServer    define.SubServer
	connections map[define.Connection]struct{}
}

// NewServer 新建服务
func NewServer(params *ServerParameter) (srv *Server, err error) {
	glog.Info("gateway::NewServer()")
	srv = &Server{
		ServerParameter: *params,
		connections:     make(map[define.Connection]struct{}),
	}
	srv.wsServer, err = websocket.NewServer(&websocket.ServerParameter{
		WebSocketBindAddress: srv.WebSocketBindAddress,
	}, srv)
	return
}

// Run 启动WebSocket服务
func (srv *Server) Run(closer *util.SafeCloser) (err error) {
	glog.Info("gateway::Server::Run()")
	srv.closer = closer
	if err = srv.wsServer.Run(closer); err != nil {
		glog.Errorln("gateway::Server::Run() error:", err)
		return err
	}
	err = srv.closer.Add(ServerName, func() {
		glog.Infoln("gateway::Server::Run() to close")
	})
	return
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Info("gateway::Server::Close()")
	defer srv.closer.Done(ServerName)
	// 关闭WebSocket服务
	if err = srv.wsServer.Close(timeout); err != nil {
		glog.Errorln("gateway::Server::Close() websocket server close error:", err)
	}
	// 关闭所有链接
	srv.Lock()
	connections := make([]define.Connection, len(srv.connections))
	i := 0
	for conn := range srv.connections {
		connections[i] = conn
	}
	srv.Unlock()
	for _, conn := range connections {
		conn.Close(true)
	}
	return err
}

// OnNewConnection 连接新建处理
func (srv *Server) OnNewConnection(conn define.Connection) {
	glog.Info("gateway::Server::OnNewConnection()")
	srv.Lock()
	defer srv.Unlock()
	srv.connections[conn] = struct{}{}
}

// OnCloseConnection 连接关闭处理
func (srv *Server) OnCloseConnection(conn define.Connection) {
	glog.Info("gateway::Server::OnCloseConnection()")
	srv.Lock()
	defer srv.Unlock()
	delete(srv.connections, conn)
}

// OnReceivedCommand 收到命令
func (srv *Server) OnReceivedCommand(conn define.Connection, command *protocol.Command) {
	glog.Infof("gateway::Server::OnReceivedCommand() command %s from %s\n", command.Name, conn)
}
