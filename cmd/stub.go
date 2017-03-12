// Copyright © 2017 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/zhangpeihao/shutdown"
	"github.com/zhangpeihao/zim/pkg/app"
	"github.com/zhangpeihao/zim/pkg/broker/httpapi"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

const (
	// Tag 消息broker tag
	Tag = "gateway"
)

var (
	globalContext context.Context
)

// stubCmd represents the stub command
var stubCmd = &cobra.Command{
	Use:   "stub",
	Short: "测试用桩服务",
	Long: `测试用桩服务

提供桩服务，接收Gateway消息，并回消息`,
	Run: func(cmd *cobra.Command, args []string) {
		appController, err := app.NewController(nil)
		if err != nil {
			log.Fatal("new app controller error:", err)
			return
		}
		appController.AddApp(&app.App{
			ID:       "stub",
			Key:      cfgKey,
			KeyBytes: []byte(cfgKey),
		})
		globalContext = appController.SaveIntoContext(shutdown.NewContext())
		http.HandleFunc("/"+Tag, HandleHTTP)
		listener, err := net.Listen("tcp", cfgStubBindAddress)
		if err != nil {
			log.Fatal("listen error:", err)
			return
		}
		go func() {
			servererr := http.Serve(listener, nil)
			if servererr != nil {
				log.Fatal("Serve error:", err)
				if !IsExit() {
					os.Exit(1)
				}
			}
		}()
		shutdown.WaitAndShutdown(globalContext, time.Second*time.Duration(3), func(timeout time.Duration) error {
			SetExitFlag()
			return nil
		})
	},
}

func init() {
	RootCmd.AddCommand(stubCmd)

	stubCmd.PersistentFlags().StringVar(&cfgStubBindAddress, "stub-addr", ":8880", "service stub绑定地址")
	stubCmd.PersistentFlags().StringVar(&cfgHTTPBrokerURL, "http-broker-url", "http://127.0.0.1:8771", "HTTP API Broker请求URL")
	stubCmd.PersistentFlags().StringVar(&cfgAppID, "appid", "test", "App ID")
	stubCmd.PersistentFlags().StringVar(&cfgKey, "key", "1234567890", "App security key")

}

// HandleHTTP 登入处理
func HandleHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		payload []byte
		err     error
		cmd     *protocol.Command
		ok      bool
	)

	if payload, err = ioutil.ReadAll(r.Body); err != nil {
		glog.Warningf("HandleLogin() Read payload error: %s\n",
			err)
		w.WriteHeader(400)
		return
	}
	if cmd, err = httpapi.ParseCommand(globalContext, Tag, r.Header, payload, 10); err != nil {
		glog.Warningf("HandleLogin() ParseCommand error: %s\n",
			err)
		w.WriteHeader(400)
		return
	}
	glog.Info("got cmd:", cmd)
	switch cmd.FirstPartName() {
	case protocol.Login:
		var loginCmd *protocol.GatewayLoginCommand
		loginCmd, ok = cmd.Data.(*protocol.GatewayLoginCommand)
		if !ok {
			glog.Warningf("HandleLogin() cmd type error: %s\n",
				err)
			w.WriteHeader(400)
			return
		}

		Reponse(w, &protocol.Command{
			AppID: cmd.AppID,
			Name:  protocol.Push2User,
			Data: &protocol.Push2UserCommand{
				Tags: "*",
			},
			Payload: []byte(fmt.Sprintf(`{"from":"%s","msg":"user %s enter"}`,
				"system", loginCmd.UserID)),
		})
	case protocol.Message:
		var msgCmd *protocol.GatewayMessageCommand
		msgCmd, ok = cmd.Data.(*protocol.GatewayMessageCommand)
		if !ok {
			glog.Warningf("HandleLogin() cmd type error: %s\n",
				err)
			w.WriteHeader(400)
			return
		}

		Reponse(w, &protocol.Command{
			AppID: cmd.AppID,
			Name:  protocol.Push2User,
			Data: &protocol.Push2UserCommand{
				Tags: "*",
			},
			Payload: []byte(fmt.Sprintf(`{"from":"%s","msg":"%s"}`,
				msgCmd.UserID, string(cmd.Payload))),
		})
	default:
		glog.Warningf("HandleLogin() unsupport command name: %s\n",
			err)
		w.WriteHeader(400)
	}
}

// Reponse 发响应
func Reponse(w http.ResponseWriter, cmd *protocol.Command) {
	err := httpapi.ComposeCommand(globalContext, Tag, w.Header(), cmd)
	if err != nil {
		glog.Errorf("Reponse() ComposeCommand error: %s\n", err)
		w.WriteHeader(400)
		return
	}
	if len(cmd.Payload) > 0 {
		w.Write(cmd.Payload)
	}
}
