// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package serialize

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
	testSerializer := Serializer{
		Version:   "!1",
		ProbeByte: '!',
		NewParseEngine: func() (engine ParseEngine) {
			return &TestEngine{
				Buffer: new(bytes.Buffer),
			}
		},
		Compose: func(cmd *protocol.Command) ([]byte, error) {
			buf := new(bytes.Buffer)
			buf.WriteByte('!')
			buf.Write(cmd.Payload)
			buf.WriteByte('$')
			return buf.Bytes(), nil
		},
	}
	Register(&testSerializer)
}

type TestEngine struct {
	Buffer *bytes.Buffer
}

func (e *TestEngine) Parse(br *bufio.Reader) (cmd *protocol.Command, err error) {
	var (
		str string
	)
	_, err = io.Copy(e.Buffer, br)
	if err != nil && err != io.EOF {
		return
	}

	if str, err = e.Buffer.ReadString('$'); err != nil {
		return
	}

	cmd = &protocol.Command{
		Version: "!1",
		AppID:   "test",
		Name:    "msg/foo/bar",
		Data:    &protocol.GatewayMessageCommand{},
		Payload: []byte(str)[1 : len(str)-1],
	}
	return
}

func (e *TestEngine) Close() error {
	e.Buffer.Reset()
	return nil
}

func TestParser(t *testing.T) {
	var (
		err          error
		cmd          *protocol.Command
		composeBytes []byte
	)

	buf := bytes.NewBufferString("!12345$")

	p := NewParser(buf)
	cmd, err = p.ReadCommand()
	assert.NoError(t, err, "Should return io.EOF")
	if !bytes.Equal([]byte(`12345`), cmd.Payload) {
		t.Error("Should read all bytes")
	}
	composeBytes, err = Compose(cmd)
	if err != nil {
		t.Error("Compose error:", err)
	} else {
		if !bytes.Equal(composeBytes, []byte(`!12345$`)) {
			t.Errorf("Compose result : %s, expect: %s", string(composeBytes), `!12345$`)
		}
	}
	buf.WriteString("!6789$!abc")
	cmd, err = p.ReadCommand()
	assert.NoError(t, err, "Should return io.EOF")
	if !bytes.Equal([]byte(`6789`), cmd.Payload) {
		t.Error("Should read all bytes")
	}
	buf.WriteString("d$")
	cmd, err = p.ReadCommand()
	assert.NoError(t, err, "Should return io.EOF")
	if !bytes.Equal([]byte(`abcd`), cmd.Payload) {
		t.Error("Should read all bytes")
	}
}

type TestReader struct{}

var ErrTestReader = errors.New("raise error for test")

func (r *TestReader) Read(p []byte) (int, error) {
	return 0, ErrTestReader
}

func TestError(t *testing.T) {
	var (
		err error
	)

	p := NewParser(new(TestReader))

	_, err = p.ReadCommand()
	if err != ErrTestReader {
		t.Error("ParseReader(TestReader) should return error: ", ErrTestReader)
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
