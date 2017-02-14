// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package app

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"os"

	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/util"
)

// App 应用数据
type App struct {
	Name       string  `json:"name"`
	Key        string  `json:"key"`
	KeyBytes   []byte  `json:"-"`
	RouteMap   InfoMap `json:"router"`
	Router     *Router `json:"-"`
	TokenCheck string  `json:"token-check"`
}

// CheckSum CheckSum接口
type CheckSum interface {
	CheckSumSHA1(fields ...[]byte) string
	CheckSumSHA256(fields ...[]byte) string
	CheckSumMD5(fields ...[]byte) string
}

// Service App服务
type Service interface {
	// GetCheckSum 根据应用标志获取CheckSum接口
	GetCheckSum(name string) CheckSum
}

// NewApp 新建App
func NewApp(config string) (*App, error) {
	glog.Infof("define::NewApp(%s)\n", config)
	f, err := os.Open(config)
	if err != nil {
		glog.Errorf("define::NewApp(%s) os.Open() error: %s\n", config, err)
		return nil, err
	}
	dec := json.NewDecoder(f)
	var app App
	err = dec.Decode(&app)
	if err != nil {
		glog.Errorf("define::NewApp(%s) json.Decode error: %s\n", config, err)
		return nil, err
	}
	app.Router, err = NewRouter(app.RouteMap)
	if err != nil {
		glog.Errorf("define::NewApp(%s) NewRouter error: %s\n", config, err)
		return nil, err
	}
	app.KeyBytes = []byte(app.Key)
	return &app, nil
}

// CheckSumSHA1 取得CheckSum SHA1算法
func (app *App) CheckSumSHA1(fields ...[]byte) string {
	h := sha1.New()
	return util.CheckSumWithKey(h, app.KeyBytes, fields...)
}

// CheckSumSHA256 取得CheckSum SHA1算法
func (app *App) CheckSumSHA256(fields ...[]byte) string {
	h := sha256.New()
	return util.CheckSumWithKey(h, app.KeyBytes, fields...)
}

// CheckSumMD5 取得CheckSum MD5算法
func (app *App) CheckSumMD5(fields ...[]byte) string {
	h := md5.New()
	return util.CheckSumWithKey(h, app.KeyBytes, fields...)
}
