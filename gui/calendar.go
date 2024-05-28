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
	"github.com/btagrass/gobiz/utl"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

func NewCalendarGui(win fyne.Window) fyne.CanvasObject {
	resolutionLabel := &widget.Label{
		Text: i18n.T("l.Resolution"),
	}
	resolutionEntry := &widget.Entry{
		Text:      "4032 × 3024,3024 × 4032,2316 × 3088",
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.Resolution")),
	}
	resolutionContainer := container.NewHBox(resolutionLabel, resolutionEntry)
	errorValueLabel := &widget.Label{
		Text: i18n.T("l.ErrorValue"),
	}
	errorValueEntry := &widget.Entry{
		Text:      "1",
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\d+", i18n.T("r.InputCorrect", "$l.ErrorValue")),
	}
	errorValueContainer := container.NewHBox(errorValueLabel, errorValueEntry)
	latencyTimeLabel := &widget.Label{
		Text: i18n.T("l.LatencyTime"),
	}
	latencyTimeEntry := &widget.Entry{
		Text:      "120s",
		Scroll:    container.ScrollNone,
		Validator: validation.NewRegexp("\\w+", i18n.T("r.Input", "$l.LatencyTime")),
	}
	latencyTimeContainer := container.NewHBox(latencyTimeLabel, latencyTimeEntry)
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
	var calendarPhotos []mdl.CalendarPhoto
	cols := map[int]map[string]any{
		0: {
			"header": i18n.T("l.Date"),
			"width":  150,
		},
		1: {
			"header": i18n.T("l.Images"),
			"width":  140,
		},
		2: {
			"header": i18n.T("l.Lives"),
			"width":  140,
		},
		3: {
			"header": i18n.T("l.Similars"),
			"width":  140,
		},
		4: {
			"header": "",
			"width":  100,
		},
		5: {
			"header": "",
			"width":  100,
		},
	}
	calendarTable := widget.NewTable(
		func() (int, int) {
			return len(calendarPhotos) + 1, len(cols)
		},
		func() fyne.CanvasObject {
			return &widget.Label{}
		},
		func(id widget.TableCellID, template fyne.CanvasObject) {
			label := template.(*widget.Label)
			if id.Row == 0 {
				label.Alignment = fyne.TextAlignCenter
				label.SetText(cast.ToString(cols[id.Col]["header"]))
			} else {
				label.Wrapping = fyne.TextWrapBreak
				calendar := calendarPhotos[id.Row-1]
				if id.Col == 0 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(calendar.Date)
				} else if id.Col == 1 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(cast.ToString(len(calendar.Images)))
				} else if id.Col == 2 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(cast.ToString(len(calendar.Lives)))
				} else if id.Col == 3 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(cast.ToString(len(calendar.Similars)))
				} else if id.Col == len(cols)-2 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(i18n.T("l.Check"))
				} else if id.Col == len(cols)-1 {
					label.Alignment = fyne.TextAlignCenter
					label.SetText(i18n.T("l.Clean"))
				}
			}
		},
	)
	calendarTable.OnSelected = func(id widget.TableCellID) {
		if id.Col == len(cols)-2 {
			err := latencyTimeEntry.Validate()
			if err != nil {
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				return
			}
			go func() {
				err = s.Use[*svc.PhotoSvc]().ViewCalendarPhotos(calendarPhotos[id.Row-1].Date, calendarPhotos[id.Row-1], latencyTimeEntry.Text)
				if err != nil {
					logrus.Error(err)
					dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				}
			}()
		} else if id.Col == len(cols)-1 {
			err := resolutionEntry.Validate()
			if err != nil {
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				return
			}
			err = errorValueEntry.Validate()
			if err != nil {
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
				return
			}
			err = s.Use[*svc.PhotoSvc]().ClearCalendarPhotos(calendarPhotos[id.Row-1].Date, utl.Split(errorValueEntry.Text, ',', '，'), cast.ToInt(errorValueEntry.Text))
			if err != nil {
				logrus.Error(err)
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			} else {
				dialog.ShowInformation(i18n.T("l.Info"), i18n.T("l.CleaningCompleted"), win)
			}
		}
	}
	for i := 0; i < len(cols); i++ {
		calendarTable.SetColumnWidth(i, cast.ToFloat32(cols[i]["width"]))
	}
	calendarContainer := container.NewScroll(calendarTable)
	calendarContainer.SetMinSize(fyne.NewSize(600, 600))
	checkButton := &widget.Button{
		Text: i18n.T("l.Check"),
		Icon: theme.SearchIcon(),
	}
	checkButton.OnTapped = func() {
		err := resolutionEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		err = errorValueEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		err = latencyTimeEntry.Validate()
		if err != nil {
			dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			return
		}
		go func() {
			checkButton.Disable()
			clear(calendarPhotos)
			calendarTable.Refresh()
			calendarPhotos, err = s.Use[*svc.PhotoSvc]().CheckCalendarPhotos(dateEntry.Text, cast.ToInt(errorValueEntry.Text))
			calendarTable.Refresh()
			if err != nil {
				logrus.Error(err)
				dialog.ShowInformation(i18n.T("l.Error"), err.Error(), win)
			}
			checkButton.Enable()
		}()
	}
	return container.NewVBox(
		resolutionContainer,
		errorValueContainer,
		latencyTimeContainer,
		dateContainer,
		checkButton,
		calendarContainer,
	)
}
