// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroadcastSignal(t *testing.T) {
	bc := NewBroadcastSignal()

	var counter int32
	for i := 0; i < 10; i++ {
		go func() {
			atomic.AddInt32(&counter, 1)
			wait := bc.Wait()
			<-wait
			atomic.AddInt32(&counter, -1)
			bc.RemoveWait(wait)
		}()
	}

	time.Sleep(time.Second)
	var v int32
	v = atomic.LoadInt32(&counter)
	assert.Equal(t, int32(10), v)
	bc.Broadcast()
	time.Sleep(time.Second * 2)
	v = atomic.LoadInt32(&counter)
	assert.Equal(t, int32(0), v)
	bc.Close()
}
