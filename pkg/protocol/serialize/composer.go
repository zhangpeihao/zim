// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

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
