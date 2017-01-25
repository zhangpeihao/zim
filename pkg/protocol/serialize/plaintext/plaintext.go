// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package plaintext 纯文本格式

用多行来分隔Command字段
第一行：信令版本（纯文本格式：第一个字符为：'t'，后面是协议版本号）
第二行：信令所属App ID
第三行：信令名
第四行：信令数据
第五行：信令负载
*/
package plaintext

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"strings"
)

const (
	// Version 版本
	Version = "t1"
	// ProbeByte 协议首字节
	ProbeByte byte = 't'
)

var (
	// CommandSep 命令分割字符
	CommandSep = []byte{'\n'}
)

const (
	// CommandVersionLine 信令版本索引
	CommandVersionLine = 0
	// CommandAppIDLine 信令所属App ID行索引
	CommandAppIDLine = 1
	// CommandNameLine 信令名行索引
	CommandNameLine = 2
	// CommandDataLine 信令数据行索引
	CommandDataLine = 3
	// CommandPayloadLine 信令负载行索引
	CommandPayloadLine = 4
	// CommandLines 信令行数
	CommandLines = 5
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
	if message == nil || len(message) == 0 || message[0] != 't' {
		glog.Warningln("protocol::serialize::plaintext::Parse() message unsupport\n")
		err = define.ErrUnsupportProtocol
		return
	}
	lines := bytes.SplitN(message, CommandSep, CommandLines)
	if len(lines) != CommandLines {
		glog.Warningln("protocol::serialize::plaintext::Parse() message has %s lines\n", len(lines))
		err = protocol.ErrParseFailed
		return
	}
	cmd = &protocol.Command{
		Version: strings.Trim(string(lines[CommandVersionLine]), "\r\t "),
		AppID:   strings.Trim(string(lines[CommandAppIDLine]), "\r\t "),
		Name:    strings.Trim(string(lines[CommandNameLine]), "\r\t "),
	}

	cmd.Payload = lines[CommandPayloadLine]
	if err = cmd.Parse(lines[CommandDataLine]); err != nil {
		return nil, err
	}
	return cmd, err
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	buf := bytes.NewBufferString(Version)
	buf.WriteByte('\n')
	buf.WriteString(cmd.AppID)
	buf.WriteByte('\n')
	buf.WriteString(cmd.Name)
	buf.WriteByte('\n')
	if cmd.Data != nil {
		enc := json.NewEncoder(buf)
		enc.Encode(cmd.Data)
	} else {
		buf.WriteByte('\n')
	}
	buf.Write(cmd.Payload)

	return buf.Bytes(), nil
}
