// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"bytes"
	"crypto/tls"
	//	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/protocol"
	//	"github.com/zhangpeihao/zim/pkg/util"
)

const (
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

// Publish 发布
func (b *BrokerImpl) Publish(tag string, cmd *protocol.Command) (resp *protocol.Command, err error) {
	glog.Infof("broker::httpapi::Publish(%s)%s\n", tag, cmd)
	defer glog.Infof("broker::httpapi::Publish() done\n")

	var req *http.Request
	req, err = http.NewRequest("POST", b.RequestURL+"/"+tag,
		bytes.NewBuffer(cmd.Payload))
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Publish() error: %s\n", err)
		return
	}

	err = ComposeCommand(b.ctx, tag, req.Header, cmd)
	if err != nil {
		glog.Errorf("invoker::driver::httpapi::Publish() ComposeCommand error: %s\n", err)
		return
	}
	glog.Infof("req.Header: %+v\n", req.Header)
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
		glog.Errorf("broker::httpapi::Publish() http response status %d\n", httpResp.StatusCode)
		return nil, fmt.Errorf("http response status %d", httpResp.StatusCode)
	}

	// Check response
	var respPayload []byte
	if respPayload, err = ioutil.ReadAll(httpResp.Body); err != nil {
		glog.Warningf("broker::httpapi::Publish() Read response payload error: %s\n",
			err)
		return nil, nil
	}
	glog.Infof("resp.Header: %+v\n", httpResp.Header)
	resp, _ = ParseCommand(b.ctx, tag, httpResp.Header, respPayload, b.timeout)

	return
}
