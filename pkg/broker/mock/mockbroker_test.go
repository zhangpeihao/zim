// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package mockbroker

import (
	"fmt"
	"testing"
	"time"

	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

var (
	cmdPublish, cmdSubscribe *protocol.Command
)

func TestMockBroker(t *testing.T) {
	var err error
	PublishMockHandler["t1"] = func(tag string, cmd *protocol.Command) (err error) {
		if tag != "t1" {
			t.Errorf("PublishMockHandler expect tag: %s, got: %s\n", "t1", tag)
			err = fmt.Errorf("publishMockHandler expect tag: %s, got: %s", "t1", tag)
			return
		}
		if !cmdPublish.Equal(cmd) {
			t.Errorf("PublishMockHandler cmdPublish: %s, got: %s\n", cmdPublish, cmd)
			err = fmt.Errorf("publishMockHandler cmdPublish: %s, got: %s", cmdPublish, cmd)
		}
		return err
	}

	// 初始化broker
	if err = broker.Init("test"); err != nil {
		t.Fatal("broker.Init() error:", err)
	}

	b := broker.Get("mock")
	if b == nil {
		t.Fatal(`broker.Get("mock") return nil`)
	}

	cmdPublish = &protocol.Command{
		Version: "t1",
		AppID:   "test",
		Name:    "msg/foo/bar",
		Data:    &protocol.GatewayMessageCommand{},
		Payload: []byte("foo bar"),
	}
	err = b.Publish("t1", cmdPublish)
	if err != nil {
		t.Error("b.Publish error:", err)
	}

	err = b.Subscribe("t2", func(tag string, cmd *protocol.Command) (err error) {
		return nil
	})
	if err == nil {
		t.Error("b.Subscribe should error if no handle set")
	}
	SubscribeMockHandler = func(tag string) (cmd *protocol.Command, err error) {
		time.Sleep(time.Second)
		cmdSubscribe = &protocol.Command{
			Version: "t1",
			AppID:   "test",
			Name:    "msg/foo/bar",
			Data:    &protocol.GatewayMessageCommand{},
			Payload: []byte("foo bar"),
		}
		return cmdSubscribe, nil
	}

	err = b.Subscribe("t2", func(tag string, cmd *protocol.Command) (err error) {
		if tag != "t2" {
			t.Errorf("Subscribe expect tag: %s, got: %s\n", "t1", tag)
			err = fmt.Errorf("subscribe expect tag: %s, got: %s", "t1", tag)
			return
		}
		if !cmdSubscribe.Equal(cmd) {
			t.Errorf("Subscribe cmdSubscribe: %s, got: %s\n", cmdSubscribe, cmd)
			err = fmt.Errorf("subscribe cmdSubscribe: %s, got: %s", cmdSubscribe, cmd)
		}

		return err
	})

}
