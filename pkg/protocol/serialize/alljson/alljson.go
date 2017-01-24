// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package alljson JSON格式

所有以'{'开头的数据，作为JSON格式处理
*/
package alljson

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
)

const (
	// Version 版本
	Version = "j1"
	// ProbeByte 协议首字节
	ProbeByte byte = '{'
)

var (
	serializer *serialize.Serializer = &serialize.Serializer{
		Version:   Version,
		ProbeByte: ProbeByte,
		Parse:     Parse,
		Compose:   Compose,
	}
)

func init() {
	serialize.Register(serializer)
}

// Parse 解析信令
func Parse(message []byte) (cmd *protocol.Command, err error) {
	if message == nil || len(message) == 0 || message[0] != '{' {
		glog.Warningln("protocol::serialize::alljson::Parse() message unsupport")
		err = define.ErrUnsupportProtocol
		return
	}
	var cmdData protocol.Command
	cmd = &cmdData
	if err = json.Unmarshal(message, cmd); err != nil {
		glog.Warningln("protocol::serialize::alljson::Parse() unmarshal error:", err)
		return
	}
	if cmd.Version != Version {
		glog.Warningf("protocol::serialize::alljson::Parse() unsupport version: %s\n", cmd.Version)
		err = define.ErrUnsupportProtocol
		return
	}

	if cmd.Data != nil {
		jsonData, err := json.Marshal(cmd.Data)
		if err != nil {
			return nil, err
		}

		if err = cmd.Parse(jsonData); err != nil {
			return nil, err
		}
	}
	return cmd, err
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	return json.Marshal(cmd)
}
