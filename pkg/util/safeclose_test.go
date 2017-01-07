// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"flag"
	"fmt"
	"testing"
	"time"
)

func init() {
	flag.Set("v", "4")
	flag.Set("logtostderr", "true")
}

type TestSignal struct{}

func (t *TestSignal) String() string {
	return "test"
}

func (t *TestSignal) Signal() {
}

func closeFunc() {
}

func RunLoop(t *testing.T, sc *SafeCloser, name string) {
	ch := make(chan struct{})
	err := sc.Add(name, func() {
		ch <- struct{}{}
	})
	if err != nil {
		t.Fatalf("SafeClose::Add() return error: %s", err)
	}
	for {
		select {
		case <-ch:
			t.Logf("goroutine %s closed", name)
			sc.Done(name)
			return
		case <-time.After(time.Second):
		}
	}
}

func BlockedLoop(t *testing.T, sc *SafeCloser, name string) {
	err := sc.Add(name, closeFunc)
	if err != nil {
		t.Fatalf("SafeClose::Add() return error: %s", err)
	}
	t.Logf("goroutine %s blocked", name)
	time.Sleep(time.Hour)
}

func TestSafeCloser(t *testing.T) {
	var err error
	sc := NewSafeCloser()

	for i := 0; i < 4; i++ {
		go RunLoop(t, sc, fmt.Sprintf("%d", i))
	}
	time.Sleep(time.Second * time.Duration(3))
	if err = sc.Add("0", closeFunc); err == nil {
		t.Error("SafeClose::Add() same name should return error")
	}

	if err = sc.Done("UnexistedName"); err == nil {
		t.Error("SafeClose::Done() name not existed should return error")
	}
	if sc.IsClose() {
		t.Error("SafeClose::IsClose() should return false before Close() function be invoked")
	}
	if err = sc.Close(time.Second * time.Duration(3)); err != nil {
		t.Errorf("SafeClose::Close return error: %s", err)
	}
	if !sc.IsClose() {
		t.Error("SafeClose::IsClose() should return true after Close() function be invoked")
	}
}

func TestSafeCloserInfiniteTimeout(t *testing.T) {
	var err error
	sc := NewSafeCloser()

	for i := 0; i < 4; i++ {
		go RunLoop(t, sc, fmt.Sprintf("%d", i))
	}

	go func() {
		time.Sleep(time.Second * time.Duration(3))
		sc.Lock()
		sc.terminationSignalsCh <- new(TestSignal)
		sc.Unlock()
	}()

	closed := false
	closeFunc := func() {
		closed = true
	}

	if err = sc.WaitAndClose(0, closeFunc); err != nil {
		t.Errorf("SafeClose::WaitAndClose return error: %s", err)
	}
	if !closed {
		t.Error("Close function not be invoked")
	}
}

func TestSafeCloserBlocked(t *testing.T) {
	var err error
	sc := NewSafeCloser()

	for i := 0; i < 4; i++ {
		go RunLoop(t, sc, fmt.Sprintf("%d", i))
	}
	go BlockedLoop(t, sc, "blocked")

	time.Sleep(time.Second * time.Duration(3))
	if err = sc.Close(time.Second * time.Duration(3)); err == nil {
		t.Error("SafeClose::Close when blocled should return error")
	}
}
