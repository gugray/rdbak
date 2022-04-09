package internal

import "time"

type bookmark struct {
	Id           uint64    `json:"_id"`
	Link         string    `json:"link"`
	Title        string    `json:"title"`
	Excerpt      string    `json:"excerpt"`
	Type         string    `json:"type"`
	Created      time.Time `json:"created"`
	CollectionId uint64    `json:"collectionId"`
}

type listRes struct {
	Result       bool        `json:"result"'`
	Items        []*bookmark `json:"items"'`
	Count        int         `json:"count"`
	ErrorMessage string      `json:"errorMessage"'`
}

type loginRes struct {
	Result       bool   `json:"result"'`
	ErrorMessage string `json:"errorMessage"'`
}
