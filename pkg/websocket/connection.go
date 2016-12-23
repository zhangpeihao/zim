// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
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
	c *websocket.Conn
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
		lines := bytes.SplitN(message, protocol.CommandSep, protocol.CommandLines)
		if len(lines) != protocol.CommandLines {
			glog.Warningln("websocket::connection::Run() message has %s lines\n", len(lines))
			err = protocol.ErrParseFailed
			break
		}
		cmd = &protocol.Command{
			Name: string(lines[protocol.CommandNameLine]),
		}
		data := lines[protocol.CommandDataLine]
		cmd.Payload = string(lines[protocol.CommandPayloadLine])
		switch cmd.Name {
		case protocol.Login:
			var loginCmd protocol.GatewayLoginCommand
			if err = json.Unmarshal(data, &loginCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				break
			}
			cmd.Data = &loginCmd
		case protocol.Close:
			var closeCmd protocol.GatewayCloseCommand
			if err = json.Unmarshal(data, &closeCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				break
			}
			cmd.Data = &closeCmd
		case protocol.Message:
			var msgCmd protocol.GatewayMessageCommand
			if err = json.Unmarshal(data, &msgCmd); err != nil {
				glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
				break
			}
			cmd.Data = &msgCmd
		}
	}
	return cmd, nil
}

// Close 关闭链接（define::Connection接口函数）
func (conn *Connection) Close(force bool) error {
	return conn.c.Close()
}

// ToString 字符串输出（define::Connection接口函数）
func (conn *Connection) String() string {
	return "webdocket[" + conn.c.RemoteAddr().String() + "]"
}
