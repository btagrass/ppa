package mdl

import (
	"github.com/btagrass/gobiz/mdl"
)

type App struct {
	mdl.Mdl
	Lang     string
	Url      string
	UserName string
	Password string
	Token    string
}
