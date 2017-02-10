// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

// Package context 上下文定义
package context

import (
	"github.com/zhangpeihao/zim/pkg/app"
)

// Context 环境接口
type Context interface {
	app.Service
}
