// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import (
	"strconv"

	"github.com/zhangpeihao/zim/pkg/util"
)

// GatewayLoginCommand 网关登入信令
type GatewayLoginCommand struct {
	// UserID 用户ID
	UserID string `json:"userid"`
	// DeviceID 设备ID
	DeviceID string `json:"deviceid"`
	// Timestamp Unix时间戳（单位秒）
	Timestamp int64 `json:"timestamp"`
	// Token 认证字=md5(<app key>,UserID,DeviceID,Timestamp)
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

// CalToken 计算Token
func (cmd *GatewayLoginCommand) CalToken(key []byte) string {
	return util.CheckSumMD5(key, []byte(cmd.UserID),
		[]byte(cmd.DeviceID),
		[]byte(strconv.Itoa(int(cmd.Timestamp))))
}
