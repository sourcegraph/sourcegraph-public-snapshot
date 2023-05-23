package gerrit

import (
	"fmt"
	"net/http"
	"net/url"
)

// ListProjectsArgs defines options to be set on ListProjects method calls.
type ListProjectsArgs struct {
	Cursor *Pagination
	// If true, only fetches repositories with type CODE
	OnlyCodeProjects bool
}

// ListProjectsResponse defines a response struct returned from ListProjects method calls.
type ListProjectsResponse map[string]*Project

type Change struct {
	ID       string `json:"id"`
	Project  string `json:"project"`
	Branch   string `json:"branch"`
	ChangeID string `json:"change_id"`
	Topic    string `json:"topic"`
	Subject  string `json:"subject"`
	Status   string `json:"status"`
	Created  string `json:"created"`
	Updated  string `json:"updated"`
	Owner    struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"owner"`
	// Add more fields as needed
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

type Pagination struct {
	PerPage int
	// Either Skip or Page should be set. If Skip is non-zero, it takes precedence.
	Page int
	Skip int
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
