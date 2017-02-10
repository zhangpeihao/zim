// Copyright 2016-2017 Zhang Peihao <zhangpeihao@gmail.com>

// Package broker 异步消息接口
package broker

import (
	"time"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
)

// Broker 异步消息接口
type Broker interface {
	util.SafeCloseServer
	// Publish 发布消息到消息队列
	Publish(tag string, cmd *protocol.Command) (*protocol.Command, error)
	// Subscribe 从消息队列订阅消息
	Subscribe(tag string, handler SubscribeHandler) error
	// String 串行化输出
	String() string
}

// SubscribeHandler 订阅消息处理函数，返回error，消息将保留在队列中
type SubscribeHandler func(tag string, cmd *protocol.Command) error

var (
	brokers = make(map[string]Broker)
)

// Set 获取Broker
func Set(name string, broker Broker) {
	brokers[name] = broker
}

// Get 获取Broker
func Get(name string) Broker {
	if broker, found := brokers[name]; found {
		return broker
	}
	return nil
}

// Run 运行
func Run(closer *util.SafeCloser) (err error) {
	glog.Infoln("broker::define::Run()")
	defer glog.Infoln("broker::define::Run() done")
	for _, broker := range brokers {
		if err = broker.Run(closer); err != nil {
			glog.Errorln("broker::Run() error:", err)
			break
		}
	}
	return
}

// Close 关闭
func Close(timeout time.Duration) (err error) {
	glog.Warningln("broker::define::Close()")
	defer glog.Warningln("broker::define::Close() done")
	for _, broker := range brokers {
		go broker.Close(timeout)
	}
	return nil
}
