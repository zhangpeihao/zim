// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import (
	"crypto/md5"
	"fmt"
)

// GatewayCommonCommand 网关通用信令
type GatewayCommonCommand struct {
	// ID 用户ID
	ID string `json:"id"`
	// Timestamp Unix时间戳（单位秒）
	Timestamp int64 `json:"timestamp"`
	// Token 认证字
	Token string `json:"token"`
}

// GatewayLoginCommand 网关登入信令
type GatewayLoginCommand struct {
	// GatewayCommonCommand 公用命令
	GatewayCommonCommand
}

// GatewayCloseCommand 网关关闭信令
type GatewayCloseCommand struct {
	// GatewayCommonCommand 公用命令
	GatewayCommonCommand
}

// GatewayMessageCommand 网关消息信令
type GatewayMessageCommand struct {
	// GatewayCommonCommand 公用命令
	GatewayCommonCommand
}

// Key 密钥类型
type Key []byte

// Token 取得Token
func (key Key) Token(cmd *GatewayCommonCommand) string {
	h := md5.New()
	h.Write([]byte(cmd.ID))
	h.Write([]byte(fmt.Sprintf("%d", cmd.Timestamp)))
	h.Write(key)
	return fmt.Sprintf("%X", h.Sum(nil))
}
