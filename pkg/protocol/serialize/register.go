// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"github.com/zhangpeihao/zim/pkg/protocol"
)

// Serializer 串行化函数，所有实现都必须完成以下函数
type Serializer struct {
	// Version 串行化版本
	Version string
	// ProbeByte 协议首字节
	ProbeByte byte
	// Parse 解析信令
	Parse func(message []byte) (cmd *protocol.Command, err error)
	// Compose 将信令编码
	Compose func(cmd *protocol.Command) ([]byte, error)
}

var (
	// 注册列表
	probeByteRegisters = make(map[byte]*Serializer)
	versionRegisters   = make(map[string]*Serializer)
)

// Register 注册串行化新建函数
func Register(serializer *Serializer) {
	probeByteRegisters[serializer.ProbeByte] = serializer
	versionRegisters[serializer.Version] = serializer
}
