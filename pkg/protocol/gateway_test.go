// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import "testing"

func TestKey_Token(t *testing.T) {
	key := Key("1234567890")
	cmd := GatewayLoginCommand{"123", "web", 1234567, "67D0D82549FA531D1E3A70371B99BB76"}

	token := key.Token(&cmd)
	if token != cmd.Token {
		t.Errorf("TestKey_Token expect: %s, got: %s\n", cmd.Token, token)
	}
}
