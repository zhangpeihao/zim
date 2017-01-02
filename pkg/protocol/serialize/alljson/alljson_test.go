// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package alljson

import (
	"bytes"
	"encoding/json"
	"github.com/zhangpeihao/zim/pkg/protocol"
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
