// Copyright © 2017 Zhang Peihao <zhangpeihao@gmail.com>
//

package cmd

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/cobra"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"github.com/zhangpeihao/zim/pkg/util"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

func init() {
	RootCmd.AddCommand(stressCmd)

	stressCmd.PersistentFlags().StringVar(&cfgWebSocketURL, "ws-url", "ws://127.0.0.1:8870/ws", "WebSocket服务URL")
	stressCmd.PersistentFlags().StringVar(&cfgAppID, "appid", "test", "App ID")
	stressCmd.PersistentFlags().StringVar(&cfgKey, "key", "1234567890", "客户端Token验证密钥")
	stressCmd.PersistentFlags().UintVar(&cfgNumber, "number", 100, "连接数")
	stressCmd.PersistentFlags().UintVar(&cfgBase, "base", 1, "起始ID")
	stressCmd.PersistentFlags().UintVar(&cfgInterval, "interval", 10, "消息发送间隔时间（单位：秒）")

	stressCmd.PersistentFlags().StringVar(&cfgInfluxdbAddress, "influxdb-address", "http://localhost:8086", "influxDB地址")
	stressCmd.PersistentFlags().StringVar(&cfgInfluxdbDB, "influxdb-db", "stress", "influxDB 数据库名")
	stressCmd.PersistentFlags().StringVar(&cfgInfluxdbUser, "influxdb-user", "root", "influxDB用户名")
	stressCmd.PersistentFlags().StringVar(&cfgInfluxdbPassword, "influxdb-password", "root", "influxDB密码")
	stressCmd.PersistentFlags().UintVar(&cfgInfluxdbInterval, "influxdb-interval", 60, "influxDB数据消息发送间隔时间（单位：秒）")
}

// stressCmd represents the stress command
var stressCmd = &cobra.Command{
	Use:   "stress",
	Short: "压力测试客户端",
	Long: `压力测试客户端

按照设定的连接数，发送消息体大小，发送频率，向指定服务器
建立连接，发送消息，并校验接受数据。`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfgInterval == 0 {
			return errors.New("Interval必须大于0")
		}

		if cfgNumber == 0 || cfgNumber > 65500 {
			return errors.New("客户端数量必须大于0，小于65500")
		}

		gInterval = time.Second * time.Duration(cfgInterval)
		gInfluxdbInterval = time.Second * time.Duration(cfgInfluxdbInterval)

		go stressInfluxDBLoop()

		// 启动客户端
		for i := uint(0); i < cfgNumber; i++ {
			go stressLoop(i + cfgBase)
		}

		terminationSignalsCh := make(chan os.Signal, 1)
		util.WaitAndClose(terminationSignalsCh, time.Second*time.Duration(3), func() {
			SetExitFlag()
		})
		fmt.Println("Wait close gate done")

		gCloseGate.Wait()
		stressSummary(nil)
		return nil
	},
}

func stressInfluxDBLoop() {
	gCloseGate.Add(1)
	defer gCloseGate.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	now := time.Now()
	for !IsExit() {
		stressSummary(&now)
		now = <-ticker.C
	}
}

func stressLoop(id uint) {
	gCloseGate.Add(1)
	defer gCloseGate.Done()

	idstr := strconv.Itoa(int(id))
	now := time.Now().Unix()
	tokenKey := protocol.Key([]byte(cfgKey))
	loginCmd := &protocol.GatewayLoginCommand{
		UserID:    idstr,
		DeviceID:  "web",
		Timestamp: now,
		Token:     "",
	}
	loginCmd.Token = tokenKey.Token(loginCmd)

	cmd := &protocol.Command{
		Version: "t1",
		AppID:   cfgAppID,
		Name:    "login",
		Data:    loginCmd,
		Payload: []byte(fmt.Sprintf(`{"id":"%d","message":"foo bar"}`, id)),
	}

	dialer := &websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}
	c, _, err := dialer.Dial(cfgWebSocketURL, nil)
	if err != nil {
		log.Printf("client[%d] Dial error: %s\n", id, err)
		CountError()
		return
	}

	// Login
	message, err := serialize.Compose(cmd)
	if err != nil {
		log.Printf("client[%d] serialize.Compose error: %s\n", id, err)
		CountError()
		return
	}
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("client[%d] login error: %s\n", id, err)
		CountError()
		return
	}

	// 消息
	cmd.Name = "msg"
	message, err = serialize.Compose(cmd)
	if err != nil {
		log.Printf("client[%d] serialize.Compose error: %s\n", id, err)
		CountError()
		return
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		defer c.Close()
		for !IsExit() {
			t, message, err := c.ReadMessage()
			if err != nil {
				if !IsExit() {
					log.Printf("client[%d] read: %s\n", id, err)
					CountError()
				}
				return
			}
			if !stressCheck(t, message) {
				CountCheckError()
			}
			CountReceive()
		}
	}()

	ticker := time.NewTicker(gInterval)
	defer ticker.Stop()

	for !IsExit() {
		select {
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				if !IsExit() {
					log.Printf("client[%d] write: %s\n", id, err)
					CountCheckError()
				}
				return
			}
			CountSend()
		case <-done:
			c.Close()
			return
		}
	}
}

func stressCheck(int, []byte) bool {
	return true
}

func stressSummary(now *time.Time) error {
	if now == nil {
		sNow := time.Now()
		now = &sNow
	} else if gPreTime != nil && (*gPreTime).Add(gInfluxdbInterval).After(*now) {
		return nil
	}
	gPreTime = now
	// Make client
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     cfgInfluxdbAddress,
		Username: cfgInfluxdbUser,
		Password: cfgInfluxdbPassword,
	})
	if err != nil {
		return err
	}
	defer c.Close()

	// Create a new point batch
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  cfgInfluxdbDB,
		Precision: "s",
	})

	errorCounter := atomic.LoadInt32(&gErrorCounter)
	recevieCounter := atomic.LoadInt32(&gReceiveCounter)
	sendCounter := atomic.LoadInt32(&gSendCounter)
	checkErrorCounter := atomic.LoadInt32(&gCheckErrorCounter)

	log.Printf("errorCounter: %d, recevieCounter: %d, sendCounter: %d, checkErrorCounter: %d\n",
		errorCounter, recevieCounter, sendCounter, checkErrorCounter)

	atomic.AddInt32(&gErrorCounter, -errorCounter)
	atomic.AddInt32(&gReceiveCounter, -recevieCounter)
	atomic.AddInt32(&gSendCounter, -sendCounter)
	atomic.AddInt32(&gCheckErrorCounter, -checkErrorCounter)

	// Create a point and add to batch
	tags := map[string]string{"stress": "client"}
	fields := map[string]interface{}{
		"error":      errorCounter,
		"receive":    recevieCounter,
		"send":       sendCounter,
		"checkerror": checkErrorCounter,
	}
	pt, err := client.NewPoint("counter", tags, fields, *now)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	return c.Write(bp)
}
