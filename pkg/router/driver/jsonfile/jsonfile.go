// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package jsonfile 通过JSON文件配置的路由

路由配置格式：
	{
	  "<信令名前缀匹配串>":{"protocol":"httpapi","parameter":"http://xxx/api"}
	}
*/
package jsonfile

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/invoker"
	"github.com/zhangpeihao/zim/pkg/invoker/driver/httpapi"
	"github.com/zhangpeihao/zim/pkg/router"
	"os"
)

// Router 通过JSON文件路由
type Router struct {
	invokers map[string]invoker.Invoker
}

// NewRouter 新建路由
func NewRouter(file string) (r *Router, err error) {
	glog.Infof("router::driver::jsonfile::NewRouter(%s)\n", file)
	var f *os.File
	if f, err = os.Open(file); err != nil {
		glog.Errorf("router::driver::jsonfile::NewRouter(%s) error: %s\n",
			file, err)
		return
	}

	// 加载查询Map
	appSearchMap := make(map[string]router.SampleMap)
	dec := json.NewDecoder(f)
	err = dec.Decode(&appSearchMap)
	if err != nil {
		glog.Errorf("router::driver::jsonfile::NewRouter(%s) json.Decode error: %s\n",
			file, err)
		return
	}

	r = &Router{
		invokers: make(map[string]invoker.Invoker),
	}

	for app, searchMap := range appSearchMap {
		for key, routeInfo := range searchMap {
			if routeInfo.Protocol != httpapi.Name {
				glog.Errorf("router::driver::jsonfile::NewRouter(%s) unsupport protocol %s\n",
					routeInfo.Protocol)
				return nil, define.ErrUnsupportProtocol
			}
			r.invokers[app+"#"+key] = httpapi.NewInvoker(routeInfo.Parameter)
		}
	}

	return
}

// Find 查询路由
func (r *Router) Find(app, name string) invoker.Invoker {
	inv, found := r.invokers[app+"#"+name]
	if !found {
		return nil
	}
	return inv
}
