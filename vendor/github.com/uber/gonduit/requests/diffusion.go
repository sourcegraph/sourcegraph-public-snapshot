package requests

import (
	"github.com/uber/gonduit/entities"
)

// DiffusionQueryCommitsRequest represents a request to the
// diffusion.querycommits call.
type DiffusionQueryCommitsRequest struct {
	IDs            []uint64 `json:"ids"`
	PHIDs          []string `json:"phids"`
	Names          []string `json:"names"`
	RepositoryPHID string   `json:"repositoryPHID"`
	NeedMessages   bool     `json:"needMessages"`
	BypassCache    bool     `json:"bypassCache"`
	Before         string   `json:"before"`
	After          string   `json:"after"`
	Limit          uint64   `json:"limit"`
	Request
}

// DiffusionRepositorySearchRequest represents a request to
// diffusion.repository.search API method.
type DiffusionRepositorySearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *DiffusionRepositorySearchConstraints `json:"constraints,omitempty"`
	// Attachments specified what additional data should be returned with each
	// result.
	Attachments *DiffusionRepositorySearchAttachments `json:"attachments,omitempty"`

	*entities.Cursor
	Request
}

// DiffusionRepositorySearchConstraints describes search criteria for request.
type DiffusionRepositorySearchConstraints struct {
	IDs        []int    `json:"ids,omitempty"`
	PHIDs      []string `json:"phids,omitempty"`
	Callsigns  []string `json:"callsigns,omitempty"`
	ShortNames []string `json:"shortnames,omitempty"`
	Status     string   `json:"status,omitempty"`
	Hosted     string   `json:"hosted,omitempty"`
	Types      []string `json:"types,omitempty"`
	URIs       []string `json:"uris,omitempty"`
	Projects   []string `json:"projects,omitempty"`
	Owners     []string `json:"owners,omitempty"`
	Spaces     []string `json:"spaces,omitempty"`
}

// DiffusionRepositorySearchAttachments contains fields that specify what
// additional data should be returned with search results.
type DiffusionRepositorySearchAttachments struct {
	// URIs returns a list of associated URIs for each repository.
	URIs bool `json:"uris,omitempty"`
	// Metrics returns commit count, most recent commit and other metrics
	// for each repository.
	Metrics bool `json:"metrics,omitempty"`
	// Projects requests to get information about projects.
	Projects bool `json:"projects,omitempty"`
}
