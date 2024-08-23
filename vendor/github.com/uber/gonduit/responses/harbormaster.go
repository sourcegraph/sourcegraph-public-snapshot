package responses

import (
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// HarbormasterBuildableSearchResponse contains fields that are in server
// response to differential.revision.search.
type HarbormasterBuildableSearchResponse struct {
	// Data contains search results.
	Data []*HarbormasterBuildableSearchResponseItem `json:"data"`

	// Curson contains paging data.
	Cursor SearchCursor `json:"cursor,omitempty"`
}

// HarbormasterBuildableSearchResponseItem contains information about a
// particular search result.
type HarbormasterBuildableSearchResponseItem struct {
	ResponseObject
	Fields HarbormasterBuildableSearchResponseItemFields `json:"fields"`
	SearchCursor
}

// HarbormasterBuildableSearchResponseItemFields is a collection of object
// fields.
type HarbormasterBuildableSearchResponseItemFields struct {
	ObjectPHID      string             `json:"ObjectPHID"`
	ContainerPHID   string             `json:"ContainerPHID"`
	BuildableStatus BuildableStatus    `json:"buildableStatus"`
	IsManual        bool               `json:"isManual"`
	URI             string             `json:"uri"`
	DateCreated     util.UnixTimestamp `json:"dateCreated"`
	DateModified    util.UnixTimestamp `json:"dateModified"`
}

// BuildableStatus is a container of status value.
type BuildableStatus struct {
	Value entities.BuildableStatus `json:"value"`
}

// HarbormasterBuildSearchResponse contains fields that are in server
// response to differential.revision.search.
type HarbormasterBuildSearchResponse struct {
	// Data contains search results.
	Data []*HarbormasterBuildSearchResponseItem `json:"data"`

	// Curson contains paging data.
	Cursor SearchCursor `json:"cursor,omitempty"`
}

// HarbormasterBuildSearchResponseItem contains information about a
// particular search result.
type HarbormasterBuildSearchResponseItem struct {
	ResponseObject
	Fields HarbormasterBuildSearchResponseItemFields `json:"fields"`
	SearchCursor
}

// HarbormasterBuildSearchResponseItemFields is a collection of object
// fields.
type HarbormasterBuildSearchResponseItemFields struct {
	BuildablePHID string             `json:"buildablePHID"`
	BuildPlanPHID string             `json:"buildPlanPHID"`
	BuildStatus   BuildStatus        `json:"buildStatus"`
	InitiatorPHID string             `json:"initiatorPHID"`
	Name          string             `json:"name"`
	DateCreated   util.UnixTimestamp `json:"dateCreated"`
	DateModified  util.UnixTimestamp `json:"dateModified"`
}

// BuildStatus is a container of status value.
type BuildStatus struct {
	Value entities.BuildStatus `json:"value"`
}
