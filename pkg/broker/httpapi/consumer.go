// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/zhangpeihao/shutdown"
	"github.com/zhangpeihao/zim/pkg/broker"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
)

// Subscribe 订阅
func (b *BrokerImpl) Subscribe(tag string, handler broker.SubscribeHandler) error {
	glog.Infof("broker::httpapi::Subscribe(%s)\n", tag)
	defer glog.Infof("broker::httpapi::Subscribe(%s) done\n", tag)
	if err := shutdown.ExitWaitGroupAdd(b.ctx, 1); err != nil {
		glog.Errorf("broker::httpapi::Subscribe(%s) ExitWaitGroupAdd error: %s", tag, err)
		return err
	}
	defer shutdown.ExitWaitGroupDone(b.ctx)
	queue := make(chan *protocol.Command, b.queueSize)
	b.Lock()
	b.queues[tag] = queue
	b.Unlock()
FOR_LOOP:
	for {
		select {
		case cmd := <-queue:
			func() {
				util.RecoverFromPanic()
				handler(tag, cmd)
			}()
		case <-b.ctx.Done():
			glog.Infof("broker::httpapi::Subscribe(%s) break by context", tag)
			break FOR_LOOP
		}
	}
	return nil
}

// ServeHTTP 处理HTTP链接
func (b *BrokerImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	glog.Infoln("broker::httpapi::ServeHTTP()")
	defer r.Body.Close()

	tag := strings.ToLower(strings.Trim(r.URL.Path, "/"))
	glog.Infoln("broker::httpapi::ServeHTTP() tag: ", tag)
	if tag == "debug.html" {
		if b.Debug {
			//			b.HandleDebug(w, r)
		} else {
			glog.Warningln("broker::httpapi::ServeHTTP() not in debug mode")
			w.WriteHeader(404)
			return
		}
	} else {
		b.Lock()
		queue, ok := b.queues[tag]
		b.Unlock()
		if !ok {
			glog.Warningln("broker::httpapi::ServeHTTP() no tag(", tag, ")")
			w.WriteHeader(404)
			return
		}

		if payload, err := ioutil.ReadAll(r.Body); err != nil {
			glog.Warningf("broker::httpapi::ServeHTTP() Read payload error: %s\n",
				err)
			w.WriteHeader(400)
		} else {
			if cmd, err := ParseCommand(b.ctx, tag, r.Header, payload, b.timeout); err != nil {
				glog.Warningf("broker::httpapi::ServeHTTP() ParseCommand error: %s\n",
					err)
				w.WriteHeader(400)
			} else {
				queue <- cmd
				w.WriteHeader(200)
			}
		}
	}
}
