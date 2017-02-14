// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package gateway

import (
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/zim/pkg/app"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/broker/register"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"github.com/zhangpeihao/zim/pkg/websocket"

	// 加载Broker
	_ "github.com/zhangpeihao/zim/pkg/broker/httpapi"
	_ "github.com/zhangpeihao/zim/pkg/broker/mock"
)

const (
	// ServerName 服务名
	ServerName = "gateway"
	// LoginTimeout 登入超时时间（单位：秒）
	LoginTimeout = 3600
)

// ServerParameter 网关服务参数
type ServerParameter struct {
	websocket.WSParameter
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
	// apps 应用Map
	apps map[string]*app.App
	// tag 消息队列tag
	tag string
}

// NewServer 新建服务
func NewServer() (srv *Server, err error) {
	glog.Infoln("gateway::NewServer()")
	srv = &Server{
		ServerParameter: ServerParameter{
			AppConfigs: viper.GetStringSlice("gateway.app-config"),
		},
		connections: make(map[string][]define.Connection),
		apps:        make(map[string]*app.App),
	}
	tag := viper.GetString("gateway.broker-tag")
	if len(tag) == 0 {
		srv.tag = ServerName
	} else {
		srv.tag = tag
	}
	srv.wsServer, err = websocket.NewServer(srv)
	if err != nil {
		return nil, err
	}
	if err = register.Init(srv, "gateway.broker"); err != nil {
		glog.Warningln("gateway::NewServer() register.Init() error:", err)
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
	if err = broker.Run(closer); err != nil {
		glog.Errorln("gateway::Server::Run() brocker.Run() error:", err)
		return err
	}
	if err = srv.wsServer.Run(closer); err != nil {
		glog.Errorln("gateway::Server::Run() wsServer error:", err)
		return err
	}
	err = srv.closer.Add(ServerName, func(timeout time.Duration) error {
		glog.Infoln("gateway::Server::Run() to close")
		return srv.Close(timeout)
	})
	return
}

// Close 退出
func (srv *Server) Close(timeout time.Duration) (err error) {
	glog.Infoln("gateway::Server::Close()")
	defer glog.Warningln("gateway::Server::Close() Done")
	defer srv.closer.Done(ServerName)
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
		resp     *protocol.Command
	)
	appConfig, ok := srv.apps[command.AppID]
	if !ok {
		glog.Warningln("gateway::Server::OnReceivedCommand() No application found",
			command.AppID)
		conn.Close(false)
		return define.ErrKnownApp
	}

	// Route
	broker := appConfig.Router.Find(command.Name)
	if broker == nil {
		glog.Warningf("gateway::Server::OnReceivedCommand() no route to %s\n", command.Name)
		return define.ErrAuthFailed
	}

	// 检查登入
	if !conn.IsLogin() {
		if command.Name != protocol.Login {
			glog.Warningln("gateway::Server::OnReceivedCommand() first command must be login! got:",
				command.Name)
			conn.Close(false)
			return define.ErrUnsupportProtocol
		}

		loginCmd, ok = command.Data.(*protocol.GatewayLoginCommand)
		if !ok {
			glog.Warningf("gateway::Server::OnReceivedCommand() invoke (%s) error %s\n",
				command.Name, err)
			return define.ErrNeedAuth
		}
		if strings.ToLower(appConfig.TokenCheck) == "yes" {
			now := time.Now().Unix()
			if loginCmd.Timestamp+LoginTimeout < now {
				glog.Warningf("gateway::Server::OnReceivedCommand() login timeout! loginCmd.Timestamp: %d, LoginTimeout: %d, now: %d\n",
					loginCmd.Timestamp, LoginTimeout, now)
				return define.ErrNeedAuth
			}
			token := loginCmd.CalToken(appConfig.KeyBytes)
			if token != strings.ToUpper(loginCmd.Token) {
				glog.Warningf("gateway::Server::OnReceivedCommand() token unmatch! loginCmd.Token: %s, token: %s\n",
					loginCmd.Token, token)
				return define.ErrNeedAuth
			}
		}
		glog.Infof("gateway::Server::OnReceivedCommand() login: %+v\n", loginCmd)
		resp, err = broker.Publish(srv.tag, command)
	} else {
		resp, err = broker.Publish(srv.tag, command)
	}

	if err != nil {
		glog.Warningf("gateway::Server::OnReceivedCommand() invoke (%s) error %s\n",
			command.Name, err)
		return define.ErrAuthFailed
	}
	if resp != nil && resp.Name == protocol.Close {
		glog.Warningln("gateway::Server::OnReceivedCommand() invoke response close")
		conn.Close(false)
		return define.ErrAuthFailed
	}

	if !conn.IsLogin() {
		conn.LoginSuccess(command.AppID, loginCmd.UserID, loginCmd.DeviceID, command.Version)
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

	glog.Infof("gateway::Server::OnReceivedCommand() invoke(%s) response %s",
		command.Name, resp)
	go srv.OnPushToUser(resp)
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

	// Todo: Find user by id and tag
	// Todo: Lockfree
	srv.Lock()
	if pushCmd.Tags == "*" {
		glog.Infof("Push message to all\n")
		// Push to all users
		for _, connections := range srv.connections {
			for _, conn := range connections {
				glog.Infof("Push to user %s%s", conn.ID(), touser)
				conn.Send(touser)
			}
		}
	} else {
		glog.Infof("Push message to %+v\n", pushCmd.UserIDList)
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
	}
	srv.Unlock()
}

// GetCheckSum Context接口
func (srv *Server) GetCheckSum(appid string) app.CheckSum {
	if appConfig, ok := srv.apps[appid]; ok {
		return appConfig
	}
	return nil
}
