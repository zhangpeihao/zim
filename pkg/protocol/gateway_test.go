// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package protocol

import "testing"

func TestKey_Token(t *testing.T) {
	key := Key("1234567890")
	cmd := GatewayCommonCommand{"123", 1234567, "E6B8D4E28E8DF1C331460DE60D9792FF"}

	token := key.Token(&cmd)
	if token != cmd.Token {
		t.Errorf("TestKey_Token expect: %s, got: %s\n", cmd.Token, token)
	}
}
