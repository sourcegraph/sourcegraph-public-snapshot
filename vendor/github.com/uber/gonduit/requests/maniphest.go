package requests

import (
	"encoding/json"
	"errors"

	"github.com/uber/gonduit/constants"
	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// ManiphestQueryRequest represents a request to maniphest.query.
type ManiphestQueryRequest struct {
	IDs          []string                      `json:"ids"`
	PHIDs        []string                      `json:"phids"`
	OwnerPHIDs   []string                      `json:"ownerPHIDs"`
	AuthorPHIDs  []string                      `json:"authorPHIDs"`
	ProjectPHIDs []string                      `json:"projectPHIDs"`
	CCPHIDs      []string                      `json:"ccPHIDs"`
	FullText     string                        `json:"fullText"`
	Status       constants.ManiphestTaskStatus `json:"status"`
	Order        constants.ManiphestQueryOrder `json:"order"`
	Limit        uint64                        `json:"limit"`
	Offset       uint64                        `json:"offset"`
	Request
}

// ManiphestCreateTaskRequest represents a request to maniphest.createtask.
type ManiphestCreateTaskRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	OwnerPHID    string   `json:"ownerPHID"`
	ViewPolicy   string   `json:"viewPolicy"`
	EditPolicy   string   `json:"editPolicy"`
	CCPHIDs      []string `json:"ccPHIDs"`
	Priority     int      `json:"priority"`
	ProjectPHIDs []string `json:"projectPHIDs"`
	Request
}

// ManiphestGetTaskTransactions represents a request to maniphest.gettasktransactions.
type ManiphestGetTaskTransactions struct {
	IDs []string `json:"ids"`
	Request
}

// ManiphestSearchRequest represents a request to maniphest.search API method.
type ManiphestSearchRequest struct {
	// QueryKey is builtin or saved query to use. It is optional and sets initial constraints.
	QueryKey string `json:"queryKey,omitempty"`
	// Constraints contains additional filters for results. Applied on top of query if provided.
	Constraints *ManiphestSearchConstraints `json:"constraints,omitempty"`
	// Attachments specified what additional data should be returned with each result.
	Attachments *ManiphestSearchAttachments `json:"attachments,omitempty"`

	*entities.Cursor
	Request
}

// ManiphestSearchAttachments contains fields that specify what additional data should be returned with search results.
type ManiphestSearchAttachments struct {
	// Subscribers if true instructs server to return subscribers list for each task.
	Subscribers bool `json:"subscribers,omitempty"`
	// Columns requests to get the workboard columns where an object appears.
	Columns bool `json:"columns,omitempty"`
	// Projects requests to get information about projects.
	Projects bool `json:"projects,omitempty"`
}

// ManiphestRequestSearchOrder describers how results should be ordered.
type ManiphestRequestSearchOrder struct {
	// Builtin is the name of predefined order to use.
	Builtin string
	// Order is list of columns to use for sorting, e.g. ["color", "-name", "id"],
	Order []string
}

// UnmarshalJSON parses JSON  into an instance of ManiphestRequestSearchOrder type.
func (o *ManiphestRequestSearchOrder) UnmarshalJSON(data []byte) error {
	if o == nil {
		return errors.New("maniphest search order is nil")
	}
	if jerr := json.Unmarshal(data, &o.Builtin); jerr == nil {
		return nil
	}

	return json.Unmarshal(data, &o.Order)
}

// MarshalJSON creates JSON our of ManiphestRequestSearchOrder instance.
func (o *ManiphestRequestSearchOrder) MarshalJSON() ([]byte, error) {
	if o == nil {
		return nil, errors.New("maniphest search order is nil")
	}
	if o.Builtin != "" {
		return json.Marshal(o.Builtin)
	}
	if len(o.Order) > 0 {
		return json.Marshal(o.Order)
	}

	return nil, nil
}

// ManiphestSearchConstraints describes search criteria for request.
type ManiphestSearchConstraints struct {
	// IDs - search for objects with specific IDs.
	IDs []int `json:"ids,omitempty"`
	// PHIDs - search for objects with specific PHIDs.
	PHIDs []string `json:"phids,omitempty"`
	// AssignedTo - search for tasks owned by a user from a list.
	AssignedTo []string `json:"assigned,omitempty"`
	// Authors - search for tasks with given authors.
	Authors []string `json:"authorPHIDs,omitempty"`
	// Statuses - search for tasks with given statuses.
	Statuses []string `json:"statuses,omitempty"`
	// Priorities - search for tasks with given priorities.
	Priorities []int `json:"priorities,omitempty"`
	// Subtypes - search for tasks with given subtypes.
	Subtypes []string `json:"subtypes,omitempty"`
	// Column PHIDs ??? - no doc on phab site.
	ColumnPHIDs []string `json:"columnPHIDs,omitempty"`
	// OpenParents - search for tasks that have parents in open state.
	OpenParents *bool `json:"hasParents,omitempty"`
	// OpenSubtasks - search for tasks that have child tasks in open state.
	OpenSubtasks *bool `json:"hasSubtasks,omitempty"`
	// ParentIDs - search for children of these parents.
	ParentIDs []int `json:"parentIDs,omitempty"`
	// SubtaskIDs - Search for tasks that have these children.
	SubtaskIDs []int `json:"subtaskIDs,omitempty"`
	// CreatedAfter - search for tasks created after given date.
	CreatedAfter *util.UnixTimestamp `json:"createdStart,omitempty"`
	// CreatedBefore - search for tasks created before given date.
	CreatedBefore *util.UnixTimestamp `json:"createdEnd,omitempty"`
	// ModifiedAfter - search for tasks modified after given date.
	ModifiedAfter *util.UnixTimestamp `json:"modifiedStart,omitempty"`
	// ModifiedBefore - search for tasks modified before given date.
	ModifiedBefore *util.UnixTimestamp `json:"modifiedEnd,omitempty"`
	// ClosedAfter - search for tasks closed after given date.
	ClosedAfter *util.UnixTimestamp `json:"closedStart,omitempty"`
	// ClosedBefore - search for tasks closed before given date.
	ClosedBefore *util.UnixTimestamp `json:"closedEnd,omitempty"`
	// ClosedBy - search for tasks closed by people with given PHIDs.
	ClosedBy []string `json:"closerPHIDs,omitempty"`
	// Query - find objects matching a fulltext search query.
	Query string `json:"query,omitempty"`
	// Subscribers - search for objects with certain subscribers.
	Subscribers []string `json:"subscribers,omitempty"`
	// Projects - search for objects tagged with given projects.
	Projects []string `json:"projects,omitempty"`
	// Spaces - search for objects in certain spaces.
	Spaces []string `json:"spaces,omitempty"`
}
