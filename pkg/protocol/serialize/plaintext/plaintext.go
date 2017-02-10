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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
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
	serializer = &serialize.Serializer{
		Version:        Version,
		ProbeByte:      ProbeByte,
		NewParseEngine: NewParseEngine,
		Compose:        Compose,
	}
)

type engine struct {
	lines      [CommandLines][]byte
	linesIndex int
}

func init() {
	serialize.Register(serializer)
}

// NewParseEngine 新建解析器
func NewParseEngine() serialize.ParseEngine {
	return &engine{}
}

// Parse 解析
func (e *engine) Parse(br *bufio.Reader) (cmd *protocol.Command, err error) {
	for e.linesIndex < CommandLines {
		// 一行一行读取
		var line []byte
		line, err = br.ReadBytes('\n')
		if err != nil {
			glog.Warningln("protocol::serialize::plaintext::Parse() error:", err)
			return
		}
		e.lines[e.linesIndex] = append(e.lines[e.linesIndex], line...)
		if line[len(line)-1] == '\n' {
			// 完整
			e.linesIndex++
		}
	}
	defer e.reset()
	cmd = &protocol.Command{
		Version: strings.Trim(string(e.lines[CommandVersionLine]), "\r\t\n "),
		AppID:   strings.Trim(string(e.lines[CommandAppIDLine]), "\r\t\n "),
		Name:    strings.Trim(string(e.lines[CommandNameLine]), "\r\t\n "),
	}
	if len(cmd.Version) == 0 || cmd.Version[0] != ProbeByte {
		err = define.ErrUnsupportProtocol
		fmt.Println("define.ErrUnsupportProtocol")
		return
	}
	cmd.Payload = bytes.Trim(e.lines[CommandPayloadLine], "\r\n")

	if err = cmd.ParseData(e.lines[CommandDataLine]); err != nil {
		fmt.Println("cmd.ParseData error:", err)
		return
	}
	return
}

// Close 关闭
func (e *engine) Close() error {
	e.reset()
	return nil
}

func (e *engine) reset() {
	for i := 0; i < CommandLines; i++ {
		e.lines[i] = nil
	}
	e.linesIndex = 0
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
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}
