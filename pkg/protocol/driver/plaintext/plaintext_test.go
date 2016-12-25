// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package plaintext

import (
	"encoding/json"
	"github.com/pkg/errors"
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
msg/foo/bar
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"msg/foo/bar",
				&protocol.GatewayMessageCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
login
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"login",
				&protocol.GatewayLoginCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
close
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"close",
				&protocol.GatewayCloseCommand{},
				[]byte("foo bar"),
			},
		},
	}

	for index, testCase := range testCases {
		cmd, err := Parse(testCase.Message)
		if err != nil {
			t.Errorf("TestPlainText Case[%d]\nParse %s error: %s",
				index, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestPlainText Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index, testCase.Message, cmd, testCase.ExpectCommand)
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
msg/foo/bar
{"id":"","timestamp":0,"token":""}
foo bar`),
			define.ErrUnsupportProtocol,
		},
		{
			[]byte(`t1
msg/foo/bar`),
			protocol.ErrParseFailed,
		},
		{
			[]byte(`t1
msg/foo/bar
{"Format error JSON
foo bar`),
			ErrJSONError,
		},
		{
			[]byte(`t1
login
{"Format error JSON
foo bar`),
			ErrJSONError,
		},
		{
			[]byte(`t1
close
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
					index+1, testCase.Message, ErrJSONError, testCase.Error)
			}
		} else if err != testCase.Error {
			t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
				index+1, testCase.Message, reflect.TypeOf(err), testCase.Error)
		}
	}
}
