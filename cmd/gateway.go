// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	cfgWebSocketBindAddress string
	cfgPushBindAddress string
)

// gatewayCmd gateway命令
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "网关服务",
	Long: `提供TCP, UDP和WebSocket等多种协议的接入网关服务

客户端使用指定协议与网关建立连接，并保持链接的活跃。客户端发送
请求到网关，网关通过HTTP请求将服务路由到指定的应用服务上，应用
服务的HTTP响应通过连接返回给客户端。
应用服务使用HTTP协议将推送消息给网关，网关将负载信息推送到指定
客户端或者一批客户端`,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("gateway called")
	},
}

func init() {
	RootCmd.AddCommand(gatewayCmd)

	gatewayCmd.PersistentFlags().StringVar(&cfgWebSocketBindAddress, "ws-bind", ":8870", "WebSocket服务绑定地址")
	gatewayCmd.PersistentFlags().StringVar(&cfgPushBindAddress, "push-bind", ":8871", "推送服务绑定地址")

}
