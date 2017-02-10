// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package alljson

import (
	"bufio"
	"bytes"
	"flag"
	"testing"

	"github.com/zhangpeihao/zim/pkg/protocol"
)

type JSONData struct {
	Data interface{} `json:"data"`
}
type TestCopyCase struct {
	SrcJSON string
	Dest    interface{}
	Expect  interface{}
}

type TestCase struct {
	Message       []byte
	ExpectCommand protocol.Command
}

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestAllJson(t *testing.T) {
	testCases := []TestCase{
		{
			[]byte(`{"version":"j1","appid":"test","name":"msg/foo/bar","data":{"userid":""},"payload":"foo bar"}`),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "msg/foo/bar",
				Data:    &protocol.GatewayMessageCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":"login","data":{"userid":"","deviceid":"","timestamp":0,"token":""},"payload":"foo bar"}`),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "login",
				Data:    &protocol.GatewayLoginCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":"close","data":{"userid":""},"payload":"foo bar"}`),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "close",
				Data:    &protocol.GatewayCloseCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":"p2u","data":{},"payload":"foo bar"}`),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "p2u",
				Data:    &protocol.Push2UserCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":"hb","payload":"foo bar"}`),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "hb",
				Data:    nil,
				Payload: []byte("foo bar"),
			},
		},
	}

	engine := NewParseEngine()

	for index, testCase := range testCases {
		cmd, err := engine.Parse(bufio.NewReader(bytes.NewBuffer(testCase.Message)))
		if err != nil {
			t.Errorf("TestAllJson Case[%d]\nParse %s error: %s",
				index+1, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestAllJson Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index+1, testCase.Message, cmd, testCase.ExpectCommand)
		}

		buf, err := Compose(cmd)
		if err != nil {
			t.Errorf("TestAllJson Case[%d] error: %s\n", index+1, err)
		}
		if bytes.Compare(buf, testCase.Message) != 0 {
			t.Errorf("TestAllJson Case[%d]\nCompose %s\nGot: %s,\nExpect: %s",
				index+1, cmd, buf, testCase.Message)
		}
	}
}

type TestErrorCase struct {
	Message []byte
}

func TestError(t *testing.T) {
	testCases := []TestErrorCase{
		{
			[]byte(`{"version":"j1","appid":"test","name":`),
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":"msg/foo/bar","data":{"userid","payload":"foo bar"}`),
		},
	}
	engine := NewParseEngine()

	for index, testCase := range testCases {
		_, err := engine.Parse(bufio.NewReader(bytes.NewBuffer(testCase.Message)))
		if err == nil {
			t.Errorf("TestError Case[%d]\nParse %s\nshould return error",
				index+1, testCase.Message)
		}
	}
}
