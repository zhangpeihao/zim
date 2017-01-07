// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package router

import (
	"bytes"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/invoker"
	"github.com/zhangpeihao/zim/pkg/invoker/driver/httpapi"
)

// Info 最简单的路由信息
type Info struct {
	Protocol  string `json:"protocol"`
	Parameter string `json:"parameter"`
}

// InfoMap 最简单的路由信息Map
type InfoMap map[string]Info

// Router 路由
type Router struct {
	defaultInvoker invoker.Invoker
	invokers       map[string]invoker.Invoker
}

// NewRouter 新建Router
func NewRouter(routerMap InfoMap) (r *Router, err error) {
	r = &Router{
		invokers: make(map[string]invoker.Invoker),
	}

	for key, routeInfo := range routerMap {
		switch routeInfo.Protocol {
		case httpapi.Name:
			if key == "*" {
				r.defaultInvoker = httpapi.NewInvoker(routeInfo.Parameter)
			} else {
				r.invokers[key] = httpapi.NewInvoker(routeInfo.Parameter)
			}
		default:
			glog.Errorf("router::driver::jsonfile::NewRouter(%s) unsupport protocol %s\n",
				routeInfo.Protocol, routeInfo.Protocol)
			return nil, define.ErrUnsupportProtocol
		}
	}
	return r, nil
}

// Find 查询路由
func (r *Router) Find(name string) invoker.Invoker {
	glog.Infof("Router::Find(%s)\n", name)
	inv, found := r.invokers[name]
	if !found {
		return r.defaultInvoker
	}
	return inv
}

// String 输出
func (r *Router) String() string {
	buf := new(bytes.Buffer)
	if r.defaultInvoker != nil {
		buf.WriteString("*: ")
		buf.WriteString(r.defaultInvoker.String())
		buf.WriteString("\n")
	}
	for key, inv := range r.invokers {
		buf.WriteString(key)
		buf.WriteString(": ")
		buf.WriteString(inv.String())
		buf.WriteString("\n")
	}
	return string(buf.Bytes())
}
