// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/broker/register"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

// BrokerImpl HTTP API实现的Broker
type BrokerImpl struct {
	sync.Mutex
	// RequestURL 请求地址
	RequestURL string
	// BindAddress 绑定地址
	BindAddress string
	// Debug 调试模式
	Debug bool
	// listener HTTP侦听对象
	listener net.Listener
	// httpServer HTTP服务
	httpServer *http.Server
	// queue 消息队列
	queues map[string]chan *protocol.Command
	// queusSize 消息队列长度
	queueSize int
	// ctx 上下文接口
	ctx context.Context
	// timeout 消息超时时间（单位：秒）
	timeout int
}

const (
	// ServerName 服务名
	ServerName = "broker-httpapi"
	// DefaultQueueSize 默认队列长度，如果设置的队列长度小于最小队列长度，则使用默认队列长度
	DefaultQueueSize = 1000
	// MinQueueSize 最小队列长度
	MinQueueSize = 64
	// DefaultTimeout 默认超时时间（5分钟）
	DefaultTimeout = 300
	// Name 调用类型的名称
	Name = "httpapi"
	// HeaderAgent HTTP User-Agent
	HeaderAgent = "zim"
	// HeaderAppID AppID
	HeaderAppID = "Zim-Appid"
	// HeaderName 信令名
	HeaderName = "Zim-Name"
	// HeaderData Data
	HeaderData = "Zim-Data"
	// HeaderPayloadMD5 Payload MD5值
	HeaderPayloadMD5 = "Zim-Payloadmd5"
	// HeaderNonce Nonce
	HeaderNonce = "Zim-Nonce"
	// HeaderTimestamp Timestamp
	HeaderTimestamp = "Zim-Timestamp"
	// HeaderCheckSum CheckSum
	HeaderCheckSum = "Zim-Checksum"
	// DefaultBindAddress 默认绑定地址
	DefaultBindAddress = ":8771"
	// DefaultRequestURL 默认请求地址
	DefaultRequestURL = "http://127.0.0.1:8880"
)

func init() {
	register.Register(Name, NewHTTPAPIBroker)
}

// NewHTTPAPIBroker 新建服务
func NewHTTPAPIBroker(viperPerfix string) (broker.Broker, error) {
	glog.Infoln("push::broker::NewHTTPAPIBroker")
	queueSize := viper.GetInt(viperPerfix + ".httpapi.queue-size")
	if queueSize < MinQueueSize {
		queueSize = DefaultQueueSize
	}
	timeout := viper.GetInt(viperPerfix + ".httpapi.timeout")
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	b := &BrokerImpl{
		RequestURL:  viper.GetString(viperPerfix + ".httpapi.request-url"),
		BindAddress: viper.GetString(viperPerfix + ".httpapi.subscribe-bind"),
		Debug:       viper.GetBool("debug"),
		queues:      make(map[string]chan *protocol.Command),
		queueSize:   queueSize,
		timeout:     timeout,
	}
	if len(b.BindAddress) == 0 {
		b.BindAddress = DefaultBindAddress
	}
	if len(b.RequestURL) == 0 {
		b.RequestURL = DefaultRequestURL
	}
	b.httpServer = &http.Server{Handler: b}
	if b.Debug {
		glog.Warningln("httpapi broker in debug mode!!!")
	}

	return b, nil
}

// Run 运行
func (b *BrokerImpl) Run(ctx context.Context) (err error) {
	glog.Infof("broker::httpapi::Run()\n")
	b.ctx = ctx
	b.listener, err = net.Listen("tcp4", b.BindAddress)
	if err != nil {
		glog.Errorf("broker::httpapi::Run() listen(%s) error: %s\n",
			b.BindAddress, err)
		return
	}
	var httpErr error
	go func() {
		httpErr = b.httpServer.Serve(b.listener)
	}()
	time.Sleep(time.Second)
	if httpErr != nil {
		glog.Errorf("broker::httpapi::Run() http.Server(%s) error: %s\n",
			b.BindAddress, err)
		return httpErr
	}

	return err
}

// Close 关闭
func (b *BrokerImpl) Close(timeout time.Duration) (err error) {
	glog.Warningln("broker::httpapi::Close()")
	defer glog.Warningln("broker::httpapi::Close() Done")

	// 关闭HTTP服务
	if b.listener != nil {
		err = b.listener.Close()
	}
	return err
}

// String 发布
func (b *BrokerImpl) String() string {
	return Name
}
