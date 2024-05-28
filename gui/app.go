package gui

import (
	"ppa/svc"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/btagrass/gobiz/i18n"
	s "github.com/btagrass/gobiz/svc"
	"github.com/sirupsen/logrus"
)

func NewAppGui(win fyne.Window) fyne.CanvasObject {
	app := s.Use[*svc.AppSvc]().GetApp()
	langLabel := &widget.Label{
		Text: i18n.T("l.Lang"),
	}
	langSelect := &widget.Select{
		Selected: app.Lang,
		Options:  []string{"English", "中文"},
	}
	langTip := &widget.Label{
		Text:       i18n.T("l.LangTip"),
		Importance: widget.HighImportance,
	}
	langContainer := container.NewHBox(langLabel, langSelect, langTip)
	divider := &widget.Label{
		Text:     strings.Repeat("-", 200),
		Wrapping: fyne.TextTruncate,
	}
	urlLabel := &widget.Label{
		Text: i18n.T("l.Url"),
	}
	urlEntry := &widget.Entry{
		Text:      app.Url,
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.Url")),
	}
	urlContainer := container.NewHBox(urlLabel, urlEntry)
	userNameLabel := &widget.Label{
		Text: i18n.T("l.UserName"),
	}
	userNameEntry := &widget.Entry{
		Text:      app.UserName,
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.UserName")),
	}
	userNameContainer := container.NewHBox(userNameLabel, userNameEntry)
	passwordLabel := &widget.Label{
		Text: i18n.T("l.Password"),
	}
	passwordEntry := &widget.Entry{
		Text:      app.Password,
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.Password")),
	}
	passwordContainer := container.NewHBox(passwordLabel, passwordEntry)
	saveButton := &widget.Button{
		Text: i18n.T("l.Save"),
		Icon: theme.DocumentSaveIcon(),
	}
	saveButton.OnTapped = func() {
		err := urlEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		err = userNameEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		err = passwordEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		go func() {
			saveButton.Disable()
			app.Lang = langSelect.Selected
			app.Url = urlEntry.Text
			app.UserName = userNameEntry.Text
			app.Password = passwordEntry.Text
			err = s.Use[*svc.AppSvc]().Save(*app)
			if err != nil {
				logrus.Error(err)
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			}
			saveButton.Enable()
		}()
	}
	return container.NewVBox(
		langContainer,
		divider,
		urlContainer,
		userNameContainer,
		passwordContainer,
		saveButton,
	)
}
