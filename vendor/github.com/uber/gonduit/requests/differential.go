package requests

import (
	"github.com/uber/gonduit/constants"
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// DifferentialGetCommitMessageRequest represents a request to the
// differential.getcommitmessage call.
type DifferentialGetCommitMessageRequest struct {
	RevisionID uint64                                         `json:"revision_id"`
	Fields     []string                                       `json:"fields"`
	Edit       constants.DifferentialGetCommitMessageEditType `json:"edit"`
	Request
}

// DifferentialQueryRequest represents a request to the
// differential.query call.
type DifferentialQueryRequest struct {
	Authors          []string                           `json:"authors"`
	CCs              []string                           `json:"ccs"`
	Reviewers        []string                           `json:"reviewers"`
	Paths            [][]string                         `json:"paths"`
	CommitHashes     [][]string                         `json:"commitHashes"`
	Status           constants.DifferentialStatusLegacy `json:"status"`
	Order            constants.DifferentialQueryOrder   `json:"order"`
	Limit            uint64                             `json:"limit"`
	Offset           uint64                             `json:"offset"`
	IDs              []uint64                           `json:"ids"`
	PHIDs            []string                           `json:"phids"`
	Subscribers      []string                           `json:"subscribers"`
	ResponsibleUsers []string                           `json:"responsibleUsers"`
	Branches         []string                           `json:"branches"`
	Request
}

// DifferentialQueryDiffsRequest represents a request
// to the differential.querydiffs call.
type DifferentialQueryDiffsRequest struct {
	IDs         []uint64 `json:"ids"`
	RevisionIDs []uint64 `json:"revisionIDs"`
	Request
}

// DifferentialGetCommitPathsRequest represents a request to the
// differential.getcommitpaths call.
type DifferentialGetCommitPathsRequest struct {
	RevisionID uint64 `json:"revision_id"`
	Request
}

// DifferentialRevisionSearchRequest represents a request to
// differential.revision.search API method.
type DifferentialRevisionSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *DifferentialRevisionSearchConstraints `json:"constraints,omitempty"`
	// Attachments specified what additional data should be returned with each
	// result.
	Attachments *DifferentialRevisionSearchAttachments `json:"attachments,omitempty"`

	*entities.Cursor
	Request
}

// DifferentialRevisionSearchAttachments contains fields that specify what
// additional data should be returned with search results.
type DifferentialRevisionSearchAttachments struct {
	// Reviewers requests to get the reviewers for each revision.
	Reviewers bool `json:"reviewers,omitempty"`
	// Subscribers if true instructs server to return subscribers list for each task.
	Subscribers bool `json:"subscribers,omitempty"`
	// Projects requests to get information about projects.
	Projects bool `json:"projects,omitempty"`
}

// DifferentialRevisionSearchConstraints describes search criteria for request.
type DifferentialRevisionSearchConstraints struct {
	IDs              []int               `json:"ids,omitempty"`
	PHIDs            []string            `json:"phids,omitempty"`
	ResponsiblePHIDs []string            `json:"responsiblePHIDs,omitempty"`
	AuthorPHIDs      []string            `json:"authorPHIDs,omitempty"`
	ReviewerPHIDs    []string            `json:"reviewerPHIDs,omitempty"`
	RepositoryPHIDs  []string            `json:"repositoryPHIDs,omitempty"`
	Statuses         []string            `json:"statuses,omitempty"`
	CreatedStart     *util.UnixTimestamp `json:"createdStart,omitempty"`
	CreatedEnd       *util.UnixTimestamp `json:"createdEnd,omitempty"`
	ModifiedStart    *util.UnixTimestamp `json:"modifiedStart,omitempty"`
	ModifiedEnd      *util.UnixTimestamp `json:"modifiedEnd,omitempty"`
	Query            string              `json:"query,omitempty"`
	Subscribers      []string            `json:"subscribers,omitempty"`
	Projects         []string            `json:"projects,omitempty"`
}

// DifferentialDiffSearchRequest represents a request to
// differential.diff.search API method.
type DifferentialDiffSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *DifferentialDiffSearchConstraints `json:"constraints,omitempty"`
	// Attachments specified what additional data should be returned with each
	// result.
	Attachments *DifferentialDiffSearchAttachments `json:"attachments,omitempty"`

	*entities.Cursor
	Request
}

// DifferentialDiffSearchAttachments contains fields that specify what
// additional data should be returned with search results.
type DifferentialDiffSearchAttachments struct {
	// Get the local commits (if any) for each diff.
	Commits bool `json:"commits,omitempty"`
}

// DifferentialDiffSearchConstraints describes search criteria for request.
type DifferentialDiffSearchConstraints struct {
	IDs           []int    `json:"ids,omitempty"`
	PHIDs         []string `json:"phids,omitempty"`
	RevisionPHIDs []string `json:"revisionPHIDs,omitempty"`
}
