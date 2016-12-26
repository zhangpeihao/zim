// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package plaintext

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	pushprotocol "github.com/zhangpeihao/zim/pkg/push/driver/httpserver/protocol"
	"io"
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
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"test",
				"msg/foo/bar",
				&protocol.GatewayMessageCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
login
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"test",
				"login",
				&protocol.GatewayLoginCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
close
{"id":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"test",
				"close",
				&protocol.GatewayCloseCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
p2u
{"appid":"","userid":"","timestamp":0,"token":""}
foo bar`),
			protocol.Command{
				"t1",
				"test",
				"p2u",
				&pushprotocol.Push2UserCommand{},
				[]byte("foo bar"),
			},
		},
		{
			[]byte(`t1
test
hb

foo bar`),
			protocol.Command{
				"t1",
				"test",
				"hb",
				nil,
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

		cmd, err = ParseReader(bytes.NewBuffer(testCase.Message))
		if err != nil {
			t.Errorf("TestPlainText Case[%d]\nParse %s error: %s",
				index, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestPlainText Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index, testCase.Message, cmd, testCase.ExpectCommand)
		}

		buf := Compose(cmd)
		if bytes.Compare(buf, testCase.Message) != 0 {
			t.Errorf("TestPlainText Case[%d]\nCompose %s\nGot: %s,\nExpect: %s",
				index, cmd, buf, testCase.Message)
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
					index+1, testCase.Message, ErrJSONError, testCase.Error)
			}
		} else if err != testCase.Error {
			t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
				index+1, testCase.Message, reflect.TypeOf(err), testCase.Error)
		}
	}
}

type TestReader struct{}

func (r *TestReader) Read(p []byte) (int, error) {
	return 0, errors.New("Error")
}

func TestParseReaderError(t *testing.T) {
	var (
		r   io.Reader
		err error
	)
	_, err = ParseReader(r)
	if err == nil {
		t.Error("ParseReader(nil) should return error")
	}
	_, err = ParseReader(new(TestReader))
	if err == nil {
		t.Error("ParseReader(TestReader) should return error")
	}
}
