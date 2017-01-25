// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package httpapi 提供基于http的调用

将Command发送的HTTP服务器

Command.Name: 作为请求URL的path最后一段，例如：HTTP服务的请求地址是"http://localhost/test"，那么，Command.Name为"msg/foo/bar"将被路由到"http://localhost/test/msg/foo/bar"

Command.Payload: 作为POST内容发送
*/
package httpapi

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	// Name 调用类型的名称
	Name = "httpapi"
	// HeaderAgent HTTP User-Agent
	HeaderAgent = "zim"
	// HeaderUserID 用户ID名
	HeaderUserID = "zim-UserID"
	// HeaderAppID AppID名
	HeaderAppID = "zim-AppID"
	// Timeout 请求超时
	Timeout = time.Duration(30 * time.Second)
)

var (
	// Transport 默认HTTP传输对象
	Transport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		Dial: dialTimeout,
	}
)

// diaoTimeout 指定超时时间的连接器
func dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, Timeout)
}

// Invoker 调用器
type Invoker struct {
	// RequestURL 请求地址
	RequestURL string
}

// NewInvoker 新建调用器
func NewInvoker(requestURL string) *Invoker {
	glog.Infof("invoker::driver::httpapi::NewInvoker(%s)\n", requestURL)
	return &Invoker{
		RequestURL: requestURL,
	}
}

// Invoke 通过HTTP API调用，发送信令
func (invoker *Invoker) Invoke(userID string, reqCmd *protocol.Command) (respCmd *protocol.Command, err error) {
	glog.Infof("invoker::driver::httpapi::Invoker::Invoke\n")
	var req *http.Request

	req, err = http.NewRequest("POST", invoker.RequestURL+"/"+reqCmd.Name,
		bytes.NewBuffer(reqCmd.Payload))
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Invoke() error: %s\n", err)
		return
	}
	req.Header.Set("User-Agent", HeaderAgent)
	if len(userID) > 0 {
		req.Header.Set(HeaderUserID, userID)
	}
	req.Header.Set(HeaderAppID, reqCmd.AppID)
	req.Close = true

	client := &http.Client{Transport: Transport}
	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Invoke() http error: %s\n", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		glog.Errorf("invoker::driver::httpapi::Invoke() http response status %d\n", res.StatusCode)
		err = fmt.Errorf("http response status %d", res.StatusCode)
		return
	}
	var rawbody []byte
	rawbody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Invoke() http response error: %s\n", err)
		return
	}
	if len(rawbody) == 0 {
		glog.Infoln("invoker::driver::httpapi::Invoke() http is empty")
		return
	}

	respCmd, err = serialize.Parse(rawbody)
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Invoke() parse http response error: %s\n", err)
		return
	}

	return
}

// String 输出
func (invoker *Invoker) String() string {
	return "invoke http API (" + invoker.RequestURL + ")"
}
