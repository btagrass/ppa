package job

import (
	s "github.com/btagrass/gobiz/svc"
	"github.com/btagrass/gobiz/sys/svc"
	"github.com/sirupsen/logrus"
)

func init() {
	err := s.Use[*svc.JobSvc]().AddJobs(&PhotoHashJob{})
	if err != nil {
		logrus.Error(err)
	}
}
