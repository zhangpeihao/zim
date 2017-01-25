// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zhangpeihao/zim/pkg/gateway"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/push/driver/httpserver"
	"github.com/zhangpeihao/zim/pkg/util"
	"github.com/zhangpeihao/zim/pkg/websocket"
	"time"
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
		glog.Infoln("gateway run")
		// 构建Gateway服务
		gatewaySrv, err = gateway.NewServer(&gateway.ServerParameter{
			WSParameter: websocket.WSParameter{
				WSBindAddress:  viper.GetString("gateway.ws-bind"),
				WSSBindAddress: viper.GetString("gateway.wss-bind"),
				Debug:          viper.GetBool("debug"),
				CertFile:       viper.GetString("gateway.wss-cert-file"),
				KeyFile:        viper.GetString("gateway.wss-key-file"),
			},
			PushHTTPServerParameter: httpserver.PushHTTPServerParameter{
				BindAddress: viper.GetString("gateway.push-bind"),
				Debug:       viper.GetBool("debug"),
			},
			Key:        protocol.Key(cfgKey),
			AppConfigs: viper.GetStringSlice("gateway.app-config"),
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

	gatewayCmd.PersistentFlags().String("ws-bind", ":8870", "WebSocket服务绑定地址")
	viper.BindPFlag("gateway.ws-bind", gatewayCmd.PersistentFlags().Lookup("ws-bind"))

	gatewayCmd.PersistentFlags().String("wss-bind", ":8872", "WebSocket加密服务绑定地址")
	viper.BindPFlag("gateway.wss-bind", gatewayCmd.PersistentFlags().Lookup("wss-bind"))

	gatewayCmd.PersistentFlags().String("push-bind", ":8871", "推送服务绑定地址")
	viper.BindPFlag("gateway.push-bind", gatewayCmd.PersistentFlags().Lookup("push-bind"))

	gatewayCmd.PersistentFlags().StringSlice("app-config", nil, "应用配置文件.")
	viper.BindPFlag("gateway.app-config", gatewayCmd.PersistentFlags().Lookup("app-config"))

	gatewayCmd.PersistentFlags().String("wss-cert-file", "", "WebSocket加密服务证书文件路径")
	viper.BindPFlag("gateway.wss-cert-file", gatewayCmd.PersistentFlags().Lookup("wss-cert-file"))

	gatewayCmd.PersistentFlags().String("wss-key-file", "", "WebSocket加密服务密钥文件路径")
	viper.BindPFlag("gateway.wss-key-file", gatewayCmd.PersistentFlags().Lookup("wss-key-file"))
}
