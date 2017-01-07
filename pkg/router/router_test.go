// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package router

import (
	"strings"
	"testing"
)

func TestRouter(t *testing.T) {
	routerMap := InfoMap{
		"*":      {"httpapi", "http://test/im"},
		"login":  {"httpapi", "http://test/im"},
		"msg":    {"httpapi", "http://test/im"},
		"logout": {"httpapi", "http://test/im"},
	}
	r, err := NewRouter(routerMap)
	if err != nil {
		t.Fatalf("NewRouter error: %s\n", err)
	}
	inv := r.Find("msg")
	if inv == nil {
		t.Fatalf("Find(%s) return nil", "msg")
	}

	inv = r.Find("xxx")
	if inv == nil {
		t.Fatalf("Find(%s) return nil", "msg")
	}

	strs := strings.Split(strings.Trim(r.String(), "\n"), "\n")

	if len(strs) != len(routerMap) {
		t.Errorf("Router string Got:\n%s\n", r.String())
	}
}

func TestErrorRouter(t *testing.T) {
	routerMap := InfoMap{
		"login": {"xxx", "http://test/im"},
	}
	_, err := NewRouter(routerMap)
	if err == nil {
		t.Errorf("TestErrorRouter should return error\n")
	}
}
