// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	buff := GetBuffer()
	buff.WriteString("do be do be do")
	assert.Equal(t, "do be do be do", buff.String())
	PutBuffer(buff)
	assert.Equal(t, 0, buff.Len())
}
