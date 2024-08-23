package responses

import (
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// DifferentialQueryResponse is the response of calling differential.query.
type DifferentialQueryResponse []*entities.DifferentialRevision

// DifferentialQueryDiffsResponse is the response of calling differential.querydiffs.
type DifferentialQueryDiffsResponse []*entities.DifferentialDiff

// DifferentialGetCommitPathsResponse is the response of calling
// differential.getcommitpaths.
type DifferentialGetCommitPathsResponse []string

// DifferentialGetCommitMessageResponse is the response of calling
// differential.getcommitmessage.
type DifferentialGetCommitMessageResponse string

// DifferentialRevisionSearchResponse contains fields that are in server
// response to differential.revision.search.
type DifferentialRevisionSearchResponse struct {
	// Data contains search results.
	Data []*DifferentialRevisionSearchResponseItem `json:"data"`

	// Cursor contains paging data.
	Cursor SearchCursor `json:"cursor,omitempty"`
}

// DifferentialRevisionSearchResponseItem contains information about a
// particular search result.
type DifferentialRevisionSearchResponseItem struct {
	ResponseObject
	Fields      DifferentialRevisionSearchResponseItemFields `json:"fields"`
	Attachments DifferentialRevisionSearchAttachments        `json:"attachments"`
	SearchCursor
}

// DifferentialRevisionSearchResponseItemFields is a collection of object
// fields.
type DifferentialRevisionSearchResponseItemFields struct {
	Title          string                     `json:"title"`
	URI            string                     `json:"uri"`
	AuthorPHID     string                     `json:"authorPHID"`
	Status         DifferentialRevisionStatus `json:"status"`
	RepositoryPHID string                     `json:"repositoryPHID"`
	DiffPHID       string                     `json:"diffPHID"`
	Summary        string                     `json:"summary"`
	TestPlan       string                     `json:"testPlan"`
	IsDraft        bool                       `json:"isDraft"`
	HoldAsDraft    bool                       `json:"holdAsDraft"`
	DateCreated    util.UnixTimestamp         `json:"dateCreated"`
	DateModified   util.UnixTimestamp         `json:"dateModified"`
}

// DifferentialRevisionStatus represents item status returned by response.
type DifferentialRevisionStatus struct {
	Value  string `json:"value"`
	Name   string `json:"name"`
	Closed bool   `json:"closed"`
}

// DifferentialRevisionSearchAttachments holds possible attachments for the API
// method.
type DifferentialRevisionSearchAttachments struct {
	Reviewers   SearchAttachmentReviewers   `json:"reviewers"`
	Subscribers SearchAttachmentSubscribers `json:"subscribers"`
	Projects    SearchAttachmentProjects    `json:"projects"`
}

// DifferentialDiffSearchResponse contains fields that are in server
// response to differential.diff.search.
type DifferentialDiffSearchResponse struct {
	// Data contains search results.
	Data []*DifferentialDiffSearchResponseItem `json:"data"`

	// Cursor contains paging data.
	Cursor SearchCursor `json:"cursor,omitempty"`
}

// DifferentialDiffSearchResponseItem contains information about a
// particular search result.
type DifferentialDiffSearchResponseItem struct {
	ResponseObject
	Fields      DifferentialDiffSearchResponseItemFields `json:"fields"`
	Attachments DifferentialDiffSearchAttachments        `json:"attachments"`
	SearchCursor
}

// DifferentialDiffSearchResponseItemFields is a collection of object
// fields.
type DifferentialDiffSearchResponseItemFields struct {
	RevisionPHID   string                `json:"revisionPHID"`
	AuthorPHID     string                `json:"authorPHID"`
	RepositoryPHID string                `json:"repositoryPHID"`
	Refs           []DifferentialDiffRef `json:"refs"`
	DateCreated    util.UnixTimestamp    `json:"dateCreated"`
	DateModified   util.UnixTimestamp    `json:"dateModified"`
}

type DifferentialDiffRef struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type DifferentialDiffSearchAttachments struct {
	Commits SearchAttachmentCommits `json:"commits"`
}
