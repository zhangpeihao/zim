// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package websocket

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/shutdown"
	"github.com/zhangpeihao/zim/pkg/app"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	_ "github.com/zhangpeihao/zim/pkg/protocol/serialize/register"
	"github.com/zhangpeihao/zim/pkg/util/rand"
)

var (
	globalContext context.Context
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestHandler struct {
	sync.Mutex
	conn        define.Connection
	lastCommand *protocol.Command
}

// OnNewConnection 当有新连接建立
func (handler *TestHandler) OnNewConnection(conn define.Connection) {
	handler.Lock()
	defer handler.Unlock()
	handler.conn = conn
	fmt.Printf("connection(%s, %s, %s, %s)\n", conn.ID(), conn.UserID(), conn.AppID(), conn.DeviceID())
}

// OnCloseConnection 当有连接关闭
func (handler *TestHandler) OnCloseConnection(conn define.Connection) {
	handler.Lock()
	defer handler.Unlock()
	handler.conn = nil
}

// OnReceivedCommand 当收到命令
func (handler *TestHandler) OnReceivedCommand(conn define.Connection, command *protocol.Command) error {
	handler.Lock()
	defer handler.Unlock()
	handler.lastCommand = command
	conn.LoginSuccess(command.AppID, "123", "web", "t1")
	fmt.Printf("connection(%s, %s, %s, %s)\n", conn.ID(), conn.UserID(), conn.AppID(), conn.DeviceID())
	if !conn.IsLogin() {
		fmt.Println("Login state error")
	}
	conn.Send(command)
	return nil
}

func TestServer(t *testing.T) {
	handler := new(TestHandler)
	wsPort := rand.IntnRange(12300, 32300)
	wssPort := wsPort + 1
	viper.Set("gateway.ws-bind", fmt.Sprintf(":%d", wsPort))
	viper.Set("gateway.wss-bind", fmt.Sprintf(":%d", wssPort))
	viper.Set("gateway.wss-cert-file", "./httpcert/cert.pem")
	viper.Set("gateway.wss-key-file", "./httpcert/key.pem")
	s, err := NewServer(handler)
	if err != nil {
		t.Fatal("NewServer error:", err)
	}

	cfgKey := "123"
	appController, err := app.NewController(nil)
	if err != nil {
		log.Fatal("new app controller error:", err)
		return
	}
	appController.AddApp(&app.App{
		ID:       "test",
		Key:      cfgKey,
		KeyBytes: []byte(cfgKey),
	})
	globalContext = appController.SaveIntoContext(shutdown.NewContext())

	if err = s.Run(globalContext); err != nil {
		t.Fatal("Run error:", err)
	}

	// New websocket client
	client, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://localhost:%d/ws", wsPort), nil)
	if err != nil {
		t.Fatal("WebSocket Dial error:", err)
	}
	defer client.Close()

	buf := new(bytes.Buffer)
	buf.WriteString("t1\n")
	buf.WriteString("test\n")
	buf.WriteString(protocol.Login)
	buf.WriteByte('\n')
	login := protocol.GatewayLoginCommand{
		UserID:    "123",
		Timestamp: time.Now().Unix(),
	}
	login.Token = login.CalToken([]byte("1234567890"))
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
	handler.Lock()
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
			if obj.UserID != login.UserID {
				t.Errorf("Expect: %s\nGot: %s\n", login.UserID, obj.UserID)
			}
			if obj.Timestamp != login.Timestamp {
				t.Errorf("Expect: %d\nGot: %d\n", login.Timestamp, obj.Timestamp)
			}
			if obj.Token != login.Token {
				t.Errorf("Expect: %s\nGot: %s\n", login.Token, obj.Token)
			}
		}
	}
	handler.Unlock()

	client.WriteMessage(websocket.PingMessage, nil)
	client.WriteMessage(websocket.PongMessage, nil)
	client.WriteMessage(websocket.BinaryMessage, []byte("xxxx"))
	client.WriteMessage(websocket.TextMessage, nil)
	time.Sleep(time.Second)
	// Close client
	client.WriteMessage(websocket.CloseMessage, nil)
	client.Close()
	time.Sleep(time.Second)
	handler.Lock()
	if handler.conn != nil {
		t.Error("No OnCloseConnection callback")
	}
	handler.Unlock()

	err = shutdown.Shutdown(globalContext, time.Second, nil)
	if err != nil {
		t.Error("Close error:", err)
	}

}
