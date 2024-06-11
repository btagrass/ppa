package gui

import (
	"ppa/mdl"
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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func NewFolderGui(win fyne.Window) fyne.CanvasObject {
	durationLabel := &widget.Label{
		Text: i18n.T("l.LatencyTime"),
	}
	durationEntry := &widget.Entry{
		Text:      "120s",
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.LatencyTime")),
	}
	durationContainer := container.NewHBox(durationLabel, durationEntry)
	dateLabel := &widget.Label{
		Text: i18n.T("l.Date"),
	}
	dateEntry := &widget.Entry{
		Text:   time.Now().Format("2006-01"),
		Scroll: container.ScrollNone,
	}
	clearButton := &widget.Button{
		Icon: theme.ContentClearIcon(),
		OnTapped: func() {
			dateEntry.SetText("")
		},
	}
	dateContainer := container.NewHBox(dateLabel, dateEntry, clearButton)
	var folderPhotos []mdl.FolderPhoto
	cols := map[int]map[string]any{
		0: {
			"header": i18n.T("l.Date"),
			"width":  150,
		},
		1: {
			"header": i18n.T("l.Photos"),
			"width":  520,
		},
		2: {
			"header": "",
			"width":  100,
		},
	}
	folderTable := widget.NewTable(
		func() (int, int) {
			return len(folderPhotos) + 1, len(cols)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, template fyne.CanvasObject) {
			label := template.(*widget.Label)
			if id.Row == 0 {
				label.Alignment = fyne.TextAlignCenter
				label.SetText(cast.ToString(cols[id.Col]["header"]))
			} else {
				label.Wrapping = fyne.TextWrapBreak
				photo := folderPhotos[id.Row-1]
				if id.Col == 0 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(photo.Date)
				} else if id.Col == 1 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(cast.ToString(len(photo.Photos)))
				} else if id.Col == len(cols)-1 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(i18n.T("l.Check"))
				}
			}
		},
	)
	folderTable.OnSelected = func(id widget.TableCellID) {
		if id.Col == len(cols)-1 {
			err := durationEntry.Validate()
			if err != nil {
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				return
			}
			go func() {
				err = s.Use[*svc.PhotoSvc]().ViewFolderPhotos(folderPhotos[id.Row-1].Date, folderPhotos[id.Row-1], durationEntry.Text)
				if err != nil {
					logrus.Error(err)
					dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				}
			}()
		}
	}
	for i := 0; i < len(cols); i++ {
		folderTable.SetColumnWidth(i, cast.ToFloat32(cols[i]["width"]))
	}
	folderContainer := container.NewScroll(folderTable)
	folderContainer.SetMinSize(fyne.NewSize(600, 600))
	checkButton := &widget.Button{
		Text: i18n.T("l.Check"),
		Icon: theme.SearchIcon(),
	}
	checkButton.OnTapped = func() {
		go func() {
			checkButton.Disable()
			clear(folderPhotos)
			folderTable.Refresh()
			var err error
			folderPhotos, err = s.Use[*svc.PhotoSvc]().CheckFolderPhotos(dateEntry.Text)
			folderTable.Refresh()
			if err != nil {
				logrus.Error(err)
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			} else {
				logrus.Info(folderPhotos)
			}
			checkButton.Enable()
		}()
	}
	return container.NewVBox(
		durationContainer,
		dateContainer,
		checkButton,
		folderContainer,
	)
}
