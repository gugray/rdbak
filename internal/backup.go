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

	fmt.Printf("%s Starting bookmarks backup\n", time.Now().Format("2006-01-02 15:04:05"))

	bookmarks := loadBookmarks(cfg.BookmarksFile)

	// Raindrip API client: start by logging in
	ac := newApiClient()
	ac.login(cfg.Email, cfg.Password)

	// We'll be looking for "lastUpdate" times later then we last saw
	var latest time.Time
	for _, bm := range bookmarks {
		if bm.LastUpdate.After(latest) {
			latest = bm.LastUpdate
		}
	}

	// Get updated and new bookmarks
	changedBookmarks := getChangedBookmarks(ac, latest)

	// Download permanent copy where ready and file still missing
	downloadCount := 0
	for _, bm := range changedBookmarks {
		if bm.Cache.Status != "ready" {
			continue
		}
		downloaded := downloadIfMissing(ac, cfg.ExportDir, bm.Id)
		if downloaded {
			downloadCount++
		}
	}

	// Marge unchanged bookmarks with changed/new
	ids := make(map[uint64]bool)
	newBookmarks := make([]*bookmark, 0, len(bookmarks)+len(changedBookmarks))
	for _, bm := range changedBookmarks {
		newBookmarks = append(newBookmarks, bm)
		ids[bm.Id] = true
	}
	for _, bm := range bookmarks {
		if _, exists := ids[bm.Id]; exists {
			continue
		}
		newBookmarks = append(newBookmarks, bm)
		ids[bm.Id] = true
	}

	// Save updated bookmarks JSON
	saveBookmarks(cfg.BookmarksFile, newBookmarks)

	// Report
	fmt.Printf("Finished. %v bookmark(s) changed; %v new file(s) downloaded.\n", len(changedBookmarks), downloadCount)
}

func downloadIfMissing(ac *apiClient, dir string, id uint64) bool {

	fn := fmt.Sprintf("%v.html", id)
	fn = path.Join(dir, fn)

	if stat, err := os.Stat(fn); err == nil && stat.Size() != 0 {
		return false
	}

	return ac.downloadFile(id, fn)
}

func saveBookmarks(fn string, bookmarks []*bookmark) {

	json, err := json.MarshalIndent(bookmarks, "", "  ")
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

func getChangedBookmarks(ac *apiClient, lastSeen time.Time) []*bookmark {

	var res []*bookmark

	page := 0
	for {
		lr := ac.listBookmarks(page)
		over := len(lr.Items) < pageSize
		for _, itm := range lr.Items {
			if itm.LastUpdate.After(lastSeen) {
				res = append(res, itm)
			} else {
				over = true
			}
		}
		if over {
			break
		}
		page += 1
	}

	return res
}
