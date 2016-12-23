// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"flag"
	"fmt"
	"testing"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

func TestRecoverFromPanic(t *testing.T) {
	panicFunc := func() {
		defer RecoverFromPanic()
		var list []int
		t.Log("nil point")
		fmt.Printf("%d\n", list[100])
	}
	panicFunc()
}
