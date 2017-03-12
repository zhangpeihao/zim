// Copyright 2017 Zhang Peihao <zhangpeihao@gmail.com>

package app

import "context"

const (
	// ContextAppController AppController名
	ContextAppController contextKey = "zim/appcontroller"
)

type contextKey string

// Controller App map controller
type Controller struct {
	apps map[string]*App
}

// NewController create a new controller from configs
func NewController(configs []string) (*Controller, error) {
	controller := &Controller{
		apps: make(map[string]*App),
	}
	if configs != nil {
		for _, config := range configs {
			appConfig, err := NewApp(config)
			if err != nil {
				return nil, err
			}
			controller.apps[appConfig.ID] = appConfig
		}
	}
	return controller, nil
}

// GetApp find app by app ID
func (controller *Controller) GetApp(appid string) *App {
	if app, ok := controller.apps[appid]; ok {
		return app
	}
	return nil
}

// AddApp append an app into map (multi-thread unsafe!)
func (controller *Controller) AddApp(app *App) {
	controller.apps[app.ID] = app
}

// SaveIntoContext 设置AppController到Context中
func (controller *Controller) SaveIntoContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ContextAppController, controller)
}

// GetAppFromContext 从Context获取App
func GetAppFromContext(ctx context.Context, appid string) *App {
	value := ctx.Value(ContextAppController)
	if value == nil {
		return nil
	}
	appController, ok := value.(*Controller)
	if !ok {
		return nil
	}
	return appController.GetApp(appid)
}
