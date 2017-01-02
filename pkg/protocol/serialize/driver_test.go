// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"testing"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestParseCase struct {
	Message       []byte
	ExpectCommand protocol.Command
}

func TestParse(t *testing.T) {
	testCases := []TestParseCase{
		{
			[]byte(`t1
test
msg/foo/bar
{"userid":""}
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
			[]byte(`{"version":"j1","appid":"test","name":"msg/foo/bar","data":{"userid":""},"payload":"` +
				base64.StdEncoding.EncodeToString([]byte("foo bar")) + `"}`),
			protocol.Command{
				"j1",
				"test",
				"msg/foo/bar",
				&protocol.GatewayMessageCommand{},
				[]byte("foo bar"),
			},
		},
	}
	for index, testCase := range testCases {
		cmd, err := Parse(testCase.Message)
		if err != nil {
			t.Errorf("TestParse Case[%d]\nParse %s error: %s",
				index, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestParse Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index, testCase.Message, cmd, testCase.ExpectCommand)
			continue
		}

		cmd, err = ParseReader(bytes.NewBuffer(testCase.Message))
		if err != nil {
			t.Errorf("TestParse Case[%d]\nParse %s error: %s",
				index, testCase.Message, err)
			continue
		}
		if !testCase.ExpectCommand.Equal(cmd) {
			t.Errorf("TestParse Case[%d]\nParse %s\nGot: %s,\nExpect: %s",
				index, testCase.Message, cmd, testCase.ExpectCommand)
			continue
		}

		buf, err := Compose(cmd)
		if err != nil {
			t.Errorf("TestParse Case[%d] compose error: %s\n", index, err)
			continue
		}
		if bytes.Compare(buf, testCase.Message) != 0 {
			t.Errorf("TestParse Case[%d]\nCompose %s\nGot: %s,\nExpect: %s",
				index, cmd, buf, testCase.Message)
		}
	}
}

type TestReader struct{}

var ErrTestReader = errors.New("raise error for test")

func (r *TestReader) Read(p []byte) (int, error) {
	return 0, ErrTestReader
}

func TestError(t *testing.T) {
	var err error
	// ParseReader
	_, err = ParseReader(nil)
	if err != define.ErrInvalidParameter {
		t.Error("ParseReader(nil) should return error: ", define.ErrInvalidParameter)
	}

	_, err = ParseReader(new(TestReader))
	if err != ErrTestReader {
		t.Error("ParseReader(TestReader) should return error: ", ErrTestReader)
	}

	// Parse
	_, err = Parse(nil)
	if err != define.ErrInvalidParameter {
		t.Error("Parse(nil) should return error: ", define.ErrInvalidParameter)
	}
	_, err = Parse([]byte{})
	if err != define.ErrInvalidParameter {
		t.Error("Parse([]byte{}) should return error: ", define.ErrInvalidParameter)
	}
	_, err = Parse([]byte{'x'})
	if err != define.ErrUnsupportProtocol {
		t.Error("Parse([]byte{'x'}) should return error: ", define.ErrUnsupportProtocol)
	}

	// Compose
	_, err = Compose(nil)
	if err != define.ErrInvalidParameter {
		t.Error("Compose(nil) should return error: ", define.ErrInvalidParameter)
	}
	_, err = Compose(&protocol.Command{Version: "xx"})
	if err != define.ErrUnsupportProtocol {
		t.Error("Compose(&protocol.Command{Version:xx}) should return error: ", define.ErrUnsupportProtocol)
	}
}
