// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/driver"
	"github.com/zhangpeihao/zim/pkg/protocol/driver/plaintext"
)

var (
	// HeartBeatCommand 心跳命令
	HeartBeatCommand = &protocol.Command{
		Name: protocol.HeartBeat,
	}
	// HeartBeatResponseCommand 心跳响应命令
	HeartBeatResponseCommand = &protocol.Command{
		Name: protocol.HeartBeatResponse,
	}
)

// Connection 连接
type Connection struct {
	login    bool
	id       string
	userID   string
	appID    string
	deviceID string
	c        *websocket.Conn
}

// NewConnection 新建连接
func NewConnection(c *websocket.Conn) *Connection {
	return &Connection{
		c: c,
	}
}

// ID 连接ID
func (conn *Connection) ID() string {
	return conn.id
}

// AppID 应用ID
func (conn *Connection) AppID() string {
	return conn.appID
}

// UserID 用户ID
func (conn *Connection) UserID() string {
	return conn.userID
}

// DeviceID 设备ID
func (conn *Connection) DeviceID() string {
	return conn.deviceID
}

// ReadCommand 读取命令
func (conn *Connection) ReadCommand() (cmd *protocol.Command, err error) {
	mt, message, err := conn.c.ReadMessage()
	if err != nil {
		glog.Warningln("websocket::connection::Run() read:", err)
		return nil, err
	}
	switch mt {
	case websocket.CloseMessage:
		err = define.ErrConnectionClosed
	case websocket.BinaryMessage:
		err = define.ErrUnsupportProtocol
	case websocket.PingMessage:
		err = conn.c.WriteMessage(websocket.PongMessage, []byte{})
		cmd = HeartBeatCommand
	case websocket.PongMessage:
		cmd = HeartBeatResponseCommand
	case websocket.TextMessage:
		if message == nil || len(message) == 0 {
			glog.Warningln("websocket::connection::ReadCommand() message unsupport\n")
			err = define.ErrUnsupportProtocol
			return nil, err
		}
		switch message[0] {
		case 't':
		default:
			glog.Warningf("websocket::connection::ReadCommand() message type[%s] unsupport\n", string(message[0]))
			err = define.ErrUnsupportProtocol
			return nil, err
		}
		cmd, err = plaintext.Parse(message)
		if err != nil {
			glog.Warningf("websocket::connection::ReadCommand() plaintext.Parse error: %s\n", err)
			return nil, err
		}
	}
	return cmd, err
}

// Close 关闭链接（define::Connection接口函数）
func (conn *Connection) Close(force bool) error {
	return conn.c.Close()
}

// ToString 字符串输出（define::Connection接口函数）
func (conn *Connection) String() string {
	return "webdocket[" + conn.c.RemoteAddr().String() + "]"
}

// LoginSuccess 登入成功
func (conn *Connection) LoginSuccess(appID, userID, deviceID string) {
	conn.appID = appID
	conn.userID = userID
	conn.id = define.ConnectionID(appID, userID)
	conn.login = true
}

// IsLogin 登入状态
func (conn *Connection) IsLogin() bool {
	return conn.login
}

// Send 发送命令
func (conn *Connection) Send(cmd *protocol.Command) error {
	message, err := driver.Compose(cmd)
	if err != nil {
		return err
	}
	return conn.c.WriteMessage(websocket.TextMessage, message)
}
