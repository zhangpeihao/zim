// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

// Package mockbroker 模拟Broker，用于测试
package mockbroker

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

type brokerImpl struct {
	sync.Mutex
	viperPerfix string
}

var (
	// PublishMockHandler Publish模拟回调函数
	PublishMockHandler = make(map[string]func(string, *protocol.Command) error)
	// SubscribeMockHandler Subscribe模拟回调函数，在函数中适当Sleep
	SubscribeMockHandler func(string) (*protocol.Command, error)
)

func init() {
	broker.Register("mock", NewMockBroker)
}

// NewMockBroker 新建Mock broker
func NewMockBroker(viperPerfix string) (broker.Broker, error) {
	return &brokerImpl{
		viperPerfix: viperPerfix,
	}, nil
}

// Publish 发布
func (b *brokerImpl) Publish(tag string, cmd *protocol.Command) error {
	b.Lock()
	handler, found := PublishMockHandler[tag]
	if !found {
		glog.Warningf("broker::mock::Publish(%s) no handler!\n", tag)
		b.Unlock()
		return fmt.Errorf("not find publish handler for tag %s", tag)
	}
	b.Unlock()

	return handler(tag, cmd)
}

// Subscribe 订阅
func (b *brokerImpl) Subscribe(tag string, handler broker.SubscribeHandler) error {
	if SubscribeMockHandler == nil {
		return fmt.Errorf("SubscribeMockHandler not set")
	}
	cmd, err := SubscribeMockHandler(tag)
	if err != nil {
		glog.Warningf("broker::mock::Subscribe(%s) SubscribeMockHandler return %s!\n", err)
		return err
	}
	return handler(tag, cmd)
}
