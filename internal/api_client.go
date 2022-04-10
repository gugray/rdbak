package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const pageSize = 40
const loginUrl = "https://api.raindrop.io/v1/auth/email/login"
const listUrl = "https://api.raindrop.io/v1/raindrops/0?sort=-lastUpdate&perpage=%v&page=%v&version=2"
const downloadUrl = "https://api.raindrop.io/v1/raindrop/%v/cache"
const collsUrl = "https://api.raindrop.io/v1/collections"
const collsChildrenUrl = "https://api.raindrop.io/v1/collections/childrens"

type apiClient struct {
	jar    *cookieJar
	client *http.Client
}

func newApiClient() *apiClient {

	ac := apiClient{}
	ac.jar = newJar()
	ac.client = &http.Client{nil, nil, ac.jar, 30 * time.Second}
	return &ac
}

func (ac *apiClient) login(email, pass string) {

	payload := map[string]interface{}{"email": email, "password": pass}
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", loginUrl, bytes.NewBuffer(payloadStr))
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

	url := fmt.Sprintf(listUrl, pageSize, page)
	req, err := http.NewRequest("GET", url, nil)
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

func (ac *apiClient) downloadFile(id uint64, fn string) bool {

	url := fmt.Sprintf(downloadUrl, id)

	resp, err := ac.client.Get(url)
	if err != nil {
		fmt.Printf("Error creating client for %v\n%v\n", url, err)
		return false
	}
	defer resp.Body.Close()

	// If we don't get a 200 we don't panic. Maybe problem is transient and download
	// will work next time
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got status %v trying to download %v\n", resp.StatusCode, url)
		return false
	}

	outf, err := os.Create(fn)
	if err != nil {
		panic(err)
	}
	defer outf.Close()

	_, err = io.Copy(outf, resp.Body)
	if err != nil {
		fmt.Printf("Error reading content from %v\n%v\n", url, err)
		return false
	}

	return true
}
