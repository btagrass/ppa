package svc

import (
	"fmt"
	"ppa/mdl"

	"github.com/btagrass/gobiz/htp"
	"github.com/btagrass/gobiz/svc"
	"github.com/samber/do"
)

// 应用服务
type AppSvc struct {
	*svc.DataSvc[mdl.App]
}

// 构造函数
func NewAppSvc(i *do.Injector) (*AppSvc, error) {
	return &AppSvc{
		DataSvc: svc.NewDataSvc[mdl.App]("ppa:apps"),
	}, nil
}

func (s *AppSvc) GetApp() *mdl.App {
	app, _ := s.Get()
	if app == nil {
		app = &mdl.App{
			Lang: "English",
		}
	}
	return app
}

func (s *AppSvc) LoginApp() (*mdl.App, error) {
	k := s.GetFullKey("loginApp")
	v, ok := s.Local.Get(k)
	if ok {
		return v.(*mdl.App), nil
	}
	var r struct {
		AccessToken string `json:"access_token"`
	}
	app := s.GetApp()
	_, err := htp.Post(fmt.Sprintf("%s/api/v1/session", app.Url), nil, map[string]any{
		"username": app.UserName,
		"password": app.Password,
	}, &r)
	if err != nil {
		return nil, err
	}
	app.Token = r.AccessToken
	s.Local.SetDefault(k, app)
	return app, nil
}
