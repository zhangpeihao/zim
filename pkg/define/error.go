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
	// ErrKnownApp 不认识的App
	ErrKnownApp = errors.New("known application")
	// ErrAuthFailed 认证失败
	ErrAuthFailed = errors.New("auth failed")
	// ErrNeedAuth 需要认证
	ErrNeedAuth = errors.New("need auth")
	// ErrNoMoreMessage 没有消息
	ErrNoMoreMessage = errors.New("no more message")
)
