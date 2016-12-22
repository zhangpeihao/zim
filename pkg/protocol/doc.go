// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

// Package protocol 协议定义
// 传输层协议根据接入协议不同，分为TCP, UDP, WebSocket。
// TCP（开发中）
// UDP (开发中）
// WebSocket: 通过URL将IM请求路由到指定应用服务
// 应用层数据数据分为两段，第一行为第一段，后面内容为第二段
// 第一段为gateway协议内容
// 第二段为应用服务协议内容
package protocol
