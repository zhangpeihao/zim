// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package alljson JSON格式

所有以'{'开头的数据，作为JSON格式处理
*/
package alljson

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"io"
	"io/ioutil"
)

const (
	// Version 版本
	Version = "j1"
	// ProbeByte 协议首字节
	ProbeByte byte = '{'
)

// ParseReader 使用Reader解析信令
func ParseReader(r io.Reader) (cmd *protocol.Command, err error) {
	if r == nil {
		return nil, define.ErrInvalidParameter
	}
	var buf []byte
	buf, err = ioutil.ReadAll(r)
	if err != nil {
		glog.Warningln("protocol::serialize::alljson::ParseReader() json.Unmarshal error:", err)
		return nil, err
	}
	return Parse(buf)
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
		switch cmd.FirstPartName() {
		case protocol.Login:
			var loginCmd protocol.GatewayLoginCommand
			err = Copy(cmd.Data, &loginCmd)
			if err != nil {
				glog.Warningln("protocol::serialize::alljson::Parse() copy loginCmd error:", err)
				return
			}
			cmd.Data = &loginCmd
		case protocol.Close:
			var closeCmd protocol.GatewayCloseCommand
			err = Copy(cmd.Data, &closeCmd)
			if err != nil {
				glog.Warningln("protocol::serialize::alljson::Parse() copy closeCmd error:", err)
				return
			}
			cmd.Data = &closeCmd
		case protocol.Message:
			var msgCmd protocol.GatewayMessageCommand
			err = Copy(cmd.Data, &msgCmd)
			if err != nil {
				glog.Warningln("protocol::serialize::alljson::Parse() copy msgCmd error:", err)
				return
			}
			cmd.Data = &msgCmd
		case protocol.Push2User:
			var pushCmd protocol.Push2UserCommand
			err = Copy(cmd.Data, &pushCmd)
			if err != nil {
				glog.Warningln("protocol::serialize::alljson::Parse() copy pushCmd error:", err)
				return
			}
			cmd.Data = &pushCmd
		}
	}
	return cmd, err
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	return json.Marshal(cmd)
}

// Copy 复制对象
func Copy(src, dest interface{}) (err error) {
	// Todo: 通过反射复制对象
	jsonData, err := json.Marshal(src)
	if err != nil {
		fmt.Println("1")
		return err
	}

	if err = json.Unmarshal(jsonData, dest); err != nil {
		fmt.Println("2")
	}
	return
}
