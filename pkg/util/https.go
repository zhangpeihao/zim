// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"crypto/tls"
	"github.com/golang/glog"
	"net"
)

// NewHTTPSListener 新建HTTPS侦听
func NewHTTPSListener(certFile, keyFile, addr string) (listener net.Listener, err error) {
	glog.Infof("util::NewHTTPSListener(%s, %s, %s)\n", certFile, keyFile, addr)

	config := new(tls.Config)
	config.NextProtos = append(config.NextProtos, "http/1.1")
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	listener, err = tls.Listen("tcp", addr, config)
	if err != nil {
		glog.Errorf("util::NewHTTPSListener(%s, %s, %s) error: %s\n",
			certFile, keyFile, addr, err)
		return nil, err
	}

	return
}
