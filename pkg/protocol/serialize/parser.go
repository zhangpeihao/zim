// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"bufio"
	"bytes"
	"io"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

// ParseEngine 解析引擎函数
type ParseEngine interface {
	Parse(br *bufio.Reader) (cmd *protocol.Command, err error)
	Close() error
}

// Parser 解析器
type Parser struct {
	// reader bufio reader
	reader *bufio.Reader
	// engine 解析引擎
	engine ParseEngine
	// probeByte 解析用字节
	probeByte byte
}

// NewParser 新建解析器
func NewParser(r io.Reader) (parser *Parser) {
	return &Parser{
		reader: bufio.NewReader(r),
	}
}

// ReadCommand 将数据写入解析器，如果能够解析出命令，则回调
func (parser *Parser) ReadCommand() (cmd *protocol.Command, err error) {
	var (
		serializer *Serializer
		ok         bool
		probeByte  []byte
	)

	if parser.engine == nil {
		// 重新构建
		probeByte, err = parser.reader.Peek(1)
		if err != nil {
			return
		}

		if probeByte == nil || len(probeByte) == 0 {
			glog.Warningln("protocol::serialize::Parser::ReadFrom() probe byte is empty")
			return nil, io.EOF
		}
		serializer, ok = probeByteRegisters[probeByte[0]]
		if !ok {
			err = define.ErrUnsupportProtocol
			glog.Warningf("protocol::serialize::Parser::ReadFrom() Unsupport probe byte: 0X%2X\n", probeByte)
			return
		}
		parser.engine = serializer.NewParseEngine()
	}

	return parser.engine.Parse(parser.reader)
}

// Close 关闭
func (parser *Parser) Close() error {
	if parser.engine != nil {
		parser.engine.Close()
		parser.engine = nil
	}
	return nil
}

// Parse 通过字节解析
func Parse(data []byte) (cmd *protocol.Command, err error) {
	buf := bytes.NewBuffer(data)
	parser := NewParser(buf)
	return parser.ReadCommand()
}
