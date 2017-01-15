// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package gateway

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/app"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/push/driver/httpserver"
	"github.com/zhangpeihao/zim/pkg/util"
	"github.com/zhangpeihao/zim/pkg/websocket"
	"strings"
	"sync"
	"time"
)

const (
	// ServerName 服务名
	ServerName = "gateway"
)

// ServerParameter 网关服务参数
type ServerParameter struct {
	websocket.WSParameter
	httpserver.PushHTTPServerParameter
	// Key 验证密钥
	Key protocol.Key
	// AppConfigs 应用配置
	AppConfigs []string
}

// Server 网关服务
type Server struct {
	// ServerParameter 服务参数
	ServerParameter
	// closer 安全退出
	closer *util.SafeCloser
	// 锁
	sync.Mutex
	// wsServer WebSocket服务
	wsServer define.SubServer
	// connections 连接Map
	connections map[string][]define.Connection
	// pushServer Push服务
	pushServer *httpserver.Server
	// apps 应用Map
	apps map[string]*app.App
}

// NewServer 新建服务
func NewServer(params *ServerParameter) (srv *Server, err error) {
	glog.Infoln("gateway::NewServer()")
	srv = &Server{
		ServerParameter: *params,
		connections:     make(map[string][]define.Connection),
		apps:            make(map[string]*app.App),
	}
	srv.wsServer, err = websocket.NewServer(&srv.ServerParameter.WSParameter, srv)
	if err != nil {
		return nil, err
	}
	srv.pushServer, err = httpserver.NewServer(&srv.ServerParameter.PushHTTPServerParameter, srv)
	if err != nil {
		return nil, err
	}
	glog.Infof("srv.AppConfigs: %+v\n", srv.AppConfigs)
	for _, config := range srv.AppConfigs {
		appConfig, err := app.NewApp(config)
		if err != nil {
			return nil, err
		}
		srv.apps[appConfig.Name] = appConfig
	}
	return
}

// Run 启动WebSocket服务
func (srv *Server) Run(closer *util.SafeCloser) (err error) {
	glog.Infoln("gateway::Server::Run()")
	srv.closer = closer
	if err = srv.wsServer.Run(closer); err != nil {
		glog.Errorln("gateway::Server::Run() wsServer error:", err)
		return err
	}
	if err = srv.pushServer.Run(closer); err != nil {
		glog.Errorln("gateway::Server::Run() pushServer run error:", err)
		return err
	}
	err = srv.closer.Add(ServerName, func() {
		glog.Infoln("gateway::Server::Run() to close")
	})
	return
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("gateway::Server::Close()")
	defer srv.closer.Done(ServerName)
	// 关闭WebSocket服务
	if err = srv.wsServer.Close(timeout); err != nil {
		glog.Errorln("gateway::Server::Close() websocket server close error:", err)
	}
	// 关闭所有链接
	srv.Lock()
	var connections []define.Connection
	for _, conn := range srv.connections {
		connections = append(connections, conn...)
	}
	srv.Unlock()
	for _, conn := range connections {
		conn.Close(true)
	}
	return err
}

// OnNewConnection 连接新建处理
func (srv *Server) OnNewConnection(conn define.Connection) {
	glog.Infoln("gateway::Server::OnNewConnection()")
	// Todo: Login timeout
}

// OnCloseConnection 连接关闭处理
func (srv *Server) OnCloseConnection(conn define.Connection) {
	glog.Infoln("gateway::Server::OnCloseConnection()")
	srv.Lock()
	defer srv.Unlock()
	delete(srv.connections, conn.ID())
}

// OnReceivedCommand 收到命令
func (srv *Server) OnReceivedCommand(conn define.Connection, command *protocol.Command) (err error) {
	glog.Infof("gateway::Server::OnReceivedCommand() command %s from %s\n", command.Name, conn)
	var (
		loginCmd *protocol.GatewayLoginCommand
		ok       bool
	)
	appConfig, ok := srv.apps[command.AppID]
	if !ok {
		glog.Warningln("gateway::Server::OnReceivedCommand() No application found",
			command.AppID)
		conn.Close(false)
		return define.ErrKnownApp
	}
	// 检查登入
	if command.Name != protocol.Login {
		if !conn.IsLogin() {
			glog.Warningln("gateway::Server::OnReceivedCommand() first command must be login! got:",
				command.Name)
			conn.Close(false)
			return define.ErrUnsupportProtocol
		}
	} else {
		loginCmd, ok = command.Data.(*protocol.GatewayLoginCommand)
		if !ok {
			glog.Warningf("gateway::Server::OnReceivedCommand() invoke (%s) error %s\n",
				command.Name, err)
			return define.ErrNeedAuth
		}
		glog.Infof("gateway::Server::OnReceivedCommand() login: %+v\n", loginCmd)
	}

	if len(loginCmd.UserID) == 0 {
		glog.Warningf("gateway::Server::OnReceivedCommand() login userID is empty\n")
		return define.ErrAuthFailed
	}

	// Route
	ink := appConfig.Router.Find(command.Name)

	if ink == nil {
		glog.Warningf("gateway::Server::OnReceivedCommand() no route to %s\n", command.Name)
		return
	}

	resp, err := ink.Invoke(loginCmd.UserID, command)
	if err != nil {
		glog.Warningf("gateway::Server::OnReceivedCommand() invoke (%s) error %s\n",
			command.Name, err)
		return
	}

	if loginCmd != nil {
		// Todo: Parse response
		conn.LoginSuccess(command.AppID, loginCmd.UserID, loginCmd.DeviceID)
		connid := conn.ID()
		srv.Lock()
		connections, find := srv.connections[connid]
		var oldConn define.Connection
		if find {
			var index int
			for index, oldConn = range connections {
				if oldConn.DeviceID() == conn.DeviceID() {
					glog.Warningf("gateway::Server::OnReceivedCommand() replace connection ID: %s\n", connid)
					connections[index] = conn
				} else {
					oldConn = nil
				}
			}
		}
		if oldConn == nil {
			connections = append(connections, conn)
			srv.connections[connid] = connections
		}
		srv.Unlock()
		if oldConn != nil {
			// 在锁外关闭链接，防止死锁
			oldConn.Close(false)
		}
	}

	if resp == nil {
		return
	}

	glog.Infof("gateway::Server::OnReceivedCommand() invoke(%s) response %s",
		command.Name, resp)
	if resp.Name == protocol.Push2User {
		srv.OnPushToUser(resp)
	}
	return
}

// OnPushToUser 推送消息给用户
func (srv *Server) OnPushToUser(cmd *protocol.Command) {
	glog.Infof("gateway::Server::OnPushToUser()\n")
	var (
		pushCmd *protocol.Push2UserCommand
		ok      bool
	)
	pushCmd, ok = cmd.Data.(*protocol.Push2UserCommand)
	if !ok {
		glog.Warningln("gateway::Server::OnPushToUser() parse result error")
		return
	}

	touser := cmd.Copy()
	touser.Data = nil
	srv.Lock()
	for _, id := range strings.Split(pushCmd.UserIDList, ",") {
		connections, ok := srv.connections[define.ConnectionID(cmd.AppID, id)]
		if ok {
			for _, conn := range connections {
				conn.Send(touser)
			}
		} else {
			glog.Warningf("gateway::Server::OnPushToUser() not find connection")
		}
	}
	srv.Unlock()
}
