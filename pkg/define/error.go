// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

import "errors"

var (
	// ErrConnectionClosed 连接已关闭
	ErrConnectionClosed = errors.New("connection closed")
	// ErrUnsupportProtocol 协议不支持
	ErrUnsupportProtocol = errors.New("unsupport protocol")
	// ErrInvalidParameter 非法参数
	ErrInvalidParameter = errors.New("invalid parameter")
)
