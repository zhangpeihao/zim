// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"flag"
	"github.com/jarcoal/httpmock"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/protocol/driver/plaintext"
	"testing"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestInvokerCase struct {
	StubMethod       string
	StubStatus       int
	URL              string
	Name             string
	Payload          []byte
	StubResponseData string
	InvokeError      bool
}

func init() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	Transport = httpmock.DefaultTransport
}

func TestInvoker(t *testing.T) {
	testCases := []TestInvokerCase{
		{"POST", 200, "http://test.stub/httpapi/testinvoker", "msg/foo/bar", []byte("foo"), `t1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
bar`, false},
		{"POST", 200, "http://test.stub/httpapi/testinvoker", "msg/foo/bar", []byte("foo"), `T1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
bar`, true},
		{"POST", 200, "http://test.stub/httpapi/testinvoker", "msg/foo/bar", []byte("foo"), ``, false},
		{"POST", 400, "http://test.stub/httpapi/testinvoker", "msg/foo/bar", []byte("foo"), `t1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
bar`, true},
		{"GET", 200, "xxxx://test.stub/httpapi/testinvoker", "msg/foo/bar", []byte("foo"), `t1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
bar`, true},
		{"POST", 200, "%gh&%ij", "msg/foo/bar", []byte("foo"), `t1
test
msg/foo/bar
{"id":"","timestamp":0,"token":""}
bar`, true},
	}

	for index, testCase := range testCases {
		httpmock.RegisterResponder(testCase.StubMethod,
			testCase.URL+"/"+testCase.Name,
			httpmock.NewStringResponder(testCase.StubStatus, testCase.StubResponseData))

		invoker := NewInvoker(testCase.URL)
		cmd := &protocol.Command{
			Version: "t1",
			Name:    testCase.Name,
			Payload: testCase.Payload,
		}
		resp, err := invoker.Invoke("", cmd)
		if !testCase.InvokeError {
			if err != nil {
				t.Errorf("TestError Case[%d]\nTestInvoker error: %s\n", index+1, err)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("TestError Case[%d]\nTestInvoker should error\n",
					index+1)
			}
			continue
		}
		if len(testCase.StubResponseData) > 0 {
			expectResp, err := plaintext.Parse([]byte(testCase.StubResponseData))
			if err != nil {
				t.Fatalf("TestError Case[%d]\nparse expect response error: %s\n", index+1, err)
			}
			if !resp.Equal(expectResp) {
				t.Errorf("TestError Case[%d]\nExpect reponse :%sGot: %s\n", index+1, expectResp, resp)
			}
		}
	}

}
