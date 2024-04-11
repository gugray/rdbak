package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func Backup(cfg *Config) {

	fmt.Printf("%s Starting bookmarks backup\n", time.Now().Format("2006-01-02 15:04:05"))

	bookmarks := loadBookmarks(cfg.BookmarksFile)

	// Raindrop API client: start by logging in
	ac := newApiClient()
	ac.login(cfg.Email, cfg.Password)

	idToBookmark := make(map[uint64]*bookmark)
	for _, bm := range bookmarks {
		idToBookmark[bm.Id] = bm
	}

	// Get updated and new bookmarks
	changedBookmarks := getChangedBookmarks(ac, idToBookmark)

	// Download permanent copy where ready and file still missing
	downloadCount := 0
	failedIds := make(map[uint64]bool)
	for _, bm := range changedBookmarks {
		if bm.Cache.Status != "ready" {
			continue
		}
		downloaded, err := ac.downloadFileIfMissing(bm.Id, cfg.ExportDir)
		if err != nil {
			failedIds[bm.Id] = true
		}
		if downloaded {
			downloadCount++
		}
	}

	// Marge unchanged bookmarks with changed/new
	keptIds := make(map[uint64]bool)
	newBookmarks := make([]*bookmark, 0, len(bookmarks)+len(changedBookmarks))
	for _, bm := range changedBookmarks {
		if _, exists := failedIds[bm.Id]; exists {
			continue
		}
		newBookmarks = append(newBookmarks, bm)
		keptIds[bm.Id] = true
	}
	for _, bm := range bookmarks {
		if _, exists := failedIds[bm.Id]; exists {
			continue
		}
		if _, exists := keptIds[bm.Id]; exists {
			continue
		}
		newBookmarks = append(newBookmarks, bm)
		keptIds[bm.Id] = true
	}

	// Save updated bookmarks JSON
	saveBookmarks(cfg.BookmarksFile, newBookmarks)

	// Report
	fmt.Printf("Finished. %v bookmark(s) new or changed; %v new file(s) downloaded.\n", len(changedBookmarks), downloadCount)
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

func getChangedBookmarks(ac *apiClient, storedBookmarks map[uint64]*bookmark) []*bookmark {

	var res []*bookmark

	page := 0
	for {
		lr := ac.listBookmarks(page)
		over := len(lr.Items) < pageSize
		for _, itm := range lr.Items {
			if storedBm, exists := storedBookmarks[itm.Id]; !exists {
				res = append(res, itm)
			} else if itm.LastUpdate.After(storedBm.LastUpdate) {
				res = append(res, itm)
			}
		}
		// DBG
		//over = true
		if over {
			break
		}
		page += 1
	}

	return res
}
