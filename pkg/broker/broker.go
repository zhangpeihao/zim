// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

// Package broker 异步消息接口
package broker

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

// Broker 异步消息接口
type Broker interface {
	// Publish 发布消息到消息队列
	Publish(tag string, cmd *protocol.Command) error
	// Subscribe 从消息队列订阅消息
	Subscribe(tag string, handler SubscribeHandler) error
}

// SubscribeHandler 订阅消息处理函数，返回error，消息将保留在队列中
type SubscribeHandler func(tag string, cmd *protocol.Command) error

// NewBrokerHandler 新建Broker函数，参数：viper参数perfix
type NewBrokerHandler func(string) (Broker, error)

var (
	brokerHandlers = make(map[string]NewBrokerHandler)
	brokers        = make(map[string]Broker)
)

// Register 注册Broker
func Register(name string, handler NewBrokerHandler) {
	if _, found := brokerHandlers[name]; found {
		glog.Warningf("broker::Register() Broker[%s] existed\n")
	}
	brokerHandlers[name] = handler
}

// Init 初始化
func Init(viperPerfix string) error {
	for name, brokerHandler := range brokerHandlers {
		glog.Infof("broker::Init() Init broker[%s]\n", name)
		broker, err := brokerHandler(viperPerfix)
		if err != nil {
			glog.Errorf("broker::Init() init broker[%s] init error: %s", name, err)
			return err
		}
		brokers[name] = broker
	}
	return nil
}

// Get 获取Brokern
func Get(name string) Broker {
	if broker, found := brokers[name]; found {
		return broker
	}
	return nil
}
