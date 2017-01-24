// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"io"
	"io/ioutil"
)

// ParseReader 使用Reader解析信令
func ParseReader(r io.Reader) (cmd *protocol.Command, err error) {
	if r == nil {
		return nil, define.ErrInvalidParameter
	}
	var buf []byte
	buf, err = ioutil.ReadAll(r)
	if err != nil {
		glog.Warningln("protocol::serialize::ParseReader() json.Unmarshal error:", err)
		return nil, err
	}
	return Parse(buf)
}

// Parse 解析信令
func Parse(message []byte) (cmd *protocol.Command, err error) {
	if message == nil || len(message) == 0 {
		glog.Warningln("protocol::serialize::Parse() message is empty")
		return nil, define.ErrInvalidParameter
	}
	serializer, ok := probeByteRegisters[message[0]]
	if !ok {
		return nil, define.ErrUnsupportProtocol
	}
	return serializer.Parse(message)
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	if cmd == nil {
		glog.Warningln("protocol::serialize::Compose() cmd is empty")
		return nil, define.ErrInvalidParameter
	}
	serializer, ok := versionRegisters[cmd.Version]
	if !ok {
		return nil, define.ErrUnsupportProtocol
	}
	return serializer.Compose(cmd)
}
