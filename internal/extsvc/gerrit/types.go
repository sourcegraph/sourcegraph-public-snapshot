package gerrit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	ChangeStatusNew       ChangeStatus = "NEW"
	ChangeStatusAbandoned ChangeStatus = "ABANDONED"
	ChangeStatusMerged    ChangeStatus = "MERGED"
)

type ChangeStatus string

// ListProjectsArgs defines options to be set on ListProjects method calls.
type ListProjectsArgs struct {
	Cursor *Pagination
	// If true, only fetches repositories with type CODE
	OnlyCodeProjects bool
}

// ListProjectsResponse defines a response struct returned from ListProjects method calls.
type ListProjectsResponse map[string]*Project

type Change struct {
	ID             string       `json:"id"`
	Project        string       `json:"project"`
	Branch         string       `json:"branch"`
	ChangeID       string       `json:"change_id"`
	Topic          string       `json:"topic"`
	Subject        string       `json:"subject"`
	Status         ChangeStatus `json:"status"`
	Created        time.Time    `json:"-"`
	Updated        time.Time    `json:"-"`
	Reviewed       bool         `json:"reviewed"`
	WorkInProgress bool         `json:"work_in_progress"`
	Hashtags       []string     `json:"hashtags"`
	ChangeNumber   int          `json:"_number"`
	Owner          struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"owner"`
}

func (c *Change) UnmarshalJSON(data []byte) error {
	type Alias Change
	aux := &struct {
		Created string `json:"created"`
		Updated string `json:"updated"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var created, updated time.Time

	createdParsed, err := time.Parse("2006-01-02 15:04:05.000000000", aux.Created)
	if err == nil {
		created = createdParsed
	}
	c.Created = created

	updatedParsed, err := time.Parse("2006-01-02 15:04:05.000000000", aux.Updated)
	if err == nil {
		updated = updatedParsed
	}
	c.Updated = updated

	return nil
}

func (c *Change) MarshalJSON() ([]byte, error) {
	type Alias Change
	return json.Marshal(&struct {
		Created string `json:"created"`
		Updated string `json:"updated"`
		*Alias
	}{
		Created: c.Created.Format("2006-01-02 15:04:05.000000000"),
		Updated: c.Updated.Format("2006-01-02 15:04:05.000000000"),
		Alias:   (*Alias)(c),
	})
}

type ChangeReviewComment struct {
	Message       string            `json:"message"`
	Tag           string            `json:"tag,omitempty"`
	Labels        map[string]int    `json:"labels,omitempty"`
	Notify        string            `json:"notify,omitempty"`
	NotifyDetails *NotifyDetails    `json:"notify_details,omitempty"`
	OnBehalfOf    string            `json:"on_behalf_of,omitempty"`
	Comments      map[string]string `json:"comments,omitempty"`
}

// CodeReviewKey
// Score represents the status of a review on Gerrit. Here are possible values for Vote:
//
//	+2 : approved, can be merged
//	+1 : approved, but needs additional reviews
//	 0 : no score
//	-1 : needs changes
//	-2 : rejected
const CodeReviewKey = "Code-Review"

type Reviewer struct {
	Approvals map[string]string `json:"approvals"`
	AccountID int               `json:"_account_id"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Username  string            `json:"username,omitempty"`
}

type NotifyDetails struct {
	EmailOnly bool `json:"email_only,omitempty"`
}

type Account struct {
	ID          int32  `json:"_account_id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type Group struct {
	ID          string `json:"id"`
	GroupID     int32  `json:"group_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedOn   string `json:"created_on"`
	Owner       string `json:"owner"`
	OwnerID     string `json:"owner_id"`
}

type Project struct {
	Description string            `json:"description"`
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Parent      string            `json:"parent"`
	State       string            `json:"state"`
	Branches    map[string]string `json:"branches"`
	Labels      map[string]Label  `json:"labels"`
}

type Label struct {
	Values       map[string]string `json:"values"`
	DefaultValue string            `json:"default_value"`
}

type MoveChangePayload struct {
	DestinationBranch string `json:"destination_branch"`
}

type SetCommitMessagePayload struct {
	Message string `json:"message"`
}

type Pagination struct {
	PerPage int
	// Either Skip or Page should be set. If Skip is non-zero, it takes precedence.
	Page int
	Skip int
}

// MultipleChangesError is returned by GetChange in
// the fringe situation that multiple
type MultipleChangesError struct {
	ID string
}

func (e MultipleChangesError) Error() string {
	return fmt.Sprintf("Multiple changes found with ID %s not found", e.ID)
}

type httpError struct {
	StatusCode int
	URL        *url.URL
	Body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("Gerrit API HTTP error: code=%d url=%q body=%q", e.StatusCode, e.URL, e.Body)
}

func (e *httpError) Unauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

func (e *httpError) NotFound() bool {
	return e.StatusCode == http.StatusNotFound
}
