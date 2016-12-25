// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/zhangpeihao/zim/pkg/gateway"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
	"github.com/zhangpeihao/zim/pkg/websocket"
	"time"
)

var (
	cfgWebSocketBindAddress string
	cfgPushBindAddress      string
	cfgKey                  string
	cfgRouterJSON           string
)

// gatewayCmd gateway命令
var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "网关服务",
	Long: `提供TCP, UDP和WebSocket等多种协议的接入网关服务

客户端使用指定协议与网关建立连接，并保持连接的活跃。客户端发送
请求到网关，网关通过HTTP请求将服务路由到指定的应用服务上，应用
服务的HTTP响应通过连接返回给客户端。
应用服务使用HTTP协议将推送消息给网关，网关将负载信息推送到指定
客户端或者一批客户端`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			err        error
			gatewaySrv util.SafeCloseServer
		)
		// 构建Gateway服务
		gatewaySrv, err = gateway.NewServer(&gateway.ServerParameter{
			ServerParameter: websocket.ServerParameter{
				WebSocketBindAddress: cfgWebSocketBindAddress,
				Debug:                cfgDebug,
			},
			Key: protocol.Key(cfgKey),
			JSONRouteFile:        cfgRouterJSON,
		})
		if err != nil {
			glog.Errorln("Gateway.NewServer() error:", err)
			return
		}

		// 构建安全退出对象
		closer := util.NewSafeCloser()
		if err = gatewaySrv.Run(closer); err != nil {
			glog.Errorln("gateway server run error:", err)
			return
		}

		// 等待退出信号，并安全退出
		closeTimeout := time.Second * time.Duration(10)
		closer.WaitAndClose(closeTimeout,
			func() {
				gatewaySrv.Close(closeTimeout)
			})
	},
}

func init() {
	RootCmd.AddCommand(gatewayCmd)

	gatewayCmd.PersistentFlags().StringVar(&cfgWebSocketBindAddress, "ws-bind", ":8870", "WebSocket服务绑定地址")
	gatewayCmd.PersistentFlags().StringVar(&cfgPushBindAddress, "push-bind", ":8871", "推送服务绑定地址")
	gatewayCmd.PersistentFlags().StringVar(&cfgKey, "key", "1234567890", "客户端Token验证密钥")
	gatewayCmd.PersistentFlags().StringVar(&cfgRouterJSON, "router-json", "", "JSON router file.")

}
