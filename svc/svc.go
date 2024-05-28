package svc

import (
	s "github.com/btagrass/gobiz/svc"
	"github.com/btagrass/gobiz/sys/svc"
)

func init() {
	s.Inject(svc.NewJobSvc)
	s.Inject(NewAppSvc)
	s.Inject(NewPhotoSvc)
}
