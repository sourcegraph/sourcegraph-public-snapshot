package requests

import (
	"github.com/uber/gonduit/constants"
	"github.com/uber/gonduit/entities"
)

// ProjectQueryRequest represents a request to project.query.
type ProjectQueryRequest struct {
	IDs     []string                `json:"ids"`
	Names   []string                `json:"names"`
	PHIDs   []string                `json:"phids"`
	Slugs   []string                `json:"slugs"`
	Icons   []string                `json:"icons"`
	Colors  []string                `json:"colors"`
	Status  constants.ProjectStatus `json:"status"`
	Members []string                `json:"members"`
	Limit   uint64                  `json:"limit"`
	Offset  uint64                  `json:"offset"`
	Request
}

// ProjectSearchRequest represents a request to
// project.search API method.
type ProjectSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets
	// initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of
	// query if provided.
	Constraints *ProjectSearchConstraints `json:"constraints,omitempty"`
	// Attachments specified what additional data should be returned with each
	// result.
	Attachments *ProjectSearchAttachments `json:"attachments,omitempty"`

	*entities.Cursor
	Request
}

// ProjectSearchAttachments contains fields that specify what
// additional data should be returned with search results.
type ProjectSearchAttachments struct {
	// Members requests to get the member list for the project.
	Members bool `json:"members,omitempty"`
	// Watchers requests to get the watcher list for the project.
	Watchers bool `json:"watchers,omitempty"`
	// Ancestors requests to get the full ancestor list for each project.
	Ancestors bool `json:"ancestors,omitempty"`
}

// ProjectSearchConstraints describes search criteria for request.
type ProjectSearchConstraints struct {
	IDs   []int    `json:"ids,omitempty"`
	PHIDs []string `json:"phids,omitempty"`
	Name  string   `json:"name,omitempty"`
	Slugs []string `json:"slugs,omitempty"`
	// use phids
	Members []string `json:"members,omitempty"`
	// use phids
	Watchers    []string `json:"watchers,omitempty"`
	IsMilestone bool     `json:"isMilestone,omitempty"`
	IsRoot      bool     `json:"isRoot,omitempty"`
	MinDepth    int      `json:"minDepth,omitempty"`
	MaxDepth    int      `json:"maxDepth,omitempty"`
	Subtypes    []string `json:"subtypes,omitempty"`
	Icons       []string `json:"icons,omitempty"`
	Colors      []string `json:"colors,omitempty"`
	Parents     []string `json:"parents,omitempty"`
	Ancestors   []string `json:"ancestors,omitempty"`
	Query       string   `json:"query,omitempty"`
	Spaces      []string `json:"spaces,omitempty"`
}
