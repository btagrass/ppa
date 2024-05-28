package mdl

import (
	"time"
)

type Photo struct {
	UID          string `gorm:"primaryKey"`
	Type         string
	TakenAt      time.Time
	TakenAtLocal time.Time
	Path         string
	Year         int
	Month        int
	FileRoot     string
	FileName     string
	Hash         string
	Files        []PhotoFile `gorm:"-"`
}

type PhotoFile struct {
	UID          string
	Name         string
	Size         int
	Primary      bool
	OriginalName string
	FileType     string
	Width        int
	Height       int
}

type CalendarPhoto struct {
	Date     string
	Images   []Photo
	Lives    []Photo
	Similars []Photo
}

type FolderPhoto struct {
	Date   string
	Photos []Photo
}
