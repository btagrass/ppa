package gui

import (
	"embed"
	"fmt"
	"os"
	"ppa/svc"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/btagrass/gobiz/i18n"
	s "github.com/btagrass/gobiz/svc"
	"github.com/sirupsen/logrus"
)

//go:embed res/langs
var langs embed.FS

func NewMainGui() fyne.Window {
	ap2 := s.Use[*svc.AppSvc]().GetApp()
	file, err := langs.Open(fmt.Sprintf("res/langs/%s.yaml", strings.ToLower(ap2.Lang)))
	if err != nil {
		logrus.Error(err)
	}
	err = i18n.Load(file)
	if err != nil {
		logrus.Error(err)
	}
	hostname, _ := os.Hostname()
	app := app.NewWithID(hostname)
	app.Settings().SetTheme(&Theme{})
	win := app.NewWindow(i18n.T("m.Title"))
	win.CenterOnScreen()
	win.Resize(fyne.Size{
		Width:  800,
		Height: 800,
	})
	win.SetFixedSize(true)
	win.SetMaster()
	tabs := container.NewAppTabs()
	tabs.Append(container.NewTabItem(i18n.T("m.Calendar"), NewCalendarGui(win)))
	tabs.Append(container.NewTabItem(i18n.T("m.Folder"), NewFolderGui(win)))
	tabs.Append(container.NewTabItem(i18n.T("m.Metadata"), NewMetadataGui(win)))
	tabs.Append(container.NewTabItem(i18n.T("m.App"), NewAppGui(win)))
	win.SetContent(tabs)
	return win
}
