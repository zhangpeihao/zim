// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package httpserver

import (
	"bytes"
	"flag"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/zim/pkg/protocol"
	_ "github.com/zhangpeihao/zim/pkg/protocol/serialize/register"
	"github.com/zhangpeihao/zim/pkg/util"
	"net/http"
	"testing"
	"time"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestHandler struct {
	data *protocol.Command
}

// OnPushToUser 推送消息给用户
func (handler *TestHandler) OnPushToUser(data *protocol.Command) {
	handler.data = data
}

func TestServer(t *testing.T) {
	handler := new(TestHandler)
	viper.Set("gateway.push-bind", ":12343")
	s, err := NewServer(handler)
	if err != nil {
		t.Fatal("NewServer error:", err)
	}

	closer := util.NewSafeCloser()
	if err = s.Run(closer); err != nil {
		t.Fatal("Run error:", err)
	}

	resp, err := http.Post("http://localhost:12343/p2u", "text/plain", bytes.NewBufferString(`t1
test
p2u
{"appid":"","userid":"","timestamp":0,"token":""}
foo bar`))
	if err != nil {
		t.Fatal("http.Post error: ", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("http.Post response status: %d\n", resp.StatusCode)
	}
	resp.Body.Close()

	s.Close(time.Second)
	if err = closer.Close(time.Second); err != nil {
		t.Error("Close error:", err)
	}

}
