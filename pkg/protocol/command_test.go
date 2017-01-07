// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import (
	"bytes"
	"fmt"
	"testing"
)

type CheckFunc func(cmd1 *Command, cmd2 *Command) bool

type Check struct {
	Name string
	Func CheckFunc
}

func CommandEqual(cmd1 *Command, cmd2 *Command) bool {
	return cmd1.Equal(cmd2)
}

func CommandUnequal(cmd1 *Command, cmd2 *Command) bool {
	return !cmd1.Equal(cmd2)
}

type TestEqualCase struct {
	Cmd1  *Command
	Cmd2  *Command
	Check Check
}

func (testCase TestEqualCase) String() string {
	return fmt.Sprintf("Expect %s of:%s,%s\n",
		testCase.Check.Name, testCase.Cmd1, testCase.Cmd2)
}

var (
	Equal   = Check{"equal", CommandEqual}
	Unequal = Check{"unequal", CommandUnequal}
)

type JSONError struct {
	Data interface{} `json:"data"`
}

func TestEqual(t *testing.T) {
	testCases := []TestEqualCase{
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			Equal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				nil,
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				nil,
				[]byte("foo bar"),
			},
			Equal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			&Command{
				"t2",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayCloseCommand{},
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				nil,
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&GatewayMessageCommand{},
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				make(chan string),
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				make(chan string),
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&JSONError{make(chan string)},
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&JSONError{&GatewayMessageCommand{}},
				[]byte("foo bar"),
			},
			Unequal,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&JSONError{&GatewayMessageCommand{}},
				[]byte("foo bar"),
			},
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				&JSONError{make(chan string)},
				[]byte("foo bar"),
			},
			Unequal,
		},
	}
	for index, testCase := range testCases {
		if !testCase.Check.Func(testCase.Cmd1, testCase.Cmd2) {
			t.Errorf("\nTestCommand Case[%d] failed\n%s\n",
				index+1, testCase)
		}
	}
}

type TestFirstPartNameCase struct {
	Name          string
	FirstPartName string
}

func TestFirstPartName(t *testing.T) {
	testCases := []TestFirstPartNameCase{
		{"msg/foo/bar", "msg"},
		{"msg", "msg"},
		{"", ""},
		{`123\msg/foo/bar`, `123\msg`},
	}
	cmd := new(Command)
	for index, testCase := range testCases {
		cmd.Name = testCase.Name
		if cmd.FirstPartName() != testCase.FirstPartName {
			t.Errorf("\nTestFirstPartName Case[%d] failed:\nexpect: %s\n   got: %s",
				index+1, testCase.FirstPartName, cmd.FirstPartName())
		}
	}
}

type TestStringCase struct {
	Cmd    *Command
	Expect string
}

func TestString(t *testing.T) {
	testCases := []TestStringCase{
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				nil,
				[]byte("foo bar"),
			},
			`
{
  Version: t1
  AppID: test
  Name: msg/foo/bar
  Data: nil
  Payload: [102 111 111 32 98 97 114]
}
`,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				make(chan string),
				[]byte("foo bar"),
			},
			`
{
  Version: t1
  AppID: test
  Name: msg/foo/bar
  Data: ERROR
  Payload: [102 111 111 32 98 97 114]
}
`,
		},
		{
			&Command{
				"t1",
				"test",
				"msg/foo/bar",
				"data",
				[]byte("foo bar"),
			},
			`
{
  Version: t1
  AppID: test
  Name: msg/foo/bar
  Data: "data"
  Payload: [102 111 111 32 98 97 114]
}
`,
		},
	}
	for index, testCase := range testCases {
		if testCase.Cmd.String() != testCase.Expect {
			t.Errorf("\nTestString Case[%d] failed:\nexpect: %s\n   got: %s",
				index+1, testCase.Expect, testCase.Cmd.String())
		}
	}
}

func TestCopy(t *testing.T) {
	cmd := Command{
		"t1",
		"test",
		"msg/foo/bar",
		"data",
		[]byte("foo bar"),
	}
	cpCmd := cmd.Copy()
	if cmd.Version != cpCmd.Version ||
		cmd.Name != cpCmd.Name ||
		cmd.AppID != cpCmd.AppID ||
		cmd.Data != cpCmd.Data ||
		bytes.Compare(cmd.Payload, cpCmd.Payload) != 0 {
		t.Errorf("Copy test failed")
	}
}
