// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

const (
	// Login 登入
	Login = "login"
	// Close 关闭
	Close = "close"
	// Message 消息
	Message = "msg"
	// HeartBeat 心跳
	HeartBeat = "hb"
	// HeartBeatResponse 心跳响应
	HeartBeatResponse = "hbr"
)

const (
	// CommandNameLine 命令名行索引
	CommandNameLine = 0
	// CommandDataLine 命令数据行索引
	CommandDataLine = 1
	// CommandPayloadLine 命令负载行索引
	CommandPayloadLine = 2
	// CommandLines 命令行数
	CommandLines = 3
)

var (
	// CommandSep 命令分割字符
	CommandSep = []byte{'\n'}
)

// Command 信令
type Command struct {
	// Name 信令名
	Name string
	// Data 网关信令数据
	Data interface{}
	// Payload 业务数据
	Payload string
}
