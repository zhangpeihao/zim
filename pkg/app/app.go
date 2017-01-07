// Copyright 2016 Zhang Peihao <zhangpeihao@gmail.com>

package app

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/zhangpeihao/zim/pkg/router"
	"os"
)

// App 应用数据
type App struct {
	Name     string         `json:"name"`
	Key      string         `json:"key"`
	RouteMap router.InfoMap `json:"router"`
	Router   *router.Router `json:"-"`
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
	app.Router, err = router.NewRouter(app.RouteMap)
	if err != nil {
		glog.Errorf("define::NewApp(%s) NewRouter error: %s\n", config, err)
		return nil, err
	}
	return &app, nil
}

// TokenSHA1 取得Token SHA1算法
func (app *App) TokenSHA1(fields ...string) string {
	h := sha1.New()
	for _, field := range fields {
		h.Write([]byte(field))
	}
	h.Write([]byte(app.Key))
	return fmt.Sprintf("%X", h.Sum(nil))
}

// TokenMD5 取得Token MD5算法
func (app *App) TokenMD5(fields ...string) string {
	h := md5.New()
	for _, field := range fields {
		h.Write([]byte(field))
	}
	h.Write([]byte(app.Key))
	return fmt.Sprintf("%X", h.Sum(nil))
}
