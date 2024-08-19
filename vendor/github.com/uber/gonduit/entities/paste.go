package entities

import "github.com/uber/gonduit/util"

// PasteItem is a result item for paste queries.
type PasteItem struct {
	ID          uint64             `json:"id"`
	ObjectName  string             `json:"objectName"`
	PHID        string             `json:"phid"`
	AuthorPHID  string             `json:"authorPHID"`
	FilePHID    string             `json:"filePHID"`
	Title       string             `json:"title"`
	DateCreated util.UnixTimestamp `json:"dateCreated"`
	Language    string             `json:"language"`
	URI         string             `json:"uri"`
	ParentPHID  string             `json:"parentPHID"`
	Content     string             `json:"content"`
}
