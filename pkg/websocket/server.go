// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"net"
	"net/http"
	"time"
)

const (
	// ServerName 服务名
	ServerName = "websocket"
)

// ServerParameter WebSocket服务构造参数
type ServerParameter struct {
	// WebSocketBindAddress WebSocket服务绑定地址
	WebSocketBindAddress string
}

// Server WebSocket服务
type Server struct {
	// ServerParameter WebSocket服务构造参数
	ServerParameter
	// serverHandler Server回调
	serverHandler define.ServerHandler
	// closer 安全退出锁
	closer *util.SafeCloser
	// upgrader upgrader WebSocket upgrade参数
	upgrader *websocket.Upgrader
	// listener HTTP侦听对象
	listener net.Listener
}

// NewServer 新建一个WebSocket服务实例
func NewServer(params *ServerParameter, serverHandler define.ServerHandler) (srv *Server, err error) {
	glog.Infoln("websocket::NewServer")
	srv = &Server{
		ServerParameter: *params,
		serverHandler:   serverHandler,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	return srv, err
}

// Run 启动WebSocket服务
func (srv *Server) Run(closer *util.SafeCloser) (err error) {
	glog.Infoln("websocket::Server::Run()")
	srv.closer = closer
	http.HandleFunc("/ws", srv.Handle)
	srv.listener, err = net.Listen("tcp4", srv.WebSocketBindAddress)
	if err != nil {
		glog.Errorf("websocket::Server::Run() listen(%s) error: %s\n",
			srv.WebSocketBindAddress, err)
		return
	}
	var httpErr error
	go func() {
		httpErr = http.Serve(srv.listener, nil)
	}()
	time.Sleep(time.Second)
	if httpErr != nil {
		glog.Errorf("websocket::Server::Run() http.Server(%s) error: %s\n",
			srv.WebSocketBindAddress, err)
		return httpErr
	}
	err = srv.closer.Add(ServerName, func() {
		glog.Warningln("websocket::Server::Run() to close")
		srv.listener.Close()
	})

	return err
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("websocket::Server::Close()")
	defer srv.closer.Done(ServerName)
	// 关闭HTTP服务
	if srv.listener != nil {
		err = srv.listener.Close()
	}
	return err
}

// Handle 处理HTTP链接
func (srv *Server) Handle(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("websocket::Server::Handle()")
	if srv.closer.IsClose() {
		return
	}
	// Upgrade到WebSocket连接
	c, err := srv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("websocket::Server::Handle() upgrade error: %s\n", err)
		return
	}

	// 新建连接
	conn := NewConnection(c)
	srv.serverHandler.OnNewConnection(conn)

	var cmd *protocol.Command
FOR_LOOP:
	for !srv.closer.IsClose() {
		// 读取Command
		cmd, err = conn.ReadCommand()
		if err != nil {
			if err == define.ErrConnectionClosed {
				glog.Infoln("websocket::Server::Handle() connection to close")
				conn.Close(false)
			}
			break FOR_LOOP
		}
		srv.serverHandler.OnReceivedCommand(conn, cmd)
	}
	glog.Infoln("websocket::Server::Handle() ", conn, " closed")
	srv.serverHandler.OnCloseConnection(conn)
}
