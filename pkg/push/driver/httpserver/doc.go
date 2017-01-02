// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package httpserver HTTP推送服务

通过HTTP Web API形式提供推送服务。接口数据定义参考package protocol

HTTP头信息包含以下参数：

zim-AppID: 应用ID（开发者平台分配的app ID）
zim-Nonce: 随机数（最大长度128个字符）
zim-Timestamp: 当前UTC时间戳，从1970年1月1日0点0 分0 秒开始到现在的秒数(String)
zim-CheckSum: SHA256(AppSecret + Nonce + Timestamp),三个参数拼接的字符串，进行SHA256哈希计算，转化成16进制字符(String，小写)

  CheckSum有效期：出于安全性考虑，每个checkSum的有效期为5分钟(用Timestamp计算)，建议每次请求都生成新的checkSum，同时请确认发起请求的服务器是与标准时间同步的，比如有NTP服务。
  CheckSum检验失败时会返回414错误码，具体参看code状态表。

push2user: 推送消息给指定用户
*/
package httpserver
