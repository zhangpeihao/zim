// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"github.com/zhangpeihao/zim/pkg/util"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ws       = flag.String("ws", "ws://127.0.0.1:8870/ws", "The server WebSocket address.")
	appid    = flag.String("appid", "test", "The appid.")
	key      = flag.String("key", "1234567890", "The token key.")
	number   = flag.Int("number", 1, "The number of connections.")
	baseID   = flag.Int("base-id", 1, "The base ID of connections.")
	interval = flag.Int("interval", 5, "The interval time of send message (in second).")
)

var (
	closeGate      *sync.WaitGroup
	errorCounter   int32
	receiveCounter int32
	sendCounter    int32
	exit           bool
)

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	closeGate = new(sync.WaitGroup)

	for i := 0; i < *number; i++ {
		go loop(i + (*baseID))
	}

	terminationSignalsCh := make(chan os.Signal, 1)
	util.WaitAndClose(terminationSignalsCh, time.Second*time.Duration(5), func() {
		exit = true
	})
	fmt.Println("Wait close gate done")

	closeGate.Wait()
	summary()
}

func loop(id int) {
	closeGate.Add(1)
	defer closeGate.Done()

	idstr := strconv.Itoa(id)
	now := time.Now().Unix()
	tokenKey := protocol.Key([]byte(*key))
	loginCmd := &protocol.GatewayLoginCommand{
		UserID:    idstr,
		DeviceID:  "web",
		Timestamp: now,
		Token:     "",
	}
	loginCmd.Token = tokenKey.Token(loginCmd)

	cmd := &protocol.Command{
		Version: "t1",
		AppID:   *appid,
		Name:    "login",
		Data:    loginCmd,
		Payload: []byte(fmt.Sprintf(`{"id":"%d","message":"foo bar"}`, id)),
	}

	dialer := &websocket.Dialer{
		Proxy: http.ProxyFromEnvironment,
	}
	c, _, err := dialer.Dial(*ws, nil)
	if err != nil {
		log.Printf("client[%d] Dial error: %s\n", id, err)
		atomic.AddInt32(&errorCounter, 1)
		return
	}

	// Login
	message, err := serialize.Compose(cmd)
	if err != nil {
		log.Println("serialize.Compose error:", err)
		atomic.AddInt32(&errorCounter, 1)
		return
	}
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Println("login error:", err)
		atomic.AddInt32(&errorCounter, 1)
		return
	}

	cmd.Name = "msg"
	message, err = serialize.Compose(cmd)
	if err != nil {
		log.Println("serialize.Compose error:", err)
		atomic.AddInt32(&errorCounter, 1)
		return
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		defer c.Close()
		for !exit {
			_, message, err := c.ReadMessage()
			if err != nil {
				if !exit {
					log.Println("read:", err)
					atomic.AddInt32(&errorCounter, 1)
				}
				return
			}
			log.Printf("recv: %s", message)
			atomic.AddInt32(&receiveCounter, 1)
		}
	}()

	ticker := time.NewTicker(time.Second * time.Duration(*interval))
	defer ticker.Stop()

	for !exit {
		select {
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				if !exit {
					log.Println("write:", err)
					atomic.AddInt32(&errorCounter, 1)
				}
				return
			}
			atomic.AddInt32(&sendCounter, 1)
		case <-done:
			c.Close()
			return
		}
	}
}

func summary() {
	fmt.Println("error, send, receive")
	fmt.Printf("%d, %d, %d\n", errorCounter, sendCounter, receiveCounter)
}
