// Copyright © 2016 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"fmt"
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
				WSBindAddress:  cfgWebSocketBindAddress,
				WSSBindAddress: cfgWssBindAddress,
				Debug:          cfgDebug,
				CertFile:       cfgCertFile,
				KeyFile:        cfgKeyFile,
			},
			PushHTTPServerParameter: httpserver.PushHTTPServerParameter{
				BindAddress: cfgPushBindAddress,
				Debug:       cfgDebug,
			},
			Key:        protocol.Key(cfgKey),
			AppConfigs: cfgAppConfigs,
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
	cobra.OnInitialize(initGatewayConfig)

	gatewayCmd.PersistentFlags().StringVar(&cfgWebSocketBindAddress, "ws-bind", ":8870", "WebSocket服务绑定地址")
	gatewayCmd.PersistentFlags().StringVar(&cfgWssBindAddress, "wss-bind", ":8872", "WebSocket加密服务绑定地址")
	gatewayCmd.PersistentFlags().StringVar(&cfgPushBindAddress, "push-bind", ":8871", "推送服务绑定地址")
	gatewayCmd.PersistentFlags().StringSliceVar(&cfgAppConfigs, "app-config", nil, "应用配置文件.")
	gatewayCmd.PersistentFlags().StringVar(&cfgCertFile, "wss-cert-file", "", "WebSocket加密服务证书文件路径")
	gatewayCmd.PersistentFlags().StringVar(&cfgKeyFile, "wss-key-file", "", "WebSocket加密服务密钥文件路径")

}

func initGatewayConfig() {
	fmt.Println("initGatewayConfig")
	initConfig()

	if viper.InConfig("gateway") {
		cfgAppConfigs = viper.GetStringSlice("gateway.app-config")
		cfgCertFile = viper.GetString("gateway.wss-cert-file")
		cfgKeyFile = viper.GetString("gateway.wss-key-file")
	}
}
