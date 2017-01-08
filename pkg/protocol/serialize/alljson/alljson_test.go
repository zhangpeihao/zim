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
	"io"
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

func InterfaceCompare(this, that interface{}) int {
	thisStr, err := json.Marshal(this)
	if err != nil {
		return -1
	}
	thatStr, err := json.Marshal(that)
	if err != nil {
		return -1
	}
	if bytes.Equal(thisStr, thatStr) {
		return 0
	}
	return 1
}

func TestCopy(t *testing.T) {
	testCases := []TestCopyCase{
		{
			`{"data":{"userid":"123","timestamp":123456,"token":"54321"}}`,
			&protocol.GatewayLoginCommand{},
			&protocol.GatewayLoginCommand{
				UserID:    "123",
				Timestamp: 123456,
				Token:     "54321",
			},
		},
		{
			`{"data":{"userid":"","timestamp":0,"token":""}}`,
			&protocol.GatewayLoginCommand{},
			&protocol.GatewayLoginCommand{
				UserID:    "",
				Timestamp: 0,
				Token:     "",
			},
		},
		{
			`{"data":{}}`,
			&protocol.GatewayLoginCommand{},
			&protocol.GatewayLoginCommand{
				UserID:    "",
				Timestamp: 0,
				Token:     "",
			},
		},
	}

	for index, testCase := range testCases {
		var src JSONData
		err := json.Unmarshal([]byte(testCase.SrcJSON), &src)
		if err != nil {
			t.Errorf("Test case[%d] json unmarshal error: %s\n", index, err)
			continue
		}
		err = Copy(src.Data, testCase.Dest)
		if err != nil {
			t.Errorf("Test case[%d] reflectIt error: %s\n", index, err)
			continue
		}
		if InterfaceCompare(testCase.Dest, testCase.Expect) != 0 {
			t.Errorf("Test case[%d] InterfaceCompare return 0 testCase.Dest: %+v, testCase.Expect: %+v\n",
				index, testCase.Dest, testCase.Expect)
			continue
		}
	}
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
			[]byte(fmt.Sprintf(`{"version":"j1","appid":"test","name":"p2u","data":{"useridlist":""},"payload":"%s"}`, base64Payload)),
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
				index + 1, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestAllJson Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index + 1, testCase.Message, cmd, testCase.ExpectCommand)
		}

		cmd, err = ParseReader(bytes.NewBuffer(testCase.Message))
		if err != nil {
			t.Errorf("TestAllJson Case[%d]\nParse %s error: %s",
				index + 1, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestAllJson Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index + 1, testCase.Message, cmd, testCase.ExpectCommand)
		}

		buf, err := Compose(cmd)
		if err != nil {
			t.Errorf("TestAllJson Case[%d] error: %s\n", index + 1, err)
		}
		if bytes.Compare(buf, testCase.Message) != 0 {
			t.Errorf("TestAllJson Case[%d]\nCompose %s\nGot: %s,\nExpect: %s",
				index + 1, cmd, buf, testCase.Message)
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
				index + 1, testCase.Message, testCase.Error)
		} else if reflect.TypeOf(err) == reflect.TypeOf(new(json.SyntaxError)) {
			if testCase.Error != ErrJSONError {
				t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
					index + 1, testCase.Message, err, testCase.Error)
			}
		} else if err != testCase.Error {
			t.Errorf("TestError Case[%d]\nParse %s\nGot: %+v,\nExpect: %s",
				index + 1, testCase.Message, reflect.TypeOf(err), testCase.Error)
		}
	}
}

type TestReader struct{}

func (r *TestReader) Read(p []byte) (int, error) {
	return 0, errors.New("Error")
}

func TestParseReaderError(t *testing.T) {
	var (
		r io.Reader
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
