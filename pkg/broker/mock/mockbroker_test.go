// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package mock

import (
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/broker/register"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

var (
	cmdPublish, cmdSubscribe *protocol.Command
	locker                   sync.Mutex
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestMockBroker(t *testing.T) {
	var err error
	var testTag = "mocktag"
	PublishMockHandler[testTag] = func(tag string, cmd *protocol.Command) (resp *protocol.Command, err error) {
		glog.Infof("PublishMockHandler(%s)%s\n", tag, cmd)
		glog.Infof("PublishMockHandler() done\n")
		if tag != testTag {
			t.Errorf("PublishMockHandler expect tag: %s, got: %s\n", testTag, tag)
			err = fmt.Errorf("publishMockHandler expect tag: %s, got: %s", testTag, tag)
			return
		}
		if !cmdPublish.Equal(cmd) {
			t.Errorf("PublishMockHandler cmdPublish: %s, got: %s\n", cmdPublish, cmd)
			err = fmt.Errorf("publishMockHandler cmdPublish: %s, got: %s", cmdPublish, cmd)
		}
		locker.Lock()
		cmdSubscribe = cmd.Copy()
		locker.Unlock()
		return nil, nil
	}

	// 初始化broker
	if err = register.Init("test"); err != nil {
		t.Fatal("register.Init() error:", err)
	}

	b := broker.Get("mock")
	if b == nil {
		t.Fatal(`broker.Get("mock") return nil`)
	}

	b.Run(nil)

	if b.String() != "mock" {
		t.Errorf("b.String :%s\n", b.String())
	}

	SubscribeMockHandler = func(tag string) (cmd *protocol.Command, err error) {
		glog.Infof("SubscribeMockHandler(%s)%s\n", tag)
		for {
			locker.Lock()
			if cmdSubscribe == nil {
				locker.Unlock()
				time.Sleep(time.Second)
			} else {
				cmd = cmdSubscribe.Copy()
				locker.Unlock()
				glog.Infof("SubscribeMockHandler got(%s)%s\n", tag, cmd)
				return
			}
		}
	}
	cmdPublish = &protocol.Command{
		Version: "t1",
		AppID:   "test",
		Name:    "msg/foo/bar",
		Data:    &protocol.GatewayMessageCommand{},
		Payload: []byte("foo bar"),
	}

	finishSignal := make(chan struct{})
	go func() {
		err = b.Subscribe(testTag, func(tag string, cmd *protocol.Command) (err error) {
			glog.Infof("Subscribe got(%s)%s\n", tag, cmd)
			defer func() {
				finishSignal <- struct{}{}
			}()
			if tag != testTag {
				t.Errorf("Subscribe expect tag: %s, got: %s\n", testTag, tag)
				err = fmt.Errorf("publishMockHandler expect tag: %s, got: %s", testTag, tag)
				return
			}
			if !cmdPublish.Equal(cmd) {
				t.Errorf("Subscribe cmdPublish: %s, got: %s\n", cmdPublish, cmd)
				err = fmt.Errorf("publishMockHandler cmdPublish: %s, got: %s", cmdPublish, cmd)
			}
			return err
		})
		if err == nil {
			t.Error("b.Subscribe should error if no handle set")
		}
	}()

	_, err = b.Publish(testTag, cmdPublish)
	if err != nil {
		t.Error("b.Publish error:", err)
	}

	select {
	case <-finishSignal:
	case <-time.After(time.Second * 4):
		t.Error("Test timeout")
	}
	b.Close(time.Second)
}
