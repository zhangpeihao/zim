// Copyright 2016-2017 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package app 应用配置（包含路由和CheckSum计算）

收到IM信令后，根据信令名（例如：msg/foo/bar），查询信息的处理路由。
可以提供多种路由设置与维护方式，包括：本地配置文件、Redis服务、etcd服务和consul服务。
*/
package app
