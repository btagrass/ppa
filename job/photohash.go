package job

import (
	"path/filepath"
	"ppa/mdl"
	"ppa/svc"
	"time"

	s "github.com/btagrass/gobiz/svc"
	"github.com/btagrass/gobiz/utl"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type PhotoHashJob struct {
	Arg string
}

func (j *PhotoHashJob) Run() {
	defer utl.ElapsedTime("PhotoHashJob", time.Now())
	photoSvc := s.Use[*svc.PhotoSvc]()
	dataPhotos, _, err := photoSvc.List()
	if err != nil {
		logrus.Error(err)
		return
	}
	searchPhotos, err := photoSvc.ListSearchPhotos()
	if err != nil {
		logrus.Error(err)
		return
	}
	d1, d2 := utl.Difference(dataPhotos, searchPhotos, func(t1, t2 mdl.Photo) bool {
		return t1.UID == t2.UID
	})
	uids := lo.Map(d1, func(item mdl.Photo, index int) string {
		return item.UID
	})
	err = photoSvc.Remove("uid in ?", uids)
	if err != nil {
		logrus.Error(err)
	}
	utl.ForParallel(d2, func(t mdl.Photo) error {
		hash, err := photoSvc.CalcImageHash(filepath.Join(viper.GetString("app.dir"), t.FileName))
		if err != nil {
			return err
		}
		t.Hash = cast.ToString(hash)
		return photoSvc.Save(t)
	}, 50)
}

func (j *PhotoHashJob) GetName() string {
	return "PhotoHash"
}
func (j *PhotoHashJob) GetDesc() string {
	return "图片哈希"
}
func (j *PhotoHashJob) GetCron() string {
	return "@every 1h"
}
func (j *PhotoHashJob) GetArg() string {
	return ""
}
func (j *PhotoHashJob) GetArgDesc() string {
	return ""
}

func (j *PhotoHashJob) SetArg(arg string) {
	j.Arg = arg
}
