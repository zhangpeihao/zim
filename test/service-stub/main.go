// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package main

import (
	"flag"
	"fmt"
	"github.com/zhangpeihao/zim/pkg/invoker/driver/httpapi"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/serialize"
	"log"
	"net/http"
)

var (
	addr  = flag.String("Address", ":8880", "The bind address of web API request.")
	appid = flag.String("appid", "test", "The appid.")
)

var (
	userIDQueue = make(chan string, 1000)
	p2uCommand  = `t1
test
p2u
{"useridlist":"%s"}
foo bar
`
)

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	userIDQueue <- "1"

	http.HandleFunc("/login", HandleLogin)
	http.HandleFunc("/msg", HandleMsg)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("error:", err)
	}
}

var (
	loginResponse = &protocol.Command{
		Version: "t1",
		AppID:   *appid,
		Name:    "login",
		Data:    nil,
		Payload: []byte("foo bar"),
	}
)

// HandleLogin 登入处理
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	msg, err := serialize.Compose(loginResponse)
	if err != nil {
		w.WriteHeader(500)
	} else {
		w.Write(msg)
	}
}

// HandleMsg 消息处理
func HandleMsg(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(httpapi.HeaderUserID)
	appID := r.Header.Get(httpapi.HeaderAppID)

	if len(userID) == 0 {
		w.WriteHeader(400)
		log.Printf("HandleMsg() plaintext.ParseReader no %s header\n", httpapi.HeaderUserID)
		return
	}
	if len(appID) == 0 {
		w.WriteHeader(400)
		log.Printf("HandleMsg() plaintext.ParseReader no %s header\n", httpapi.HeaderAppID)
		return
	}

	userIDQueue <- userID
	toUserID := <-userIDQueue
	w.Write([]byte(fmt.Sprintf(p2uCommand, toUserID)))
}
