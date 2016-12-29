// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
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
	login bool
	c     *websocket.Conn
}

// NewConnection 新建连接
func NewConnection(c *websocket.Conn) *Connection {
	return &Connection{
		c: c,
	}
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
			glog.Warningln("websocket::connection::Run() message unsupport\n")
			err = define.ErrUnsupportProtocol
			return nil, err
		}
		switch message[0] {
		case 't':
		default:
			glog.Warningf("websocket::connection::Run() message type[%s] unsupport\n", string(message[0]))
			err = define.ErrUnsupportProtocol
			return nil, err
		}
		lines := bytes.SplitN(message, plaintext.CommandSep, plaintext.CommandLines)
		if len(lines) != plaintext.CommandLines {
			glog.Warningf("websocket::connection::Run() message has %s lines\n", len(lines))
			err = protocol.ErrParseFailed
			return nil, err
		}
		cmd = &protocol.Command{
			Version: string(lines[plaintext.CommandVersionLine]),
			Name:    string(lines[plaintext.CommandNameLine]),
		}
		data := lines[plaintext.CommandDataLine]
		cmd.Payload = lines[plaintext.CommandPayloadLine]
		switch cmd.FirstPartName() {
		case protocol.Login:
			var loginCmd protocol.GatewayLoginCommand
			if err = json.Unmarshal(data, &loginCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				return nil, err
			}
			cmd.Data = &loginCmd
		case protocol.Close:
			var closeCmd protocol.GatewayCloseCommand
			if err = json.Unmarshal(data, &closeCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				return nil, err
			}
			cmd.Data = &closeCmd
		case protocol.Message:
			var msgCmd protocol.GatewayMessageCommand
			if err = json.Unmarshal(data, &msgCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				return nil, err
			}
			cmd.Data = &msgCmd
		case protocol.HeartBeat:
			conn.Send(HeartBeatResponseCommand)
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
func (conn *Connection) LoginSuccess() {
	conn.login = true
}

// IsLogin 登入状态
func (conn *Connection) IsLogin() bool {
	return conn.login
}

// Send 发送命令
func (conn *Connection) Send(cmd *protocol.Command) error {
	return conn.c.WriteMessage(websocket.TextMessage, plaintext.Compose(cmd))
}
