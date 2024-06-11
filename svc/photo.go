package svc

import (
	"fmt"
	"image"
	"math/bits"
	"os"
	"path/filepath"
	"ppa/mdl"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/btagrass/gobiz/htp"
	"github.com/btagrass/gobiz/svc"
	"github.com/btagrass/gobiz/uat"
	"github.com/btagrass/gobiz/utl"
	"github.com/corona10/goimagehash"
	"github.com/go-rod/rod"
	"github.com/samber/do"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type PhotoSvc struct {
	*svc.DataSvc[mdl.Photo]
	appSvc *AppSvc
	count  int
}

func NewPhotoSvc(i *do.Injector) (*PhotoSvc, error) {
	return &PhotoSvc{
		DataSvc: svc.NewDataSvc[mdl.Photo]("ppa:photos"),
		appSvc:  svc.Use[*AppSvc](),
		count:   50000,
	}, nil
}

func (s *PhotoSvc) CalcImageHash(filePath string) (uint64, error) {
	if !utl.HasSuffix(filePath, ".jpg", ".png") {
		return 0, fmt.Errorf("format not supported")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	image, _, err := image.Decode(file)
	if err != nil {
		return 0, err
	}
	hash, err := goimagehash.PerceptionHash(image)
	if err != nil {
		return 0, err
	}
	return hash.GetHash(), nil
}

func (s *PhotoSvc) CheckCalendarPhotos(date string, errorValue int) ([]mdl.CalendarPhoto, error) {
	defer utl.ElapsedTime("CheckCalendarPhotos", time.Now())
	calendarPhotos := make([]mdl.CalendarPhoto, 0)
	dataPhotos, _, err := s.List()
	if err != nil {
		return nil, err
	}
	photos, err := s.ListCalendarPhotos(date)
	if err != nil {
		return nil, err
	}
	for k, v := range photos {
		images := lo.Filter(v, func(item mdl.Photo, index int) bool {
			return item.Type == "image" && (item.FileRoot == "sidecar" && len(item.Files) > 2 || item.FileRoot == "/" && len(item.Files) > 1)
		})
		lives := lo.Filter(v, func(item mdl.Photo, index int) bool {
			return item.Type == "live"
		})
		var similars []mdl.Photo
		hashPhotos := utl.Intersect(dataPhotos, v, func(t1, t2 mdl.Photo) bool {
			return t1.UID == t2.UID
		})
		slices.SortFunc(hashPhotos, func(a, b mdl.Photo) int {
			aHash, _ := strconv.ParseUint(a.Hash, 10, 64)
			bHash, _ := strconv.ParseUint(b.Hash, 10, 64)
			return int(aHash - bHash)
		})
		for i := 0; i < len(hashPhotos)-1; i++ {
			iHash, _ := strconv.ParseUint(hashPhotos[i].Hash, 10, 64)
			jHash, _ := strconv.ParseUint(hashPhotos[i+1].Hash, 10, 64)
			distance := bits.OnesCount64(iHash ^ jHash)
			if distance <= errorValue {
				similars = append(similars, hashPhotos[i], hashPhotos[i+1])
			}
		}
		similars = lo.UniqBy(similars, func(item mdl.Photo) string {
			return item.UID
		})
		if len(images) > 0 || len(lives) > 0 || len(similars) > 0 {
			calendarPhotos = append(calendarPhotos, mdl.CalendarPhoto{
				Date:     k,
				Images:   images,
				Lives:    lives,
				Similars: similars,
			})
		}
	}
	return calendarPhotos, nil
}

func (s *PhotoSvc) CheckFolderPhotos(date string) ([]mdl.FolderPhoto, error) {
	defer utl.ElapsedTime("CheckFolderPhotos", time.Now())
	photos, err := s.ListFolderPhotos(date)
	if err != nil {
		return nil, err
	}
	var folderPhotos []mdl.FolderPhoto
	for k, v := range photos {
		uv := lo.Filter(v, func(item mdl.Photo, _ int) bool {
			if item.Year == -1 && item.Month == -1 && !utl.HasSuffix(item.FileName, ".mov.jpg", ".mp4.jpg") {
				return true
			}
			if strings.Contains(item.FileName, "20240131_063835_066D9633.jpg") {
				println(item.Files)
			}
			return lo.ContainsBy(item.Files, func(i mdl.PhotoFile) bool {
				return i.FileType == "aae"
			})
		})
		if len(uv) > 0 {
			folderPhotos = append(folderPhotos, mdl.FolderPhoto{
				Date:   k,
				Photos: uv,
			})
		}
	}
	return folderPhotos, nil
}

func (s *PhotoSvc) ClearCalendarPhotos(date string, rates []string, errorValue int) error {
	app, err := s.appSvc.LoginApp()
	if err != nil {
		return err
	}
	calendarPhotos, err := s.CheckCalendarPhotos(date, errorValue)
	if err != nil {
		return err
	}
	for _, p := range calendarPhotos {
		for _, i := range p.Images {
			slices.SortFunc(i.Files, func(a, b mdl.PhotoFile) int {
				if a.FileType == "heic" {
					return 1
				} else if b.FileType == "heic" {
					return -1
				}
				if utl.HasSuffix(a.Name, "00001.jpg") || utl.HasSuffix(a.OriginalName, "00001.jpg") {
					return -1
				} else if utl.HasSuffix(b.Name, "00001.jpg") || utl.HasSuffix(b.OriginalName, "00001.jpg") {
					return 1
				}
				aRate := fmt.Sprintf("%d × %d", a.Width, a.Height)
				aRateOk := lo.Contains(rates, aRate)
				bRate := fmt.Sprintf("%d × %d", b.Width, b.Height)
				bRateOk := lo.Contains(rates, bRate)
				if aRateOk && bRateOk || !aRateOk && !bRateOk {
					if aRate > bRate {
						return -1
					} else if aRate < bRate {
						return 1
					} else if a.Size > b.Size {
						return -1
					} else if a.Size < b.Size {
						return 1
					}
				} else if aRateOk {
					return -1
				} else if bRateOk {
					return 1
				}
				return 0
			})
			for j, f := range i.Files {
				if j == 0 {
					if f.Primary {
						continue
					}
					_, err = htp.Post(fmt.Sprintf("%s/api/v1/photos/%s/files/%s/primary", app.Url, i.UID, f.UID), map[string]string{
						"X-Auth-Token": app.Token,
					}, nil)
					if err != nil {
						return err
					}
				} else {
					_, err = htp.Delete(fmt.Sprintf("%s/api/v1/photos/%s/files/%s", app.Url, i.UID, f.UID), map[string]string{
						"X-Auth-Token": app.Token,
					}, nil)
					if err != nil {
						return err
					}
				}
			}
		}
		for _, l := range p.Lives {
			_, err = htp.Post(fmt.Sprintf("%s/api/v1/batch/photos/archive", app.Url), map[string]string{
				"X-Auth-Token": app.Token,
			}, map[string][]string{
				"photos": {l.UID},
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PhotoSvc) ListCalendarPhotos(date string) (map[string][]mdl.Photo, error) {
	app, err := s.appSvc.LoginApp()
	if err != nil {
		return nil, err
	}
	var albums []mdl.Album
	_, err = htp.Get(fmt.Sprintf("%s/api/v1/albums?count=%d&offset=0&q=&category=&order=newest&year=&type=month", app.Url, s.count), map[string]string{
		"X-Auth-Token": app.Token,
	}, &albums)
	if err != nil {
		return nil, err
	}
	if date != "" {
		albums = lo.Filter(albums, func(item mdl.Album, index int) bool {
			dateTime, err := time.Parse("January-2006", strings.ToTitle(item.Slug))
			if err != nil {
				return false
			}
			return dateTime.Format("2006-01") == date
		})
	}
	calendarPhotos := make(map[string][]mdl.Photo)
	for _, a := range albums {
		var photos []mdl.Photo
		_, err = htp.Get(fmt.Sprintf("%s/api/v1/photos?count=%d&offset=0&s=%s&merged=true&country=&camera=0&order=oldest&q=", app.Url, s.count, a.UID), map[string]string{
			"X-Auth-Token": app.Token,
		}, &photos)
		if err != nil {
			logrus.Error(err)
			continue
		}
		for _, p := range photos {
			_, err = htp.Get(fmt.Sprintf("%s/api/v1/photos/%s", app.Url, p.UID), map[string]string{
				"X-Auth-Token": app.Token,
			}, &p)
			if err != nil {
				logrus.Error(err)
				continue
			}
		}
		dateTime, err := time.Parse("January-2006", strings.ToTitle(a.Slug))
		if err != nil {
			return nil, err
		}
		calendarPhotos[dateTime.Format("2006-01")] = photos
	}
	return calendarPhotos, nil
}

func (s *PhotoSvc) ListFolderPhotos(date string) (map[string][]mdl.Photo, error) {
	app, err := s.appSvc.LoginApp()
	if err != nil {
		return nil, err
	}
	var albums []mdl.Album
	_, err = htp.Get(fmt.Sprintf("%s/api/v1/albums?count=%d&offset=0&q=&category=&order=name&year=&type=folder", app.Url, s.count), map[string]string{
		"X-Auth-Token": app.Token,
	}, &albums)
	if err != nil {
		return nil, err
	}
	if date != "" {
		dateTime, err := time.Parse("2006-01", date)
		if err != nil {
			return nil, err
		}
		albums = lo.Filter(albums, func(item mdl.Album, _ int) bool {
			return item.Slug == dateTime.Format("2006-01")
		})
	}
	folderPhotos := make(map[string][]mdl.Photo)
	for _, a := range albums {
		var photos []mdl.Photo
		_, err = htp.Get(fmt.Sprintf("%s/api/v1/photos?count=%d&offset=0&s=%s&merged=true&country=&camera=0&order=added&q=", app.Url, s.count, a.UID), map[string]string{
			"X-Auth-Token": app.Token,
		}, &photos)
		if err != nil {
			logrus.Error(err)
			continue
		}
		folderPhotos[a.Slug] = photos
	}
	return folderPhotos, nil
}

func (s *PhotoSvc) ListSearchPhotos() ([]mdl.Photo, error) {
	app, err := s.appSvc.LoginApp()
	if err != nil {
		return nil, err
	}
	var photos []mdl.Photo
	_, err = htp.Get(fmt.Sprintf("%s/api/v1/photos?count=%d&offset=0&merged=true&country=&camera=0&lens=0&label=&latlng=&year=0&month=0&color=&order=newest&q=&public=true", app.Url, s.count), map[string]string{
		"X-Auth-Token": app.Token,
	}, &photos)
	photos = lo.Filter(photos, func(item mdl.Photo, index int) bool {
		return !utl.HasSuffix(item.FileName, ".heic.jpg", ".avi.jpg", ".mov.jpg", ".mp4.jpg")
	})
	return photos, err
}

func (s *PhotoSvc) SavePhotoTime(inFilePath, outFilePath, dateTime string) error {
	data := make(map[string]any)
	if utl.HasSuffix(inFilePath, ".jpg", ".png") {
		data["DateTimeOriginal"] = dateTime
		data["DateTimeDigitized"] = dateTime
	} else if utl.HasSuffix(inFilePath, ".mov", ".mp4") {
		data["CreationTime"] = dateTime
	} else {
		return fmt.Errorf("%s file not supported", filepath.Ext(inFilePath))
	}
	return utl.WriteMetadata(inFilePath, outFilePath, data)
}

func (s *PhotoSvc) ViewCalendarPhotos(date string, photo mdl.CalendarPhoto, duration string) error {
	app := s.appSvc.GetApp()
	browser, err := uat.NewBrowser(app.Url, true, "", map[string]string{
		"start-maximized": "true",
	})
	if err != nil {
		return err
	}
	defer browser.Close()
	err = browser.Input("#auth-username", false, app.UserName)
	if err != nil {
		return err
	}
	err = browser.Input("#auth-password", false, app.Password)
	if err != nil {
		return err
	}
	err = browser.Click(".action-confirm", false, "1s")
	if err != nil {
		return err
	}
	err = browser.Goto("/library/calendar")
	if err != nil {
		return err
	}
	dateTime, err := time.Parse("2006-01", date)
	if err != nil {
		return err
	}
	uid, err := browser.Attr(fmt.Sprintf("//button[contains(text(),'%s')]", dateTime.Format("2006年1月")), true, "data-uid")
	if err != nil {
		return err
	}
	err = browser.Goto(fmt.Sprintf("/library/calendar/%s/view", *uid))
	if err != nil {
		return err
	}
	_, err = browser.ElementsFunc([]string{
		".is-photo",
	}, true, func(selector string, element *rod.Element) bool {
		uid, _ := element.Attribute("data-uid")
		if uid != nil {
			if lo.ContainsBy(photo.Images, func(item mdl.Photo) bool {
				return item.UID == *uid
			}) {
				_, err = element.Eval("() => this.style.outline='3px solid red'")
				if err != nil {
					return false
				}
			} else if lo.ContainsBy(photo.Lives, func(item mdl.Photo) bool {
				return item.UID == *uid
			}) {
				_, err = element.Eval("() => this.style.outline='3px solid green'")
				if err != nil {
					return false
				}
			} else if lo.ContainsBy(photo.Similars, func(item mdl.Photo) bool {
				return item.UID == *uid
			}) {
				_, err = element.Eval("() => this.style.outline='3px solid blue'")
				if err != nil {
					return false
				}
			} else {
				err = element.Remove()
				if err != nil {
					return false
				}
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	time.Sleep(cast.ToDuration(duration))
	return nil
}

func (s *PhotoSvc) ViewFolderPhotos(date string, photo mdl.FolderPhoto, duration string) error {
	app, err := s.appSvc.LoginApp()
	if err != nil {
		return err
	}
	browser, err := uat.NewBrowser(app.Url, true, "", map[string]string{
		"start-maximized": "true",
	})
	if err != nil {
		return err
	}
	defer browser.Close()
	err = browser.Input("#auth-username", false, app.UserName)
	if err != nil {
		return err
	}
	err = browser.Input("#auth-password", false, app.Password)
	if err != nil {
		return err
	}
	err = browser.Click(".action-confirm", false, "1s")
	if err != nil {
		return err
	}
	err = browser.Goto("/library/folders")
	if err != nil {
		return err
	}
	dateTime, err := time.Parse("2006-01", date)
	if err != nil {
		return err
	}
	uid, err := browser.Attr(fmt.Sprintf("//button[contains(text(),'%s')]", dateTime.Format("January 2006")), true, "data-uid")
	if err != nil {
		return err
	}
	err = browser.Goto(fmt.Sprintf("/library/folders/%s/view", *uid))
	if err != nil {
		return err
	}
	_, err = browser.ElementsFunc([]string{
		".is-photo",
	}, true, func(selector string, element *rod.Element) bool {
		uid, _ := element.Attribute("data-uid")
		if uid != nil {
			if lo.ContainsBy(photo.Photos, func(item mdl.Photo) bool {
				return item.UID == *uid
			}) {
				_, err = element.Eval("() => this.style.outline='3px solid red'")
				if err != nil {
					return false
				}
			} else {
				err = element.Remove()
				if err != nil {
					return false
				}
			}
		}
		return true
	})
	if err != nil {
		return err
	}
	time.Sleep(cast.ToDuration(duration))
	return nil
}
