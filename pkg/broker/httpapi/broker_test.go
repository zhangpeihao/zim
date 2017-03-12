// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/zhangpeihao/shutdown"
	"github.com/zhangpeihao/zim/pkg/app"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/broker/register"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

const (
	httpport    = 8656
	subport     = 8765
	testtag     = "testtag"
	viperPerfix = "test"
)

var (
	globalContext            context.Context
	cmdPublish, cmdSubscribe *protocol.Command
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")

	// 初始化broker
	viper.Set(viperPerfix+".httpapi.request-url", fmt.Sprintf("http://localhost:%d/test", httpport))
	viper.Set(viperPerfix+".httpapi.subscribe-bind", fmt.Sprintf("127.0.0.1:%d", subport))
}

func TestSuite(t *testing.T) {
	var err error
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

	if err = register.Init(viperPerfix); err != nil {
		fmt.Println("broker.Init() error:", err)
	}
	broker.Run(globalContext)
	t.Run("Producer", testProducer)
	t.Run("Consumer", testConsumer)

	err = shutdown.Shutdown(globalContext, time.Second*2, nil)
	if err != nil {
		t.Error("close timeout")
	}
}

func testProducer(t *testing.T) {
	var err error

	b := broker.Get(Name)
	if b == nil {
		t.Fatal(`broker.Get("httpapi") return nil`)
	}

	signal := make(chan *protocol.Command)
	http.HandleFunc("/test/"+testtag, func(w http.ResponseWriter, r *http.Request) {
		if payload, err := ioutil.ReadAll(r.Body); err != nil {
			glog.Warningf("broker::httpapi::ServeHTTP() Read payload error: %s\n",
				err)
			w.WriteHeader(400)
		} else {
			if cmd, err := ParseCommand(globalContext, testtag, r.Header, payload, 10); err != nil {
				glog.Warningf("broker::httpapi::ServeHTTP() ParseCommand error: %s\n",
					err)
				w.WriteHeader(400)
			} else {
				signal <- cmd
			}
		}
	})
	go http.ListenAndServe(fmt.Sprintf(":%d", httpport), nil)

	cmdPublish = &protocol.Command{
		Version: "",
		AppID:   "test",
		Name:    "msg/foo/bar",
		Data:    &protocol.GatewayMessageCommand{},
		Payload: []byte("foo bar"),
	}

	go func() {
		var resp *protocol.Command
		resp, err = b.Publish(testtag, cmdPublish)
		assert.NoError(t, err)
		assert.Nil(t, resp)
	}()
	time.Sleep(time.Second)
	select {
	case signalCmd := <-signal:
		if !cmdPublish.Equal(signalCmd) {
			t.Errorf("signal command not equal publish command\nsignalCmd: %s\n cmdPublish: %s\n",
				signalCmd, cmdPublish)
		}
	case <-time.After(time.Second):
		t.Error("wait signal command timeout")
	}
}

func testConsumer(t *testing.T) {
	var err error

	b := broker.Get(Name)
	if b == nil {
		t.Fatal(`broker.Get("httpapi") return nil`)
	}

	signal := make(chan *protocol.Command)
	go b.Subscribe(testtag, func(tag string, cmd *protocol.Command) error {
		signal <- cmd
		return nil
	})

	// Send test command
	cmdSubscribe = &protocol.Command{
		Version: "",
		AppID:   "test",
		Name:    "msg/foo/bar",
		Data:    &protocol.GatewayMessageCommand{},
		Payload: []byte("foo bar"),
	}
	if err = SendCommand(cmdSubscribe); err != nil {
		t.Error("SendCommand error:", err)
	}
	time.Sleep(time.Second)
	select {
	case signalCmd := <-signal:
		if !cmdSubscribe.Equal(signalCmd) {
			t.Errorf("signal command not equal publish command\nsignalCmd: %s\n cmdSubscribe: %s\n",
				signalCmd, cmdSubscribe)
		}
	case <-time.After(time.Second):
		t.Error("wait signal command timeout")
	}
}

func SendCommand(cmd *protocol.Command) (err error) {
	var req *http.Request
	req, err = http.NewRequest("POST", fmt.Sprintf("http://127.0.0.1:%d/%s", subport, testtag),
		bytes.NewBuffer(cmd.Payload))
	if err != nil {
		glog.Errorf("SendCommand() error: %s\n", err)
		return
	}

	err = ComposeCommand(globalContext, testtag, req.Header, cmd)
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Publish() ComposeCommand error: %s\n", err)
		return
	}

	req.Close = true

	client := &http.Client{Transport: Transport}
	var httpResp *http.Response
	httpResp, err = client.Do(req)
	if err != nil {
		glog.Errorf("broker::httpapi::Publish() http error: %s\n", err)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("response status code %d", httpResp.StatusCode)
	}
	return
}
