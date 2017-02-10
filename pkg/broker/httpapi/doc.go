// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

/*
Package httpapi HTTP接口形式的Broker，用于简单架构

tag: 作为请求URL的path最后一段，例如：HTTP服务的请求地址是"http://localhost/test"，那么，tag为"push"将被路由到"http://localhost/test/push"

Command.Payload: 作为POST内容发送


HTTP请求Header设置：

* Agent: zim

* Zim-Appid: <APPID>

* Zim-Name: <信令名>

* Zim-Data: <信令Data>，如果Command.Data == nil,则不设置

* Zim-Payloadmd5: <MD5(Payload)>

* Zim-Nonce: 随机数（最大长度128个字符）

* Zim-Timestamp: 当前UTC时间戳，从1970年1月1日0点0 分0 秒开始到现在的秒数(String)

* Zim-Checksum: SHA256(AppSecret + tag + AppID + Name + Data + PayloadMD5 + Nonce + Timestamp)，进行SHA256哈希计算，转化成16进制字符(String，大写)

  CheckSum有效期：出于安全性考虑，每个checkSum的有效期为5分钟(用Timestamp计算)，建议每次请求都生成新的checkSum，同时请确认发起请求的服务器是与标准时间同步的，比如有NTP服务。
  CheckSum检验失败时会返回414错误码，具体参看code状态表。

*/
package httpapi
