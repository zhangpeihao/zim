// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package jsonfile

import "testing"

func TestNewRouter(t *testing.T) {
	r, err := NewRouter("./router.json")
	if err != nil {
		t.Fatalf("NewRouter error: %s\n", err)
	}
	inv := r.Find("msg")
	if inv == nil {
		t.Fatalf("Find(%s) return nil", "msg")
	}
}
