// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/context"
	"github.com/zhangpeihao/zim/pkg/define"
	"github.com/zhangpeihao/zim/pkg/protocol"
	"github.com/zhangpeihao/zim/pkg/util"
)

// ParseCommand 解析出一个命令对象
func ParseCommand(ctx context.Context, tag string, header http.Header, payload []byte,
	timeout int) (cmd *protocol.Command, err error) {
	// 从Header中取出参数
	cmd = &protocol.Command{
		Payload: payload,
	}
	var (
		payloadMD5, nonce, timestamp, checksum, data string
	)
	if cmd.AppID = header.Get(HeaderAppID); len(cmd.AppID) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderAppID)
		return nil, define.ErrInvalidParameter
	}
	if cmd.Name = header.Get(HeaderName); len(cmd.Name) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderName)
		return nil, define.ErrInvalidParameter
	}
	if data = header.Get(HeaderData); len(data) == 0 {
		glog.Infoln("broker::httpapi::ParseCommand() no data")
	}
	if payloadMD5 = header.Get(HeaderPayloadMD5); len(payloadMD5) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderPayloadMD5)
		return nil, define.ErrInvalidParameter
	}
	if nonce = header.Get(HeaderNonce); len(nonce) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderNonce)
		return nil, define.ErrInvalidParameter
	}
	if timestamp = header.Get(HeaderTimestamp); len(timestamp) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderTimestamp)
		return nil, define.ErrInvalidParameter
	}
	if checksum = header.Get(HeaderCheckSum); len(checksum) == 0 {
		glog.Warningln("broker::httpapi::ParseCommand() miss header ", HeaderCheckSum)
		return nil, define.ErrInvalidParameter
	}

	var ts int
	if ts, err = strconv.Atoi(timestamp); err != nil {
		glog.Warningln("broker::httpapi::ParseCommand() parse timstamp(", timestamp, ") error:", err)
		return nil, define.ErrInvalidParameter
	}
	now := int(time.Now().Unix())
	if ts+timeout < now {
		glog.Warningf("broker::httpapi::ParseCommand() request timeout!\ntimstamp:%s, timeout:%d, now:%d\n",
			timestamp, timeout, now)
		return nil, define.ErrInvalidParameter
	}

	app := ctx.GetCheckSum(cmd.AppID)
	if app == nil {
		glog.Warningln("broker::httpapi::ParseCommand() no app(", cmd.AppID, ")")
		return nil, define.ErrInvalidParameter
	}
	checksumExpect := app.CheckSumSHA256([]byte(tag), []byte(cmd.AppID), []byte(cmd.Name),
		[]byte(data), []byte(payloadMD5), []byte(nonce), []byte(timestamp))
	if checksumExpect != checksum {
		glog.Warningf("broker::httpapi::ParseCommand() checksum error!\ngot: %s\nexpect: %s\n",
			checksum, checksumExpect)
		return nil, define.ErrInvalidParameter
	}

	expectPayloadMD5 := app.CheckSumMD5(cmd.Payload)
	gotPayloadMD5 := strings.ToUpper(payloadMD5)
	if gotPayloadMD5 != expectPayloadMD5 {
		glog.Warningf("broker::httpapi::ParseCommand() checksum unmatch!\ngot: %s\nexpect: %s\n",
			gotPayloadMD5, expectPayloadMD5)
		return nil, define.ErrInvalidParameter
	}

	if len(data) > 0 {
		if err = cmd.ParseData([]byte(data)); err != nil {
			glog.Warningf("broker::httpapi::ParseCommand() ParseData(%s) error %s\n",
				data, err)
			return nil, define.ErrInvalidParameter
		}
	}
	return
}

// ComposeCommand 解析出一个命令对象
func ComposeCommand(ctx context.Context, tag string, header http.Header,
	cmd *protocol.Command) (err error) {
	header.Set("User-Agent", HeaderAgent)

	app := ctx.GetCheckSum(cmd.AppID)
	if app == nil {
		glog.Warningln("broker::httpapi::Publish() no app(", cmd.AppID, ")")
		return fmt.Errorf("no AppID %s", cmd.AppID)
	}

	payloadMD5 := app.CheckSumMD5(cmd.Payload)
	header.Set(HeaderAppID, cmd.AppID)
	header.Set(HeaderName, cmd.Name)

	var data []byte
	if cmd.Data != nil {
		if data, err = json.Marshal(cmd.Data); err != nil {
			glog.Warningln("broker::httpapi::Publish() no app(", cmd.AppID, ")")
			return fmt.Errorf("marshal data error: %s", err)
		}
		header.Set(HeaderData, string(data))
	}

	var (
		nonce, timestamp, checksum string
	)
	if nonce, err = util.NewNonce(); err != nil {
		glog.Warningln("broker::httpapi::Publish() NewNonce error:", err)
		return err
	}

	timestamp = fmt.Sprintf("%d", time.Now().Unix())
	checksum = app.CheckSumSHA256([]byte(tag), []byte(cmd.AppID), []byte(cmd.Name),
		data, []byte(payloadMD5), []byte(nonce), []byte(timestamp))

	header.Set(HeaderPayloadMD5, payloadMD5)
	header.Set(HeaderNonce, nonce)
	header.Set(HeaderTimestamp, timestamp)
	header.Set(HeaderCheckSum, checksum)

	return nil
}
