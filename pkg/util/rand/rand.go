// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package rand

import (
	"math/rand"
	"sync"
	"time"
)

type SafeRand struct {
	sync.Mutex
	rand *rand.Rand
}

var staticRand = SafeRand{
	rand: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
}

// Intn 创建一个小于max的随机整数
func Intn(max int) int {
	staticRand.Lock()
	defer staticRand.Unlock()
	return staticRand.rand.Intn(max)
}

// IntnRange 创建一个范围在[min,max)之间的随机整数
func IntnRange(min, max int) int {
	staticRand.Lock()
	defer staticRand.Unlock()
	return staticRand.rand.Intn(max - min) + min
}

// Int63nRange 创建一个范围在[min,max)之间的随机int64整数
func Int63nRange(min, max int64) int64 {
	staticRand.Lock()
	defer staticRand.Unlock()
	return staticRand.rand.Int63n(max - min) + min
}

// Seed 设置随机种子
func Seed(seed int64) {
	staticRand.Lock()
	defer staticRand.Unlock()

	staticRand.rand = rand.New(rand.NewSource(seed))
}

// Perm 返回一个随机数Slice.
func Perm(n int) []int {
	staticRand.Lock()
	defer staticRand.Unlock()
	return staticRand.rand.Perm(n)
}
