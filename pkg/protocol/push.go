// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

// Push2UserCommand 推送数据
type Push2UserCommand struct {
	// UserIDList 目标用户ID，逗号分隔
	UserIDList string `json:"useridlist"`
}
