// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package plaintext 纯文本格式

用多行来分隔Command字段
第一行：信令版本（纯文本格式：第一个字符为：'t'，后面是协议版本号）
第二行：信令名
第三行：信令数据
第四行：信令负载
*/
package plaintext

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"strings"
)

var (
	// CommandSep 命令分割字符
	CommandSep = []byte{'\n'}
)

const (
	// CommandVersionLine 信令版本索引
	CommandVersionLine = 0
	// CommandNameLine 信令名行索引
	CommandNameLine = 1
	// CommandDataLine 信令数据行索引
	CommandDataLine = 2
	// CommandPayloadLine 信令负载行索引
	CommandPayloadLine = 3
	// CommandLines 信令行数
	CommandLines = 4
)

// Parse 解析信令
func Parse(message []byte) (cmd *protocol.Command, err error) {
	if message == nil || len(message) == 0 || message[0] != 't' {
		glog.Warningln("websocket::connection::Run() message unsupport\n")
		err = define.ErrUnsupportProtocol
		return
	}
	lines := bytes.SplitN(message, CommandSep, CommandLines)
	if len(lines) != CommandLines {
		glog.Warningln("websocket::connection::Run() message has %s lines\n", len(lines))
		err = protocol.ErrParseFailed
		return
	}
	cmd = &protocol.Command{
		Version: strings.Trim(string(lines[CommandVersionLine]), "\r\t "),
		Name:    strings.Trim(string(lines[CommandNameLine]), "\r\t "),
	}

	data := lines[CommandDataLine]
	cmd.Payload = lines[CommandPayloadLine]
	switch cmd.FirstPartName() {
	case protocol.Login:
		var loginCmd protocol.GatewayLoginCommand
		if err = json.Unmarshal(data, &loginCmd); err != nil {
			glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
			break
		}
		cmd.Data = &loginCmd
	case protocol.Close:
		var closeCmd protocol.GatewayCloseCommand
		if err = json.Unmarshal(data, &closeCmd); err != nil {
			glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
			break
		}
		cmd.Data = &closeCmd
	case protocol.Message:
		var msgCmd protocol.GatewayMessageCommand
		if err = json.Unmarshal(data, &msgCmd); err != nil {
			glog.Warningln("websocket::connection::Run() json.Unmarshal error:", err)
			break
		}
		cmd.Data = &msgCmd
	}
	return cmd, err
}
