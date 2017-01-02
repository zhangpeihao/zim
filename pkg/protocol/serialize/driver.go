// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"errors"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize/alljson"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize/plaintext"
	"io"
	"io/ioutil"
)

var (
	// ErrEmptyMessage 消息体为空
	ErrEmptyMessage = errors.New("empty message")
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
	if message == nil || len(message) == 0 {
		glog.Warningln("protocol::serialize::ParseReader() message is empty")
		err = define.ErrInvalidParameter
		return
	}
	switch message[0] {
	case plaintext.ProbeByte:
		return plaintext.Parse(message)
	case alljson.ProbeByte:
		return alljson.Parse(message)
	default:
		err = define.ErrUnsupportProtocol
	}
	return
}

// Compose 将信令编码
func Compose(cmd *protocol.Command) ([]byte, error) {
	if cmd == nil {
		return nil, define.ErrInvalidParameter
	}
	switch cmd.Version {
	case plaintext.Version:
		return plaintext.Compose(cmd)
	case alljson.Version:
		return alljson.Compose(cmd)
	default:
		return nil, define.ErrUnsupportProtocol
	}
}
