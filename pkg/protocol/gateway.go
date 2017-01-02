// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import (
	"crypto/md5"
	"fmt"
)

// GatewayLoginCommand 网关登入信令
type GatewayLoginCommand struct {
	// UserID 用户ID
	UserID string `json:"userid"`
	// DeviceID 设备ID
	DeviceID string `json:"deviceid"`
	// Timestamp Unix时间戳（单位秒）
	Timestamp int64 `json:"timestamp"`
	// Token 认证字
	Token string `json:"token"`
}

// GatewayCloseCommand 网关关闭信令
type GatewayCloseCommand struct {
	// UserID 用户ID
	UserID string `json:"userid"`
}

// GatewayMessageCommand 网关消息信令
type GatewayMessageCommand struct {
	// UserID 用户ID
	UserID string `json:"userid"`
}

// Key 密钥类型
type Key []byte

// Token 取得Token
func (key Key) Token(cmd *GatewayLoginCommand) string {
	h := md5.New()
	h.Write([]byte(cmd.UserID))
	h.Write([]byte(cmd.DeviceID))
	h.Write([]byte(fmt.Sprintf("%d", cmd.Timestamp)))
	h.Write(key)
	return fmt.Sprintf("%X", h.Sum(nil))
}
