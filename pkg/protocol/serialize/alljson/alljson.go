// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package alljson JSON格式

所有以'{'开头的数据，作为JSON格式处理
*/
package alljson

import (
	"bufio"
	"bytes"
	"encoding/json"

	"github.com/golang/glog"
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
	serializer = &serialize.Serializer{
		Version:        Version,
		ProbeByte:      ProbeByte,
		NewParseEngine: NewParseEngine,
		Compose:        Compose,
	}
)

// Command 命令
type Command struct {
	// Version 信令版本（第一个字母：'t'－文本协议，'p'－protobuf协议）
	Version string `json:"version"`
	// AppID AppID
	AppID string `json:"appid"`
	// Name 信令名，用'/'分隔多级信令（用于路由），例如：msg/foo/bar
	Name string `json:"name"`
	// Data 网关信令数据
	Data json.RawMessage `json:"data,omitempty"`
	// Payload 业务数据
	Payload json.RawMessage `json:"payload,omitempty"`
	// Buffer 缓存
	Buffer *bytes.Buffer `json:"-"`
}

func init() {
	serialize.Register(serializer)
}

type engine struct {
}

// NewParseEngine 新建ParseEngine
func NewParseEngine() serialize.ParseEngine {
	return &engine{}
}

// Parse 解析
func (e *engine) Parse(br *bufio.Reader) (cmd *protocol.Command, err error) {
	dec := json.NewDecoder(br)

	var jsonCmd Command
	err = dec.Decode(&jsonCmd)
	if err != nil {
		glog.Warningln("protocol::serialize::alljson::Parse() error:", err)
		return
	}
	return CopyCommand(&jsonCmd)
}

// Close 关闭
func (e *engine) Close() error {
	return nil
}

// CopyCommand 复制命令
func CopyCommand(jsonCmd *Command) (cmd *protocol.Command, err error) {
	cmd = &protocol.Command{
		Version: jsonCmd.Version,
		AppID:   jsonCmd.AppID,
		Name:    jsonCmd.Name,
	}
	if jsonCmd.Payload != nil && len(jsonCmd.Payload) > 2 {
		cmd.Payload = []byte(jsonCmd.Payload)[1 : len(jsonCmd.Payload)-1]
	}
	if jsonCmd.Data != nil {
		if err = cmd.ParseData([]byte(jsonCmd.Data)); err != nil {
			glog.Warningln("protocol::serialize::alljson::CopyCommand() error:", err)
			return nil, err
		}
	}
	return
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	var err error
	cmd.Version = Version
	buf := new(bytes.Buffer)
	buf.Write([]byte(`{"version":"`))
	buf.WriteString(cmd.Version)
	buf.Write([]byte(`","appid":"`))
	buf.WriteString(cmd.AppID)
	buf.Write([]byte(`","name":"`))
	buf.WriteString(cmd.Name)
	buf.WriteByte('"')
	if cmd.Data != nil {
		buf.Write([]byte(`,"data":`))
		var data []byte
		if data, err = json.Marshal(cmd.Data); err != nil {
			glog.Warningln("protocol::serialize::alljson::Compose() JSON Marshal error:", err)
			return nil, err
		}
		buf.Write(bytes.TrimRight(data, "\r\n"))
	}
	if cmd.Payload != nil {
		buf.Write([]byte(`,"payload":"`))
		buf.Write(cmd.Payload)
		buf.Write([]byte(`"}`))
	} else {
		buf.WriteByte('}')
	}
	return buf.Bytes(), nil
}
