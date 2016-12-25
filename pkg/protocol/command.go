// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"reflect"
	"strings"
)

const (
	// Login 登入
	Login = "login"
	// Close 关闭
	Close = "close"
	// Message 消息
	Message = "msg"
	// HeartBeat 心跳
	HeartBeat = "hb"
	// HeartBeatResponse 心跳响应
	HeartBeatResponse = "hbr"
)

// Command 信令
type Command struct {
	// Version 信令版本（第一个字母：'t'－文本协议，'p'－protobuf协议）
	Version string
	// Name 信令名，用'/'分隔多级信令（用于路由），例如：msg/foo/bar
	Name string
	// Data 网关信令数据
	Data interface{}
	// Payload 业务数据
	Payload []byte
}

// FirstPartName 第一段信令名
func (cmd *Command) FirstPartName() string {
	return strings.Split(cmd.Name, "/")[0]
}

// Equal 比较两个信令内容是否一样（用于测试）
func (cmd *Command) Equal(otherCmd *Command) bool {
	if otherCmd != nil &&
		strings.Compare(cmd.Version, otherCmd.Version) == 0 &&
		strings.Compare(cmd.Name, otherCmd.Name) == 0 &&
		bytes.Compare(cmd.Payload, otherCmd.Payload) == 0 &&
		reflect.TypeOf(cmd.Data) == reflect.TypeOf(otherCmd.Data) {
		// Serialize the data as JSON and compare
		data1, err := json.Marshal(cmd.Data)
		if err != nil {
			glog.Errorln("protocol::Command::Equal() json::Marshal(cmd.Data) error:", err)
			return false
		}
		data2, err := json.Marshal(otherCmd.Data)
		if err != nil {
			glog.Errorln("protocol::Command::Equal() json::Marshal(otherCmd.Data) error:", err)
			return false
		}
		return bytes.Compare(data1, data2) == 0
	}
	return false
}

// String 输出
func (cmd Command) String() string {
	var (
		data []byte
		err  error
	)
	if cmd.Data == nil {
		data = []byte("nil")
	} else {
		data, err = json.Marshal(cmd.Data)
		if err != nil {
			data = []byte("ERROR")
		}
	}
	return fmt.Sprintf("\n{\n  Version: %s\n  Name: %s\n  Data: %s\n  Payload: %+v\n}\n",
		cmd.Version, cmd.Name, string(data), cmd.Payload)
}