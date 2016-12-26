// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package router

import "github.com/zhangpeihao/zim/pkg/invoker"

// SampleInfo 最简单的路由信息
type SampleInfo struct {
	Protocol  string `json:"protocol"`
	Parameter string `json:"parameter"`
}

// SampleMap 最简单的路由信息Map
type SampleMap map[string]*SampleInfo

// Router 路由接口
type Router interface {
	Find(app, name string) invoker.Invoker
}
