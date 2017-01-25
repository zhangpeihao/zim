// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"bytes"
	"errors"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"testing"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestParse(t *testing.T) {
	var err error
	testSerializer := Serializer{
		Version:   "test",
		ProbeByte: '!',
		Parse: func(message []byte) (cmd *protocol.Command, err error) {
			return nil, nil
		},
		Compose: func(cmd *protocol.Command) ([]byte, error) {
			return nil, nil
		},
	}
	Register(&testSerializer)
	_, err = Parse([]byte("!12345"))
	assert.NoError(t, err)
	_, err = ParseReader(bytes.NewBufferString("!12345"))
	assert.NoError(t, err)
	_, err = Compose(&protocol.Command{Version: "test"})
	assert.NoError(t, err)
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
