// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"testing"
	"time"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestHandler struct {
	conn        define.Connection
	lastCommand *protocol.Command
}

// OnNewConnection 当有新连接建立
func (handler *TestHandler) OnNewConnection(conn define.Connection) {
	handler.conn = conn
}

// OnCloseConnection 当有连接关闭
func (handler *TestHandler) OnCloseConnection(conn define.Connection) {
	handler.conn = nil
}

// OnReceivedCommand 当收到命令
func (handler *TestHandler) OnReceivedCommand(conn define.Connection, command *protocol.Command) {
	handler.lastCommand = command
}

func TestServer(t *testing.T) {
	t.Log("Start server")
	handler := new(TestHandler)
	s, err := NewServer(&ServerParameter{
		WebSocketBindAddress: ":12343",
	}, handler)
	if err != nil {
		t.Fatal("NewServer error:", err)
	}

	closer := util.NewSafeCloser()
	if err = s.Run(closer); err != nil {
		t.Fatal("Run error:", err)
	}

	// New websocket client
	client, _, err := websocket.DefaultDialer.Dial("ws://localhost:12343/ws", nil)
	if err != nil {
		t.Fatal("WebSocket Dial error:", err)
	}
	defer client.Close()

	buf := new(bytes.Buffer)
	buf.WriteString("t1\n")
	buf.WriteString(protocol.Login)
	buf.WriteByte('\n')
	login := protocol.GatewayLoginCommand{
		GatewayCommonCommand: protocol.GatewayCommonCommand{
			ID:        "123",
			Timestamp: time.Now().Unix(),
		},
	}
	login.Token = protocol.Key("1234567890").Token(&login.GatewayCommonCommand)
	enc := json.NewEncoder(buf)
	enc.Encode(&login)
	buf.WriteString("\n")
	payload := "12345678909878676542344"
	buf.WriteString(payload)

	err = client.WriteMessage(websocket.TextMessage, buf.Bytes())
	if err != nil {
		t.Fatal("WebSocket client write message error:", err)
	}

	time.Sleep(time.Second)
	if handler.conn == nil {
		t.Error("No OnNewConnection callback")
	}
	if handler.lastCommand == nil {
		t.Fatal("No OnReceivedCommand callback")
	}
	if handler.lastCommand.Name != protocol.Login {
		t.Errorf("Expect: %s\nGot: %s\n", protocol.Login, handler.lastCommand.Name)
	} else {
		if handler.lastCommand.Data == nil {
			t.Fatal("Data is nil")
		}
		if obj, ok := handler.lastCommand.Data.(*protocol.GatewayLoginCommand); !ok {
			t.Fatal("Data type error")
		} else {
			if obj.ID != login.ID {
				t.Errorf("Expect: %s\nGot: %s\n", login.ID, obj.ID)
			}
			if obj.Timestamp != login.Timestamp {
				t.Errorf("Expect: %d\nGot: %d\n", login.Timestamp, obj.Timestamp)
			}
			if obj.Token != login.Token {
				t.Errorf("Expect: %s\nGot: %s\n", login.Token, obj.Token)
			}
		}
	}

	// Close client
	client.Close()
	time.Sleep(time.Second)
	if handler.conn != nil {
		t.Error("No OnCloseConnection callback")
	}

	s.Close(time.Second)
	if err = closer.Close(time.Second); err != nil {
		t.Error("Close error:", err)
	}

}
