// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

import "github.com/zhangpeihao/zim/pkg/protocol"

// Connection 连接接口
type Connection interface {
	ID() string
	AppID() string
	UserID() string
	DeviceID() string
	LoginSuccess(appid, userid, device string)
	IsLogin() bool
	Close(force bool) error
	String() string
	Send(cmd *protocol.Command) error
}

// ConnectionID 通过AppID和UserID组合成ConnectionID
func ConnectionID(appid, userid string) string {
	return appid + "#" + userid
}
