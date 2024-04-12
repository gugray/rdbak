package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const pageSize = 40
const maxFileNameLen = 128
const timeoutSec = 60
const loginUrl = "https://api.raindrop.io/v1/auth/email/login"
const listUrl = "https://api.raindrop.io/v1/raindrops/0?sort=-lastUpdate&perpage=%v&page=%v&version=2"
const downloadUrl = "https://api.raindrop.io/v1/raindrop/%v/cache?download"
const collsUrl = "https://api.raindrop.io/v1/collections"
const collsChildrenUrl = "https://api.raindrop.io/v1/collections/childrens"

type apiClient struct {
	jar            *cookieJar
	client         *http.Client
	reDownloadName *regexp.Regexp
}

func newApiClient() *apiClient {

	ac := apiClient{}
	ac.jar = newJar()
	ac.client = &http.Client{nil, nil, ac.jar, 0}
	ac.reDownloadName = regexp.MustCompile("attachment; filename=\"(.+)\"")
	return &ac
}

func (ac *apiClient) login(email, pass string) {

	payload := map[string]interface{}{"email": email, "password": pass}
	payloadStr, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", loginUrl, bytes.NewBuffer(payloadStr))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Bad status at login: %v: %s", resp.StatusCode, resp.Status))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var loginRes loginRes
	err = json.Unmarshal(body, &loginRes)
	if err != nil {
		panic(err)
	}
	if !loginRes.Result {
		panic(fmt.Sprintf("Login returned false: %v", loginRes.ErrorMessage))
	}
}

func (ac *apiClient) listBookmarks(page int) listRes {

	ctx, cancel := context.WithTimeout(context.Background(), timeoutSec*time.Second)
	defer cancel()
	url := fmt.Sprintf(listUrl, pageSize, page)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Accept", "application/json")

	resp, err := ac.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Bad status at list bookmarks: %v: %s", resp.StatusCode, resp.Status))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var listRes listRes
	err = json.Unmarshal(body, &listRes)
	if err != nil {
		panic(err)
	}
	if !listRes.Result {
		panic(fmt.Sprintf("List bookmarks returned false: %v", listRes.ErrorMessage))
	}
	return listRes
}

func limitLength(fn string, maxLen int) string {
	fnLen := len(fn)
	if fnLen <= maxLen {
		return fn
	}
	dotix := strings.LastIndex(fn, ".")
	if dotix == -1 {
		return fn[:maxLen]
	}
	extLen := fnLen - dotix
	if extLen >= maxLen {
		return fn[:maxLen]
	}
	res := fn[:maxLen-extLen] + fn[dotix:]
	return res
}

func safeDeleteFile(fn string) {
	if _, err := os.Stat(fn); err != nil {
		return
	}
	if err := os.Remove(fn); err != nil {
		fmt.Printf("Tried to delete %s, got error: %v", fn, err)
	}
}

func (ac *apiClient) getFileName(id uint64, resp *http.Response) string {

	// Baseline: file name is ID
	fn := fmt.Sprintf("%v", id)

	// Download file name is expected in a header
	if cdp := resp.Header.Get("Content-Disposition"); cdp != "" {
		groups := ac.reDownloadName.FindStringSubmatch(cdp)
		if groups != nil {
			name := limitLength(groups[1], maxFileNameLen)
			fn += "-" + name
		}
	}

	// Add extension based on mime type, if present
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		newExt := ""
		if strings.HasPrefix(ct, "application/pdf") {
			newExt = ".pdf"
		}
		if newExt != "" {
			fn = strings.TrimSuffix(fn, ".html")
			fn += newExt
		}
	}

	// Whee
	return fn
}

func (ac *apiClient) downloadFileIfMissing(id uint64, dir string) (bool, error) {

	etc := NewExtensibleTimeoutContext(timeoutSec)
	defer etc.Cancel()
	url := fmt.Sprintf(downloadUrl, id)
	req, _ := http.NewRequestWithContext(etc.Context(), "GET", url, nil)
	resp, err := ac.client.Do(req)
	if err != nil {
		fmt.Printf("Error creating client for %v\n%v\n", url, err)
		return false, err
	}
	defer resp.Body.Close()

	// If we don't get a 200 we don't panic. Maybe problem is transient and download
	// will work next time
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got status %v trying to download %v\n", resp.StatusCode, url)
		return false, err
	}

	fn := ac.getFileName(id, resp)
	fn = path.Join(dir, fn)

	if stat, err := os.Stat(fn); err == nil && stat.Size() != 0 {
		return false, nil
	}

	outf, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer outf.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("Error reading content from %v\n%v\n", url, err)
			outf.Close()
			safeDeleteFile(fn)
			return false, err
		}
		if n > 0 {
			outf.Write(buf[:n])
			etc.Extend()
		}
	}

	return true, nil
}
