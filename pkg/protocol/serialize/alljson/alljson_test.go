// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package alljson

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"reflect"
	"testing"
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

var (
	base64Payload = base64.StdEncoding.EncodeToString([]byte("foo bar"))
)

func TestAllJson(t *testing.T) {
	testCases := []TestCase{
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"msg/foo/bar","data":{"userid":""},"payload":"%s"}`, base64Payload)),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "msg/foo/bar",
				Data:    &protocol.GatewayMessageCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"login","data":{"userid":"","deviceid":"","timestamp":0,"token":""},"payload":"%s"}`, base64Payload)),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "login",
				Data:    &protocol.GatewayLoginCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"close","data":{"userid":""},"payload":"%s"}`, base64Payload)),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "close",
				Data:    &protocol.GatewayCloseCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"p2u","data":{},"payload":"%s"}`, base64Payload)),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "p2u",
				Data:    &protocol.Push2UserCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"hb","payload":"%s"}`, base64Payload)),
			protocol.Command{
				Version: "j1",
				AppID:   "test",
				Name:    "hb",
				Data:    nil,
				Payload: []byte("foo bar"),
			},
		},
	}

	for index, testCase := range testCases {
		cmd, err := Parse(testCase.Message)
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
	Error   error
}

var ErrJSONError = errors.New("json error")

func TestError(t *testing.T) {
	testCases := []TestErrorCase{
		{
			[]byte(fmt.Sprintf(`{"version":"T1","appid":"test","name":"msg/foo/bar","data":{"userid":""},"payload":"%s"}`, base64Payload)),
			define.ErrUnsupportProtocol,
		},
		{
			[]byte(fmt.Sprintf(`"appid":"test","name":"msg/foo/bar","data":{"userid":""},"payload":"%s"}`, base64Payload)),
			define.ErrUnsupportProtocol,
		},
		{
			[]byte(`{"version":"j1","appid":"test","name":`),
			ErrJSONError,
		},
		{
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"msg/foo/bar","data":{"userid","payload":"%s"}`, base64Payload)),
			ErrJSONError,
		},
	}
	for index, testCase := range testCases {
		_, err := Parse(testCase.Message)
		if err == nil {
			t.Errorf("TestError Case[%d]\nParse %s\nNo error\nExpect: %s",
				index+1, testCase.Message, testCase.Error)
		} else if reflect.TypeOf(err) == reflect.TypeOf(new(json.SyntaxError)) {
			if testCase.Error != ErrJSONError {
				t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
					index+1, testCase.Message, err, testCase.Error)
			}
		} else if err != testCase.Error {
			t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
				index+1, testCase.Message, reflect.TypeOf(err), testCase.Error)
		}
	}
}
