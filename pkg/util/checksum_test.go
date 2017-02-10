// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"flag"
	"testing"
	"crypto/md5"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type CheckSumCase struct {
	fields [][]byte
	expectMD5 string
	expectSHA1 string
	expectSHA256 string
}

func TestCheckSum(t *testing.T) {
	testcases := []CheckSumCase {
		CheckSumCase{
			fields: [][]byte{[]byte("12345"), []byte("67890")},
			expectMD5: "E807F1FCF82D132F9BB018CA6738A19F",
			expectSHA1: "01B307ACBA4F54F55AAFC33BB06BBBF6CA803E9A",
			expectSHA256: "C775E7B757EDE630CD0AA1113BD102661AB38829CA52A6422AB782862F268646",
		},
		CheckSumCase{
			fields: nil,
			expectMD5: "D41D8CD98F00B204E9800998ECF8427E",
			expectSHA1: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709",
			expectSHA256: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
		},
	}

	var got string
	for index, testcase := range testcases {
		got = CheckSumMD5(testcase.fields...)
		if got != testcase.expectMD5 {
			t.Errorf("Case(%d): CheckSumMD5 got: %s, expect: %s\n", index + 1, got, testcase.expectMD5)
		}
		got = CheckSumSHA1(testcase.fields...)
		if got != testcase.expectSHA1 {
			t.Errorf("Case(%d): CheckSumSHA1 got: %s, expect: %s\n", index + 1, got, testcase.expectMD5)
		}
		got = CheckSumSHA256(testcase.fields...)
		if got != testcase.expectSHA256 {
			t.Errorf("Case(%d): CheckSumSHA256 got: %s, expect: %s\n", index + 1, got, testcase.expectMD5)
		}
	}
}

func TestCheckSumWithKey (t *testing.T) {
	h := md5.New()
	got := CheckSumWithKey(h, []byte("1234567890"), []byte("foo"), []byte("bar"))
	expect := "36ECD86763CB0AD667A1B6750548AC58"
	if got != expect {
		t.Errorf("TestCheckSumWithKey got: %s, expect: %s\n", got, expect)
	}
}

func TestNewNonce(t *testing.T) {
	nonce, err := NewNonce()
	if err != nil {
		t.Errorf("NewNonce error: %s\n", err)
	}
	if len(nonce) == 0 {
		t.Errorf("NewNonce no nonce\n")
	}
}