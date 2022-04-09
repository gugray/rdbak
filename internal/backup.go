package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func Backup(cfg *Config) {

	bookmarks := loadBookmarks(cfg.BookmarksFile)

	ac := newApiClient()
	ac.login(cfg.Email, cfg.Password)

	var latest time.Time
	for _, bm := range bookmarks {
		if bm.Created.After(latest) {
			latest = bm.Created
		}
	}

	newBookmarks := getNewBookmarks(ac, latest)

	for _, bm := range newBookmarks {
		download(ac, cfg.ExportDir, bm.Id)
	}

	bookmarks = append(bookmarks, newBookmarks...)
	saveBookmarks(cfg.BookmarksFile, bookmarks)
}

func download(ac *apiClient, dir string, id uint64) {

	fn := fmt.Sprintf("%v.html", id)
	fn = path.Join(dir, fn)
	ac.downloadFile(id, fn)
}

func saveBookmarks(fn string, bookmarks []*bookmark) {

	json, err := json.Marshal(bookmarks)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fn, json, 0666)
	if err != nil {
		panic(err)
	}
}

func loadBookmarks(fn string) []*bookmark {

	bookmarks := make([]*bookmark, 0)

	if _, err := os.Stat(fn); err == nil {
		bookmarksJson, err := ioutil.ReadFile(fn)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(bookmarksJson, &bookmarks)
		if err != nil {
			panic(err)
		}
	}

	return bookmarks
}

func getNewBookmarks(ac *apiClient, lastSeen time.Time) []*bookmark {

	var res []*bookmark

	page := 0
	for {
		lr := ac.listBookmarks(page)
		over := len(lr.Items) < pageSize
		for _, itm := range lr.Items {
			if itm.Created.After(lastSeen) {
				res = append(res, itm)
			} else {
				over = true
			}
		}
		if over {
			break
		}
		page += 1
		// DBG: Quit after first page
		break
	}

	return res
}
