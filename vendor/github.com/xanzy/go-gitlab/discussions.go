//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// DiscussionsService handles communication with the discussions related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/discussions.html
type DiscussionsService struct {
	client *Client
}

// Discussion represents a GitLab discussion.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/discussions.html
type Discussion struct {
	ID             string  `json:"id"`
	IndividualNote bool    `json:"individual_note"`
	Notes          []*Note `json:"notes"`
}

func (d Discussion) String() string {
	return Stringify(d)
}

// ListIssueDiscussionsOptions represents the available ListIssueDiscussions()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-issue-discussion-items
type ListIssueDiscussionsOptions ListOptions

// ListIssueDiscussions gets a list of all discussions for a single
// issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-issue-discussion-items
func (s *DiscussionsService) ListIssueDiscussions(pid interface{}, issue int, opt *ListIssueDiscussionsOptions, options ...RequestOptionFunc) ([]*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions", PathEscape(project), issue)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ds []*Discussion
	resp, err := s.client.Do(req, &ds)
	if err != nil {
		return nil, resp, err
	}

	return ds, resp, nil
}

// GetIssueDiscussion returns a single discussion for a specific project issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#get-single-issue-discussion-item
func (s *DiscussionsService) GetIssueDiscussion(pid interface{}, issue int, discussion string, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions/%s",
		PathEscape(project),
		issue,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// CreateIssueDiscussionOptions represents the available CreateIssueDiscussion()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-issue-thread
type CreateIssueDiscussionOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// CreateIssueDiscussion creates a new discussion to a single project issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-issue-thread
func (s *DiscussionsService) CreateIssueDiscussion(pid interface{}, issue int, opt *CreateIssueDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions", PathEscape(project), issue)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// AddIssueDiscussionNoteOptions represents the available AddIssueDiscussionNote()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-issue-thread
type AddIssueDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// AddIssueDiscussionNote creates a new discussion to a single project issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-issue-thread
func (s *DiscussionsService) AddIssueDiscussionNote(pid interface{}, issue int, discussion string, opt *AddIssueDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions/%s/notes",
		PathEscape(project),
		issue,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateIssueDiscussionNoteOptions represents the available
// UpdateIssueDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-issue-thread-note
type UpdateIssueDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// UpdateIssueDiscussionNote modifies existing discussion of an issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-issue-thread-note
func (s *DiscussionsService) UpdateIssueDiscussionNote(pid interface{}, issue int, discussion string, note int, opt *UpdateIssueDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions/%s/notes/%d",
		PathEscape(project),
		issue,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteIssueDiscussionNote deletes an existing discussion of an issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#delete-an-issue-thread-note
func (s *DiscussionsService) DeleteIssueDiscussionNote(pid interface{}, issue int, discussion string, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/discussions/%s/notes/%d",
		PathEscape(project),
		issue,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListSnippetDiscussionsOptions represents the available ListSnippetDiscussions()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-snippet-discussion-items
type ListSnippetDiscussionsOptions ListOptions

// ListSnippetDiscussions gets a list of all discussions for a single
// snippet. Snippet discussions are comments users can post to a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-snippet-discussion-items
func (s *DiscussionsService) ListSnippetDiscussions(pid interface{}, snippet int, opt *ListSnippetDiscussionsOptions, options ...RequestOptionFunc) ([]*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions", PathEscape(project), snippet)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ds []*Discussion
	resp, err := s.client.Do(req, &ds)
	if err != nil {
		return nil, resp, err
	}

	return ds, resp, nil
}

// GetSnippetDiscussion returns a single discussion for a given snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#get-single-snippet-discussion-item
func (s *DiscussionsService) GetSnippetDiscussion(pid interface{}, snippet int, discussion string, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions/%s",
		PathEscape(project),
		snippet,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// CreateSnippetDiscussionOptions represents the available
// CreateSnippetDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-snippet-thread
type CreateSnippetDiscussionOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// CreateSnippetDiscussion creates a new discussion for a single snippet.
// Snippet discussions are comments users can post to a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-snippet-thread
func (s *DiscussionsService) CreateSnippetDiscussion(pid interface{}, snippet int, opt *CreateSnippetDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions", PathEscape(project), snippet)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// AddSnippetDiscussionNoteOptions represents the available
// AddSnippetDiscussionNote() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-snippet-thread
type AddSnippetDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// AddSnippetDiscussionNote creates a new discussion to a single project
// snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-snippet-thread
func (s *DiscussionsService) AddSnippetDiscussionNote(pid interface{}, snippet int, discussion string, opt *AddSnippetDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions/%s/notes",
		PathEscape(project),
		snippet,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateSnippetDiscussionNoteOptions represents the available
// UpdateSnippetDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-snippet-thread-note
type UpdateSnippetDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// UpdateSnippetDiscussionNote modifies existing discussion of a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-snippet-thread-note
func (s *DiscussionsService) UpdateSnippetDiscussionNote(pid interface{}, snippet int, discussion string, note int, opt *UpdateSnippetDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions/%s/notes/%d",
		PathEscape(project),
		snippet,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteSnippetDiscussionNote deletes an existing discussion of a snippet.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#delete-a-snippet-thread-note
func (s *DiscussionsService) DeleteSnippetDiscussionNote(pid interface{}, snippet int, discussion string, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/snippets/%d/discussions/%s/notes/%d",
		PathEscape(project),
		snippet,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListGroupEpicDiscussionsOptions represents the available
// ListEpicDiscussions() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-group-epic-discussion-items
type ListGroupEpicDiscussionsOptions ListOptions

// ListGroupEpicDiscussions gets a list of all discussions for a single
// epic. Epic discussions are comments users can post to a epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-group-epic-discussion-items
func (s *DiscussionsService) ListGroupEpicDiscussions(gid interface{}, epic int, opt *ListGroupEpicDiscussionsOptions, options ...RequestOptionFunc) ([]*Discussion, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions",
		PathEscape(group),
		epic,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ds []*Discussion
	resp, err := s.client.Do(req, &ds)
	if err != nil {
		return nil, resp, err
	}

	return ds, resp, nil
}

// GetEpicDiscussion returns a single discussion for a given epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#get-single-epic-discussion-item
func (s *DiscussionsService) GetEpicDiscussion(gid interface{}, epic int, discussion string, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions/%s",
		PathEscape(group),
		epic,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// CreateEpicDiscussionOptions represents the available CreateEpicDiscussion()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-epic-thread
type CreateEpicDiscussionOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// CreateEpicDiscussion creates a new discussion for a single epic. Epic
// discussions are comments users can post to a epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-epic-thread
func (s *DiscussionsService) CreateEpicDiscussion(gid interface{}, epic int, opt *CreateEpicDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions",
		PathEscape(group),
		epic,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// AddEpicDiscussionNoteOptions represents the available
// AddEpicDiscussionNote() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-epic-thread
type AddEpicDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// AddEpicDiscussionNote creates a new discussion to a single project epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-epic-thread
func (s *DiscussionsService) AddEpicDiscussionNote(gid interface{}, epic int, discussion string, opt *AddEpicDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions/%s/notes",
		PathEscape(group),
		epic,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateEpicDiscussionNoteOptions represents the available UpdateEpicDiscussion()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-epic-thread-note
type UpdateEpicDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// UpdateEpicDiscussionNote modifies existing discussion of a epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-existing-epic-thread-note
func (s *DiscussionsService) UpdateEpicDiscussionNote(gid interface{}, epic int, discussion string, note int, opt *UpdateEpicDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions/%s/notes/%d",
		PathEscape(group),
		epic,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteEpicDiscussionNote deletes an existing discussion of a epic.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#delete-an-epic-thread-note
func (s *DiscussionsService) DeleteEpicDiscussionNote(gid interface{}, epic int, discussion string, note int, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/epics/%d/discussions/%s/notes/%d",
		PathEscape(group),
		epic,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListMergeRequestDiscussionsOptions represents the available
// ListMergeRequestDiscussions() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-merge-request-discussion-items
type ListMergeRequestDiscussionsOptions ListOptions

// ListMergeRequestDiscussions gets a list of all discussions for a single
// merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-merge-request-discussion-items
func (s *DiscussionsService) ListMergeRequestDiscussions(pid interface{}, mergeRequest int, opt *ListMergeRequestDiscussionsOptions, options ...RequestOptionFunc) ([]*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions",
		PathEscape(project),
		mergeRequest,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ds []*Discussion
	resp, err := s.client.Do(req, &ds)
	if err != nil {
		return nil, resp, err
	}

	return ds, resp, nil
}

// GetMergeRequestDiscussion returns a single discussion for a given merge
// request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#get-single-merge-request-discussion-item
func (s *DiscussionsService) GetMergeRequestDiscussion(pid interface{}, mergeRequest int, discussion string, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions/%s",
		PathEscape(project),
		mergeRequest,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// CreateMergeRequestDiscussionOptions represents the available
// CreateMergeRequestDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-merge-request-thread
type CreateMergeRequestDiscussionOptions struct {
	Body      *string       `url:"body,omitempty" json:"body,omitempty"`
	CommitID  *string       `url:"commit_id,omitempty" json:"commit_id,omitempty"`
	CreatedAt *time.Time    `url:"created_at,omitempty" json:"created_at,omitempty"`
	Position  *NotePosition `url:"position,omitempty" json:"position,omitempty"`
}

// CreateMergeRequestDiscussion creates a new discussion for a single merge
// request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-merge-request-thread
func (s *DiscussionsService) CreateMergeRequestDiscussion(pid interface{}, mergeRequest int, opt *CreateMergeRequestDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions",
		PathEscape(project),
		mergeRequest,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// ResolveMergeRequestDiscussionOptions represents the available
// ResolveMergeRequestDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#resolve-a-merge-request-thread
type ResolveMergeRequestDiscussionOptions struct {
	Resolved *bool `url:"resolved,omitempty" json:"resolved,omitempty"`
}

// ResolveMergeRequestDiscussion resolves/unresolves whole discussion of a merge
// request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#resolve-a-merge-request-thread
func (s *DiscussionsService) ResolveMergeRequestDiscussion(pid interface{}, mergeRequest int, discussion string, opt *ResolveMergeRequestDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions/%s",
		PathEscape(project),
		mergeRequest,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// AddMergeRequestDiscussionNoteOptions represents the available
// AddMergeRequestDiscussionNote() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-merge-request-thread
type AddMergeRequestDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// AddMergeRequestDiscussionNote creates a new discussion to a single project
// merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-merge-request-thread
func (s *DiscussionsService) AddMergeRequestDiscussionNote(pid interface{}, mergeRequest int, discussion string, opt *AddMergeRequestDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions/%s/notes",
		PathEscape(project),
		mergeRequest,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateMergeRequestDiscussionNoteOptions represents the available
// UpdateMergeRequestDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-an-existing-merge-request-thread-note
type UpdateMergeRequestDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
	Resolved  *bool      `url:"resolved,omitempty" json:"resolved,omitempty"`
}

// UpdateMergeRequestDiscussionNote modifies existing discussion of a merge
// request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-an-existing-merge-request-thread-note
func (s *DiscussionsService) UpdateMergeRequestDiscussionNote(pid interface{}, mergeRequest int, discussion string, note int, opt *UpdateMergeRequestDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions/%s/notes/%d",
		PathEscape(project),
		mergeRequest,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteMergeRequestDiscussionNote deletes an existing discussion of a merge
// request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#delete-a-merge-request-thread-note
func (s *DiscussionsService) DeleteMergeRequestDiscussionNote(pid interface{}, mergeRequest int, discussion string, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/discussions/%s/notes/%d",
		PathEscape(project),
		mergeRequest,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListCommitDiscussionsOptions represents the available
// ListCommitDiscussions() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-commit-discussion-items
type ListCommitDiscussionsOptions ListOptions

// ListCommitDiscussions gets a list of all discussions for a single
// commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#list-project-commit-discussion-items
func (s *DiscussionsService) ListCommitDiscussions(pid interface{}, commit string, opt *ListCommitDiscussionsOptions, options ...RequestOptionFunc) ([]*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions",
		PathEscape(project),
		commit,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ds []*Discussion
	resp, err := s.client.Do(req, &ds)
	if err != nil {
		return nil, resp, err
	}

	return ds, resp, nil
}

// GetCommitDiscussion returns a single discussion for a specific project
// commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#get-single-commit-discussion-item
func (s *DiscussionsService) GetCommitDiscussion(pid interface{}, commit string, discussion string, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions/%s",
		PathEscape(project),
		commit,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// CreateCommitDiscussionOptions represents the available
// CreateCommitDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-commit-thread
type CreateCommitDiscussionOptions struct {
	Body      *string       `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time    `url:"created_at,omitempty" json:"created_at,omitempty"`
	Position  *NotePosition `url:"position,omitempty" json:"position,omitempty"`
}

// CreateCommitDiscussion creates a new discussion to a single project commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#create-new-commit-thread
func (s *DiscussionsService) CreateCommitDiscussion(pid interface{}, commit string, opt *CreateCommitDiscussionOptions, options ...RequestOptionFunc) (*Discussion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions",
		PathEscape(project),
		commit,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	d := new(Discussion)
	resp, err := s.client.Do(req, d)
	if err != nil {
		return nil, resp, err
	}

	return d, resp, nil
}

// AddCommitDiscussionNoteOptions represents the available
// AddCommitDiscussionNote() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-commit-thread
type AddCommitDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// AddCommitDiscussionNote creates a new discussion to a single project commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#add-note-to-existing-commit-thread
func (s *DiscussionsService) AddCommitDiscussionNote(pid interface{}, commit string, discussion string, opt *AddCommitDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions/%s/notes",
		PathEscape(project),
		commit,
		discussion,
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// UpdateCommitDiscussionNoteOptions represents the available
// UpdateCommitDiscussion() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-an-existing-commit-thread-note
type UpdateCommitDiscussionNoteOptions struct {
	Body      *string    `url:"body,omitempty" json:"body,omitempty"`
	CreatedAt *time.Time `url:"created_at,omitempty" json:"created_at,omitempty"`
}

// UpdateCommitDiscussionNote modifies existing discussion of an commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#modify-an-existing-commit-thread-note
func (s *DiscussionsService) UpdateCommitDiscussionNote(pid interface{}, commit string, discussion string, note int, opt *UpdateCommitDiscussionNoteOptions, options ...RequestOptionFunc) (*Note, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions/%s/notes/%d",
		PathEscape(project),
		commit,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	n := new(Note)
	resp, err := s.client.Do(req, n)
	if err != nil {
		return nil, resp, err
	}

	return n, resp, nil
}

// DeleteCommitDiscussionNote deletes an existing discussion of an commit.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/discussions.html#delete-a-commit-thread-note
func (s *DiscussionsService) DeleteCommitDiscussionNote(pid interface{}, commit string, discussion string, note int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/repository/commits/%s/discussions/%s/notes/%d",
		PathEscape(project),
		commit,
		discussion,
		note,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
