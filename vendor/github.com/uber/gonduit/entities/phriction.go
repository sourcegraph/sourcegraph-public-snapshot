package entities

import "github.com/uber/gonduit/util"

// PhrictionDocument represents a document in Phriction wiki.
type PhrictionDocument struct {
	PHID         string             `json:"phid"`
	URI          string             `json:"uri"`
	Slug         string             `json:"slug"`
	Version      int                `json:"version,string"`
	AuthorPHID   string             `json:"authorPHID"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	Status       string             `json:"status"`
	Description  string             `json:"description"`
	DateModified util.UnixTimestamp `json:"dateModified"`
}
