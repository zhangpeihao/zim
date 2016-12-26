// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

// Push2UserCommand 推送数据到用户
type Push2UserCommand struct {
	// AppID 应用ID
	AppID string `json:"appid"`
	// UserID 用户ID
	UserID string `json:"userid"`
	// Timestamp 时间戳
	Timestamp int64 `json:"timestamp"`
	// Token 校验Token
	Token string `json:"token"`
}
