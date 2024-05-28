package gui

import (
	"fmt"
	"path/filepath"
	"ppa/svc"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/btagrass/gobiz/i18n"
	s "github.com/btagrass/gobiz/svc"
	"github.com/btagrass/gobiz/utl"
	"github.com/sirupsen/logrus"
)

func NewMetadataGui(win fyne.Window) fyne.CanvasObject {
	dirLabel := &widget.Label{
		Text: i18n.T("l.Folder"),
	}
	dirEntry := &widget.Label{}
	openButton := &widget.Button{
		Icon: theme.FolderOpenIcon(),
		OnTapped: func() {
			dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
				if uri == nil {
					dirEntry.SetText("")
					return
				}
				dirEntry.SetText(uri.Path())
			}, win)
		},
	}
	dirContainer := container.NewHBox(dirLabel, dirEntry, openButton)
	dateTimeLabel := &widget.Label{
		Text: i18n.T("l.DateTime"),
	}
	dateTimeEntry := &widget.Entry{
		Text:      "1998-01-01 09:00:00",
		Scroll:    container.ScrollNone,
		Validator: validation.NewTime(time.DateTime),
	}
	resetButton := &widget.Button{
		Icon: theme.ContentUndoIcon(),
		OnTapped: func() {
			dateTimeEntry.SetText("1998-01-01 09:00:00")
		},
	}
	dateTimeContainer := container.NewHBox(dateTimeLabel, dateTimeEntry, resetButton)
	saveButton := &widget.Button{
		Text: i18n.T("l.Save"),
		Icon: theme.DocumentSaveIcon(),
	}
	saveButton.OnTapped = func() {
		if dirEntry.Text == "" {
			dialog.ShowInformation(i18n.T("l.Error"), i18n.T("r.Select", "$l.Folder"), win)
			return
		}
		err := dateTimeEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		go func() {
			saveButton.Disable()
			files := utl.Glob(fmt.Sprintf("%s/*.jpg", dirEntry.Text), fmt.Sprintf("%s/*.png", dirEntry.Text), fmt.Sprintf("%s/*.mov", dirEntry.Text), fmt.Sprintf("%s/*.mp4", dirEntry.Text))
			for _, f := range files {
				if utl.Contains(f, ".meta.") {
					continue
				}
				filePath := utl.InsertBefore(f, filepath.Ext(f), ".meta")
				err = s.Use[*svc.PhotoSvc]().SavePhotoTime(f, filePath, dateTimeEntry.Text)
				if err != nil {
					logrus.Error(err)
				}
			}
			dialog.ShowInformation(i18n.T("l.Info"), i18n.T("l.SaveSuccessful", len(files)), win)
			saveButton.Enable()
		}()
	}
	return container.NewVBox(dirContainer, dateTimeContainer, saveButton)
}
