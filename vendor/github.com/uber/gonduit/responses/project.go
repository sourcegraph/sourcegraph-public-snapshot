package responses

import (
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// ProjectQueryResponse represents a response from calling project.query.
type ProjectQueryResponse struct {
	Data    map[string]entities.Project `json:"data"`
	SlugMap map[string]string           `json:"sligMap"`
	Cursor  entities.Cursor             `json:"cursor"`
}

// ProjectSearchResponse contains fields that are in server
// response to project.search.
type ProjectSearchResponse struct {
	// Data contains search results.
	Data []*ProjectSearchResponseItem `json:"data"`

	// Cursor contains paging data.
	Cursor SearchCursor `json:"cursor,omitempty"`
}

// ProjectSearchResponseItem contains information about a
// particular search result.
type ProjectSearchResponseItem struct {
	ResponseObject
	Fields      ProjectSearchResponseItemFields `json:"fields"`
	Attachments ProjectSearchAttachments        `json:"attachments"`
	SearchCursor
}

// ProjectSearchResponseItemFields is a collection of object
// fields.
type ProjectSearchResponseItemFields struct {
	Name         string             `json:"name"`
	Slug         string             `json:"slug"`
	Description  string             `json:"description"`
	Subtype      string             `json:"subtype"`
	Milestone    int                `json:"milestone"`
	Depth        int                `json:"depth"`
	Parent       *ProjectParent     `json:"parent"`
	Icon         ProjectIcon        `json:"icon"`
	Color        ProjectColor       `json:"color"`
	SpacePHID    string             `json:"spacePHID"`
	DateCreated  util.UnixTimestamp `json:"dateCreated"`
	DateModified util.UnixTimestamp `json:"dateModified"`
}

type ProjectIcon struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type ProjectColor struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type ProjectParent struct {
	ID   int    `json:"id"`
	PHID string `json:"phid"`
	Name string `json:"name"`
}

type ProjectSearchAttachments struct {
	Members   SearchAttachmentMembers   `json:"members"`
	Watchers  SearchAttachmentWatchers  `json:"watchers"`
	Ancestors SearchAttachmentAncestors `json:"ancestors"`
}
