package responses

import (
	"encoding/json"
	"errors"

	"github.com/uber/gonduit/entities"
	"github.com/uber/gonduit/util"
)

// ManiphestQueryResponse is the response of calling maniphest.query.
type ManiphestQueryResponse map[string]*entities.ManiphestTask

// Get gets the task with the speicfied numeric ID.
func (res ManiphestQueryResponse) Get(key string) *entities.ManiphestTask {
	if _, ok := res[key]; ok {
		return res[key]
	}

	return nil
}

// ManiphestGetTaskTransactionsResponse is the response of calling maniphest.query.
type ManiphestGetTaskTransactionsResponse map[string][]*entities.ManiphestTaskTranscation

// ManiphestSearchResponse contains fields that are in server response to maniphest.search.
type ManiphestSearchResponse struct {
	// Data contains search results.
	Data []*ManiphestSearchResponseItem `json:"data"`
	// Curson contains paging data.
	Cursor struct {
		Limit  uint64 `json:"limit"`
		After  string `json:"after"`
		Before string `json:"before"`
	} `json:"cursor,omitempty"`
}

// ManiphestSearchAttachmentColumnBoardsColumn descrbied a column in "columns" attachment.
type ManiphestSearchAttachmentColumnBoardsColumn struct {
	// ID is column identifier.
	ID int `json:"id"`
	// PHID is column PHID.
	PHID string `json:"phid"`
	// Name is column name.
	Name string `json:"name"`
}

// ManiphestSearchAttachmentColumnBoardsColumns is a wrapper for a slice of columns.
type ManiphestSearchAttachmentColumnBoardsColumns struct {
	// Columns is collection of columns in an attachment.
	Columns []*ManiphestSearchAttachmentColumnBoardsColumn `json:"columns,omitempty"`
}

// ManiphestSearchAttachmentColumnBoards is a wrapper type for columns because sometimes columns is a map and sometimes it is an array.
type ManiphestSearchAttachmentColumnBoards struct {
	// ColumnMap is dictionary of columns.
	ColumnMap map[string]*ManiphestSearchAttachmentColumnBoardsColumns
	// Columns is collection of columns.
	Columns []*ManiphestSearchAttachmentColumnBoardsColumn
}

// UnmarshalJSON parses column data from server response.
func (b *ManiphestSearchAttachmentColumnBoards) UnmarshalJSON(data []byte) error {
	if b == nil {
		return errors.New("boards is nil")
	}
	jerr := json.Unmarshal(data, &b.ColumnMap)
	if jerr == nil {
		return nil
	}
	return json.Unmarshal(data, &b.Columns)
}

// TaskDescription contains task description data.
type TaskDescription struct {
	// Raw is raw task description.
	Raw string `json:"raw"`
}

// ManiphestSearchResponseItem contains information about a particular search result.
type ManiphestSearchResponseItem struct {
	// ID is task identifier.
	ID int `json:"id"`
	// Type is task type.
	Type string `json:"type"`
	// PHID is PHID of the task.
	PHID string `json:"phid"`
	// Fields contains task data.
	Fields struct {
		// Name is task name.
		Name string `json:"name"`
		// Description is detailed task description.
		Description *TaskDescription `json:"description"`
		// AuthorPHID is PHID of task submitter.
		AuthorPHID string `json:"authorPHID"`
		// OwnerPHID is PHID of the person who currently assigned to task.
		OwnerPHID string `json:"ownerPHID"`
		// Status is task status.
		Status ManiphestSearchResultStatus `json:"status"`
		// Priority is task priority.
		Priority ManiphestSearchResultPriority `json:"priority"`
		// Points is point value of the task.
		Points json.Number `json:"points"`
		// Subtype of the task.
		Subtype string `json:"subtype"`
		// CloserPHID is user who closed the task, if the task is closed.
		CloserPHID string `json:"closerPHID"`
		// SpacePHID is PHID of the policy space this object is part of.
		SpacePHID string `json:"spacePHID"`
		// Date created is epoch timestamp when the object was created.
		DateCreated util.UnixTimestamp `json:"dateCreated"`
		// DateModified is epoch timestamp when the object was last updated.
		DateModified util.UnixTimestamp `json:"dateModified"`
		// Policy is map of capabilities to current policies.
		Policy SearchResultPolicy `json:"policy"`
		// CustomTaskType is custom task type.
		CustomTaskType string `json:"custom.task_type"`
		// CustomSeverity is task severity custom value.
		CustomSeverity string `json:"custom.severity"`
	} `json:"fields"`
	Attachments struct {
		// Columns contains columnt data if requested.
		Columns struct {
			// Boards is ???.
			Boards *ManiphestSearchAttachmentColumnBoards `json:"boards"`
		} `json:"columns"`
		// Subscribers contains subscribers attachment data.
		Subscribers struct {
			// SubscriberPHIDs is a collection of PHIDs of persons subscribed to a task.
			SubscriberPHIDs []string `json:"subscriberPHIDs"`
			// SubscriberCount is number of subscribers.
			SubscriberCount int `json:"subscriberCount"`
			// ViewerIsSubscribed specifies if request is subscribed to this task.
			ViewerIsSubscribed bool `json:"viewerIsSubscribed"`
		} `json:"subscribers"`
		// Projects contains project attachment data.
		Projects struct {
			// ProjectPHIDs is collection of PHIDs of projects that this task is tagged with.
			ProjectPHIDs []string `json:"projectPHIDs"`
		} `json:"projects"`
	} `json:"attachments"`
}

// ManiphestSearchResultStatus represents a maniphest status as returned by maniphest.search.
type ManiphestSearchResultStatus struct {
	// Value is status value.
	Value string `json:"value"`
	// Name is status name.
	Name string `json:"name"`
	// Color is ???.
	Color string `json:"color"`
}

// ManiphestSearchResultPriority represents a priority for a maniphest item in a search result.
type ManiphestSearchResultPriority struct {
	// Value is priority value.
	Value int `json:"value"`
	// Subpriority is task subpriority value.
	Subpriority float64 `json:"subpriority"`
	// Name is priority name.
	Name string `json:"name"`
	// Color is ???.
	Color string `json:"color"`
}

// SearchResultPolicy reflects the permission policy on a maniphest item in a search result.
type SearchResultPolicy struct {
	// View is ???.
	View string `json:"view"`
	// Interact is ???.
	Interact string `json:"interact"`
	// Edit is ???.
	Edit string `json:"edit"`
}

// ManiphestSearchResultColumn represents what workboard columns an item may be a member of.
type ManiphestSearchResultColumn struct {
	// ID is column ID.
	ID int
	// PHID is column PHID.
	PHID string
	// Name is column name.
	Name string
	// ProjectPHID is PHID of project where column is defined.
	ProjectPHID string
}
