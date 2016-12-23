// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package define

// Connection 连接接口
type Connection interface {
	Close(force bool) error
	String() string
}
