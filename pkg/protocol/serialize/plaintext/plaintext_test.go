// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package plaintext

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"reflect"
	"testing"
)

type TestCase struct {
	Message       []byte
	ExpectCommand protocol.Command
}

func TestPlainText(t *testing.T) {
	testCases := []TestCase{
		{
			[]byte(`t1
test
msg/foo/bar
{"userid":""}
foo bar`),
			protocol.Command{
				Version: "t1",
				AppID:   "test",
				Name:    "msg/foo/bar",
				Data:    &protocol.GatewayMessageCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
login
{"userid":"","deviceid":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				Version: "t1",
				AppID:   "test",
				Name:    "login",
				Data:    &protocol.GatewayLoginCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
close
{"userid":""}
foo bar`),
			protocol.Command{
				Version: "t1",
				AppID:   "test",
				Name:    "close",
				Data:    &protocol.GatewayCloseCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
p2u
{"useridlist":""}
foo bar`),
			protocol.Command{
				Version: "t1",
				AppID:   "test",
				Name:    "p2u",
				Data:    &protocol.Push2UserCommand{},
				Payload: []byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
hb

foo bar`),
			protocol.Command{
				Version: "t1",
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
			t.Errorf("TestPlainText Case[%d]\nParse %s error: %s",
				index+1, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestPlainText Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index+1, testCase.Message, cmd, testCase.ExpectCommand)
		}

		buf, err := Compose(cmd)
		if err != nil {
			t.Errorf("TestPlainText Case[%d] error: %s\n", index+1, err)
		}
		if bytes.Compare(buf, testCase.Message) != 0 {
			t.Errorf("TestPlainText Case[%d]\nCompose %s\nGot: %s,\nExpect: %s",
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
			[]byte(`T1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
foo bar`),
			define.ErrUnsupportProtocol,
		},
		{
			[]byte(`t1
test
msg/foo/bar`),
			protocol.ErrParseFailed,
		},
		{
			[]byte(`t1
test
msg/foo/bar
{"Format error JSON
foo bar`),
			ErrJSONError,
		},
		{
			[]byte(`t1
test
login
{"Format error JSON
foo bar`),
			ErrJSONError,
		},
		{
			[]byte(`t1
test
close
{"Format error JSON
foo bar`),
			ErrJSONError,
		},
		{
			[]byte(`t1
test
p2u
{"Format error JSON
foo bar`),
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
