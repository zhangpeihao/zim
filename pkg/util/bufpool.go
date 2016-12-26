// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"bytes"
	"sync"
)

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// GetBuffer 从缓存池中取出一个buffer
func GetBuffer() *bytes.Buffer {
	return bufferPool.Get().(*bytes.Buffer)
}

// PutBuffer 将buffer重新放入缓存池
func PutBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bufferPool.Put(buf)
}
