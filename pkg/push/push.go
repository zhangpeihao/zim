// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package push 推送消息到客户端

业务服务器通过向推送服务发送消息，向向客户端发送消息。
*/
package push

// Message 推动数据
type Message struct {
	// AppID 应用ID
	AppID string
	// UserID 用户ID
	UserID string
	// CommandName 命令名
	CommandName string
	// Payload 负载内容
	Payload []byte
}

// Handler 推送处理回调接口
type Handler interface {
	// OnPushToUser 向指定用户推送消息
	OnPushToUser(data *Message)
}
