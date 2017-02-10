// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package util

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
)

const (
	// NonceBytes nonce字节长度
	NonceBytes = 32
)

// CheckSumWithKey 使用秘钥计算hash
func CheckSumWithKey(h hash.Hash, key []byte, fields ...[]byte) string {
	h.Write([]byte(key))
	for _, field := range fields {
		h.Write(field)
	}
	return fmt.Sprintf("%X", h.Sum(nil))
}

// CheckSum 计算hash
func CheckSum(h hash.Hash, fields ...[]byte) string {
	for _, field := range fields {
		h.Write(field)
	}
	return fmt.Sprintf("%X", h.Sum(nil))
}

// CheckSumSHA256 sha256哈希
func CheckSumSHA256(fields ...[]byte) string {
	h := sha256.New()
	return CheckSum(h, fields...)
}

// CheckSumSHA1 sha1哈希
func CheckSumSHA1(fields ...[]byte) string {
	h := sha1.New()
	return CheckSum(h, fields...)
}

// CheckSumMD5 md5哈希
func CheckSumMD5(fields ...[]byte) string {
	h := md5.New()
	return CheckSum(h, fields...)
}

// NewNonce 新建Nonce
func NewNonce() (nonce string, err error) {
	b := make([]byte, NonceBytes)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	nonce = base64.StdEncoding.EncodeToString(b)
	return
}
