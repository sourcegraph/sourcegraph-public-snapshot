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

// MergeRequestsService handles communication with the merge requests related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/merge_requests.html
type MergeRequestsService struct {
	client    *Client
	timeStats *timeStatsService
}

// MergeRequest represents a GitLab merge request.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/merge_requests.html
type MergeRequest struct {
	ID                        int              `json:"id"`
	IID                       int              `json:"iid"`
	TargetBranch              string           `json:"target_branch"`
	SourceBranch              string           `json:"source_branch"`
	ProjectID                 int              `json:"project_id"`
	Title                     string           `json:"title"`
	State                     string           `json:"state"`
	CreatedAt                 *time.Time       `json:"created_at"`
	UpdatedAt                 *time.Time       `json:"updated_at"`
	Upvotes                   int              `json:"upvotes"`
	Downvotes                 int              `json:"downvotes"`
	Author                    *BasicUser       `json:"author"`
	Assignee                  *BasicUser       `json:"assignee"`
	Assignees                 []*BasicUser     `json:"assignees"`
	Reviewers                 []*BasicUser     `json:"reviewers"`
	SourceProjectID           int              `json:"source_project_id"`
	TargetProjectID           int              `json:"target_project_id"`
	Labels                    Labels           `json:"labels"`
	Description               string           `json:"description"`
	Draft                     bool             `json:"draft"`
	WorkInProgress            bool             `json:"work_in_progress"`
	Milestone                 *Milestone       `json:"milestone"`
	MergeWhenPipelineSucceeds bool             `json:"merge_when_pipeline_succeeds"`
	DetailedMergeStatus       string           `json:"detailed_merge_status"`
	MergeError                string           `json:"merge_error"`
	MergedBy                  *BasicUser       `json:"merged_by"`
	MergedAt                  *time.Time       `json:"merged_at"`
	ClosedBy                  *BasicUser       `json:"closed_by"`
	ClosedAt                  *time.Time       `json:"closed_at"`
	Subscribed                bool             `json:"subscribed"`
	SHA                       string           `json:"sha"`
	MergeCommitSHA            string           `json:"merge_commit_sha"`
	SquashCommitSHA           string           `json:"squash_commit_sha"`
	UserNotesCount            int              `json:"user_notes_count"`
	ChangesCount              string           `json:"changes_count"`
	ShouldRemoveSourceBranch  bool             `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   bool             `json:"force_remove_source_branch"`
	AllowCollaboration        bool             `json:"allow_collaboration"`
	WebURL                    string           `json:"web_url"`
	References                *IssueReferences `json:"references"`
	DiscussionLocked          bool             `json:"discussion_locked"`
	Changes                   []struct {
		OldPath     string `json:"old_path"`
		NewPath     string `json:"new_path"`
		AMode       string `json:"a_mode"`
		BMode       string `json:"b_mode"`
		Diff        string `json:"diff"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
	} `json:"changes"`
	User struct {
		CanMerge bool `json:"can_merge"`
	} `json:"user"`
	TimeStats    *TimeStats    `json:"time_stats"`
	Squash       bool          `json:"squash"`
	Pipeline     *PipelineInfo `json:"pipeline"`
	HeadPipeline *Pipeline     `json:"head_pipeline"`
	DiffRefs     struct {
		BaseSha  string `json:"base_sha"`
		HeadSha  string `json:"head_sha"`
		StartSha string `json:"start_sha"`
	} `json:"diff_refs"`
	DivergedCommitsCount        int                    `json:"diverged_commits_count"`
	RebaseInProgress            bool                   `json:"rebase_in_progress"`
	ApprovalsBeforeMerge        int                    `json:"approvals_before_merge"`
	Reference                   string                 `json:"reference"`
	FirstContribution           bool                   `json:"first_contribution"`
	TaskCompletionStatus        *TasksCompletionStatus `json:"task_completion_status"`
	HasConflicts                bool                   `json:"has_conflicts"`
	BlockingDiscussionsResolved bool                   `json:"blocking_discussions_resolved"`
	Overflow                    bool                   `json:"overflow"`

	// Deprecated: This parameter is replaced by DetailedMergeStatus in GitLab 15.6.
	MergeStatus string `json:"merge_status"`
}

func (m MergeRequest) String() string {
	return Stringify(m)
}

// MergeRequestDiffVersion represents Gitlab merge request version.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-merge-request-diff-versions
type MergeRequestDiffVersion struct {
	ID             int        `json:"id"`
	HeadCommitSHA  string     `json:"head_commit_sha,omitempty"`
	BaseCommitSHA  string     `json:"base_commit_sha,omitempty"`
	StartCommitSHA string     `json:"start_commit_sha,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
	MergeRequestID int        `json:"merge_request_id,omitempty"`
	State          string     `json:"state,omitempty"`
	RealSize       string     `json:"real_size,omitempty"`
	Commits        []*Commit  `json:"commits,omitempty"`
	Diffs          []*Diff    `json:"diffs,omitempty"`
}

func (m MergeRequestDiffVersion) String() string {
	return Stringify(m)
}

// ListMergeRequestsOptions represents the available ListMergeRequests()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-merge-requests
type ListMergeRequestsOptions struct {
	ListOptions
	State                  *string           `url:"state,omitempty" json:"state,omitempty"`
	OrderBy                *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort                   *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Milestone              *string           `url:"milestone,omitempty" json:"milestone,omitempty"`
	View                   *string           `url:"view,omitempty" json:"view,omitempty"`
	Labels                 *Labels           `url:"labels,comma,omitempty" json:"labels,omitempty"`
	NotLabels              *Labels           `url:"not[labels],comma,omitempty" json:"not[labels],omitempty"`
	WithLabelsDetails      *bool             `url:"with_labels_details,omitempty" json:"with_labels_details,omitempty"`
	WithMergeStatusRecheck *bool             `url:"with_merge_status_recheck,omitempty" json:"with_merge_status_recheck,omitempty"`
	CreatedAfter           *time.Time        `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore          *time.Time        `url:"created_before,omitempty" json:"created_before,omitempty"`
	UpdatedAfter           *time.Time        `url:"updated_after,omitempty" json:"updated_after,omitempty"`
	UpdatedBefore          *time.Time        `url:"updated_before,omitempty" json:"updated_before,omitempty"`
	Scope                  *string           `url:"scope,omitempty" json:"scope,omitempty"`
	AuthorID               *int              `url:"author_id,omitempty" json:"author_id,omitempty"`
	AuthorUsername         *string           `url:"author_username,omitempty" json:"author_username,omitempty"`
	AssigneeID             *AssigneeIDValue  `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	ApproverIDs            *ApproverIDsValue `url:"approver_ids,omitempty" json:"approver_ids,omitempty"`
	ApprovedByIDs          *ApproverIDsValue `url:"approved_by_ids,omitempty" json:"approved_by_ids,omitempty"`
	ReviewerID             *ReviewerIDValue  `url:"reviewer_id,omitempty" json:"reviewer_id,omitempty"`
	ReviewerUsername       *string           `url:"reviewer_username,omitempty" json:"reviewer_username,omitempty"`
	MyReactionEmoji        *string           `url:"my_reaction_emoji,omitempty" json:"my_reaction_emoji,omitempty"`
	SourceBranch           *string           `url:"source_branch,omitempty" json:"source_branch,omitempty"`
	TargetBranch           *string           `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	Search                 *string           `url:"search,omitempty" json:"search,omitempty"`
	In                     *string           `url:"in,omitempty" json:"in,omitempty"`
	Draft                  *bool             `url:"draft,omitempty" json:"draft,omitempty"`
	WIP                    *string           `url:"wip,omitempty" json:"wip,omitempty"`
}

// ListMergeRequests gets all merge requests. The state parameter can be used
// to get only merge requests with a given state (opened, closed, or merged)
// or all of them (all). The pagination parameters page and per_page can be
// used to restrict the list of merge requests.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-merge-requests
func (s *MergeRequestsService) ListMergeRequests(opt *ListMergeRequestsOptions, options ...RequestOptionFunc) ([]*MergeRequest, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "merge_requests", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var m []*MergeRequest
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// ListProjectMergeRequestsOptions represents the available ListMergeRequests()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-project-merge-requests
type ListProjectMergeRequestsOptions struct {
	ListOptions
	IIDs                   *[]int            `url:"iids[],omitempty" json:"iids,omitempty"`
	State                  *string           `url:"state,omitempty" json:"state,omitempty"`
	OrderBy                *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort                   *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Milestone              *string           `url:"milestone,omitempty" json:"milestone,omitempty"`
	View                   *string           `url:"view,omitempty" json:"view,omitempty"`
	Labels                 *Labels           `url:"labels,comma,omitempty" json:"labels,omitempty"`
	NotLabels              *Labels           `url:"not[labels],comma,omitempty" json:"not[labels],omitempty"`
	WithLabelsDetails      *bool             `url:"with_labels_details,omitempty" json:"with_labels_details,omitempty"`
	WithMergeStatusRecheck *bool             `url:"with_merge_status_recheck,omitempty" json:"with_merge_status_recheck,omitempty"`
	CreatedAfter           *time.Time        `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore          *time.Time        `url:"created_before,omitempty" json:"created_before,omitempty"`
	UpdatedAfter           *time.Time        `url:"updated_after,omitempty" json:"updated_after,omitempty"`
	UpdatedBefore          *time.Time        `url:"updated_before,omitempty" json:"updated_before,omitempty"`
	Scope                  *string           `url:"scope,omitempty" json:"scope,omitempty"`
	AuthorID               *int              `url:"author_id,omitempty" json:"author_id,omitempty"`
	AuthorUsername         *string           `url:"author_username,omitempty" json:"author_username,omitempty"`
	AssigneeID             *AssigneeIDValue  `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	ApproverIDs            *ApproverIDsValue `url:"approver_ids,omitempty" json:"approver_ids,omitempty"`
	ApprovedByIDs          *ApproverIDsValue `url:"approved_by_ids,omitempty" json:"approved_by_ids,omitempty"`
	ReviewerID             *ReviewerIDValue  `url:"reviewer_id,omitempty" json:"reviewer_id,omitempty"`
	ReviewerUsername       *string           `url:"reviewer_username,omitempty" json:"reviewer_username,omitempty"`
	MyReactionEmoji        *string           `url:"my_reaction_emoji,omitempty" json:"my_reaction_emoji,omitempty"`
	SourceBranch           *string           `url:"source_branch,omitempty" json:"source_branch,omitempty"`
	TargetBranch           *string           `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	Search                 *string           `url:"search,omitempty" json:"search,omitempty"`
	Draft                  *bool             `url:"draft,omitempty" json:"draft,omitempty"`
	WIP                    *string           `url:"wip,omitempty" json:"wip,omitempty"`
}

// ListProjectMergeRequests gets all merge requests for this project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-project-merge-requests
func (s *MergeRequestsService) ListProjectMergeRequests(pid interface{}, opt *ListProjectMergeRequestsOptions, options ...RequestOptionFunc) ([]*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var m []*MergeRequest
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// ListGroupMergeRequestsOptions represents the available ListGroupMergeRequests()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-group-merge-requests
type ListGroupMergeRequestsOptions struct {
	ListOptions
	State                  *string           `url:"state,omitempty" json:"state,omitempty"`
	OrderBy                *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort                   *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Milestone              *string           `url:"milestone,omitempty" json:"milestone,omitempty"`
	View                   *string           `url:"view,omitempty" json:"view,omitempty"`
	Labels                 *Labels           `url:"labels,comma,omitempty" json:"labels,omitempty"`
	NotLabels              *Labels           `url:"not[labels],comma,omitempty" json:"not[labels],omitempty"`
	WithLabelsDetails      *bool             `url:"with_labels_details,omitempty" json:"with_labels_details,omitempty"`
	WithMergeStatusRecheck *bool             `url:"with_merge_status_recheck,omitempty" json:"with_merge_status_recheck,omitempty"`
	CreatedAfter           *time.Time        `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore          *time.Time        `url:"created_before,omitempty" json:"created_before,omitempty"`
	UpdatedAfter           *time.Time        `url:"updated_after,omitempty" json:"updated_after,omitempty"`
	UpdatedBefore          *time.Time        `url:"updated_before,omitempty" json:"updated_before,omitempty"`
	Scope                  *string           `url:"scope,omitempty" json:"scope,omitempty"`
	AuthorID               *int              `url:"author_id,omitempty" json:"author_id,omitempty"`
	AuthorUsername         *string           `url:"author_username,omitempty" json:"author_username,omitempty"`
	AssigneeID             *AssigneeIDValue  `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	ApproverIDs            *ApproverIDsValue `url:"approver_ids,omitempty" json:"approver_ids,omitempty"`
	ApprovedByIDs          *ApproverIDsValue `url:"approved_by_ids,omitempty" json:"approved_by_ids,omitempty"`
	ReviewerID             *ReviewerIDValue  `url:"reviewer_id,omitempty" json:"reviewer_id,omitempty"`
	ReviewerUsername       *string           `url:"reviewer_username,omitempty" json:"reviewer_username,omitempty"`
	MyReactionEmoji        *string           `url:"my_reaction_emoji,omitempty" json:"my_reaction_emoji,omitempty"`
	SourceBranch           *string           `url:"source_branch,omitempty" json:"source_branch,omitempty"`
	TargetBranch           *string           `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	Search                 *string           `url:"search,omitempty" json:"search,omitempty"`
	In                     *string           `url:"in,omitempty" json:"in,omitempty"`
	Draft                  *bool             `url:"draft,omitempty" json:"draft,omitempty"`
	WIP                    *string           `url:"wip,omitempty" json:"wip,omitempty"`
}

// ListGroupMergeRequests gets all merge requests for this group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-group-merge-requests
func (s *MergeRequestsService) ListGroupMergeRequests(gid interface{}, opt *ListGroupMergeRequestsOptions, options ...RequestOptionFunc) ([]*MergeRequest, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/merge_requests", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var m []*MergeRequest
	resp, err := s.client.Do(req, &m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// GetMergeRequestsOptions represents the available GetMergeRequests()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-mr
type GetMergeRequestsOptions struct {
	RenderHTML                  *bool `url:"render_html,omitempty" json:"render_html,omitempty"`
	IncludeDivergedCommitsCount *bool `url:"include_diverged_commits_count,omitempty" json:"include_diverged_commits_count,omitempty"`
	IncludeRebaseInProgress     *bool `url:"include_rebase_in_progress,omitempty" json:"include_rebase_in_progress,omitempty"`
}

// GetMergeRequest shows information about a single merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-mr
func (s *MergeRequestsService) GetMergeRequest(pid interface{}, mergeRequest int, opt *GetMergeRequestsOptions, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// GetMergeRequestApprovals gets information about a merge requests approvals
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#merge-request-level-mr-approvals
func (s *MergeRequestsService) GetMergeRequestApprovals(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*MergeRequestApprovals, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/approvals", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	a := new(MergeRequestApprovals)
	resp, err := s.client.Do(req, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, nil
}

// GetMergeRequestCommitsOptions represents the available GetMergeRequestCommits()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-merge-request-commits
type GetMergeRequestCommitsOptions ListOptions

// GetMergeRequestCommits gets a list of merge request commits.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-merge-request-commits
func (s *MergeRequestsService) GetMergeRequestCommits(pid interface{}, mergeRequest int, opt *GetMergeRequestCommitsOptions, options ...RequestOptionFunc) ([]*Commit, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/commits", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var c []*Commit
	resp, err := s.client.Do(req, &c)
	if err != nil {
		return nil, resp, err
	}

	return c, resp, nil
}

// GetMergeRequestChangesOptions represents the available GetMergeRequestChanges()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-merge-request-changes
type GetMergeRequestChangesOptions struct {
	AccessRawDiffs *bool `url:"access_raw_diffs,omitempty" json:"access_raw_diffs,omitempty"`
}

// GetMergeRequestChanges shows information about the merge request including
// its files and changes.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-merge-request-changes
func (s *MergeRequestsService) GetMergeRequestChanges(pid interface{}, mergeRequest int, opt *GetMergeRequestChangesOptions, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/changes", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// GetMergeRequestParticipants gets a list of merge request participants.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-single-merge-request-participants
func (s *MergeRequestsService) GetMergeRequestParticipants(pid interface{}, mergeRequest int, options ...RequestOptionFunc) ([]*BasicUser, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/participants", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*BasicUser
	resp, err := s.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// ListMergeRequestPipelines gets all pipelines for the provided merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-merge-request-pipelines
func (s *MergeRequestsService) ListMergeRequestPipelines(pid interface{}, mergeRequest int, options ...RequestOptionFunc) ([]*PipelineInfo, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/pipelines", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*PipelineInfo
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// CreateMergeRequestPipeline creates a new pipeline for a merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#create-merge-request-pipeline
func (s *MergeRequestsService) CreateMergeRequestPipeline(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*PipelineInfo, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/pipelines", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(PipelineInfo)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// GetIssuesClosedOnMergeOptions represents the available GetIssuesClosedOnMerge()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-issues-that-close-on-merge
type GetIssuesClosedOnMergeOptions ListOptions

// GetIssuesClosedOnMerge gets all the issues that would be closed by merging the
// provided merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#list-issues-that-close-on-merge
func (s *MergeRequestsService) GetIssuesClosedOnMerge(pid interface{}, mergeRequest int, opt *GetIssuesClosedOnMergeOptions, options ...RequestOptionFunc) ([]*Issue, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/closes_issues", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var i []*Issue
	resp, err := s.client.Do(req, &i)
	if err != nil {
		return nil, resp, err
	}

	return i, resp, nil
}

// CreateMergeRequestOptions represents the available CreateMergeRequest()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#create-mr
type CreateMergeRequestOptions struct {
	Title              *string `url:"title,omitempty" json:"title,omitempty"`
	Description        *string `url:"description,omitempty" json:"description,omitempty"`
	SourceBranch       *string `url:"source_branch,omitempty" json:"source_branch,omitempty"`
	TargetBranch       *string `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	Labels             *Labels `url:"labels,comma,omitempty" json:"labels,omitempty"`
	AssigneeID         *int    `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	AssigneeIDs        *[]int  `url:"assignee_ids,omitempty" json:"assignee_ids,omitempty"`
	ReviewerIDs        *[]int  `url:"reviewer_ids,omitempty" json:"reviewer_ids,omitempty"`
	TargetProjectID    *int    `url:"target_project_id,omitempty" json:"target_project_id,omitempty"`
	MilestoneID        *int    `url:"milestone_id,omitempty" json:"milestone_id,omitempty"`
	RemoveSourceBranch *bool   `url:"remove_source_branch,omitempty" json:"remove_source_branch,omitempty"`
	Squash             *bool   `url:"squash,omitempty" json:"squash,omitempty"`
	AllowCollaboration *bool   `url:"allow_collaboration,omitempty" json:"allow_collaboration,omitempty"`
}

// CreateMergeRequest creates a new merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#create-mr
func (s *MergeRequestsService) CreateMergeRequest(pid interface{}, opt *CreateMergeRequestOptions, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// UpdateMergeRequestOptions represents the available UpdateMergeRequest()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#update-mr
type UpdateMergeRequestOptions struct {
	Title              *string `url:"title,omitempty" json:"title,omitempty"`
	Description        *string `url:"description,omitempty" json:"description,omitempty"`
	TargetBranch       *string `url:"target_branch,omitempty" json:"target_branch,omitempty"`
	AssigneeID         *int    `url:"assignee_id,omitempty" json:"assignee_id,omitempty"`
	AssigneeIDs        *[]int  `url:"assignee_ids,omitempty" json:"assignee_ids,omitempty"`
	ReviewerIDs        *[]int  `url:"reviewer_ids,omitempty" json:"reviewer_ids,omitempty"`
	Labels             *Labels `url:"labels,comma,omitempty" json:"labels,omitempty"`
	AddLabels          *Labels `url:"add_labels,comma,omitempty" json:"add_labels,omitempty"`
	RemoveLabels       *Labels `url:"remove_labels,comma,omitempty" json:"remove_labels,omitempty"`
	MilestoneID        *int    `url:"milestone_id,omitempty" json:"milestone_id,omitempty"`
	StateEvent         *string `url:"state_event,omitempty" json:"state_event,omitempty"`
	RemoveSourceBranch *bool   `url:"remove_source_branch,omitempty" json:"remove_source_branch,omitempty"`
	Squash             *bool   `url:"squash,omitempty" json:"squash,omitempty"`
	DiscussionLocked   *bool   `url:"discussion_locked,omitempty" json:"discussion_locked,omitempty"`
	AllowCollaboration *bool   `url:"allow_collaboration,omitempty" json:"allow_collaboration,omitempty"`
}

// UpdateMergeRequest updates an existing project milestone.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#update-mr
func (s *MergeRequestsService) UpdateMergeRequest(pid interface{}, mergeRequest int, opt *UpdateMergeRequestOptions, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// DeleteMergeRequest deletes a merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#delete-a-merge-request
func (s *MergeRequestsService) DeleteMergeRequest(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// AcceptMergeRequestOptions represents the available AcceptMergeRequest()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#merge-a-merge-request
type AcceptMergeRequestOptions struct {
	MergeCommitMessage        *string `url:"merge_commit_message,omitempty" json:"merge_commit_message,omitempty"`
	SquashCommitMessage       *string `url:"squash_commit_message,omitempty" json:"squash_commit_message,omitempty"`
	Squash                    *bool   `url:"squash,omitempty" json:"squash,omitempty"`
	ShouldRemoveSourceBranch  *bool   `url:"should_remove_source_branch,omitempty" json:"should_remove_source_branch,omitempty"`
	MergeWhenPipelineSucceeds *bool   `url:"merge_when_pipeline_succeeds,omitempty" json:"merge_when_pipeline_succeeds,omitempty"`
	SHA                       *string `url:"sha,omitempty" json:"sha,omitempty"`
}

// AcceptMergeRequest merges changes submitted with MR using this API. If merge
// success you get 200 OK. If it has some conflicts and can not be merged - you
// get 405 and error message 'Branch cannot be merged'. If merge request is
// already merged or closed - you get 405 and error message 'Method Not Allowed'
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#merge-a-merge-request
func (s *MergeRequestsService) AcceptMergeRequest(pid interface{}, mergeRequest int, opt *AcceptMergeRequestOptions, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/merge", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CancelMergeWhenPipelineSucceeds cancels a merge when pipeline succeeds. If
// you don't have permissions to accept this merge request - you'll get a 401.
// If the merge request is already merged or closed - you get 405 and error
// message 'Method Not Allowed'. In case the merge request is not set to be
// merged when the pipeline succeeds, you'll also get a 406 error.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#cancel-merge-when-pipeline-succeeds
func (s *MergeRequestsService) CancelMergeWhenPipelineSucceeds(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/cancel_merge_when_pipeline_succeeds", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// RebaseMergeRequest automatically rebases the source_branch of the merge
// request against its target_branch. If you don’t have permissions to push
// to the merge request’s source branch, you’ll get a 403 Forbidden response.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#rebase-a-merge-request
func (s *MergeRequestsService) RebaseMergeRequest(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/rebase", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPut, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// GetMergeRequestDiffVersionsOptions represents the available
// GetMergeRequestDiffVersions() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-merge-request-diff-versions
type GetMergeRequestDiffVersionsOptions ListOptions

// GetMergeRequestDiffVersions get a list of merge request diff versions.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-merge-request-diff-versions
func (s *MergeRequestsService) GetMergeRequestDiffVersions(pid interface{}, mergeRequest int, opt *GetMergeRequestDiffVersionsOptions, options ...RequestOptionFunc) ([]*MergeRequestDiffVersion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/versions", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var v []*MergeRequestDiffVersion
	resp, err := s.client.Do(req, &v)
	if err != nil {
		return nil, resp, err
	}

	return v, resp, nil
}

// GetSingleMergeRequestDiffVersion get a single MR diff version
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-a-single-merge-request-diff-version
func (s *MergeRequestsService) GetSingleMergeRequestDiffVersion(pid interface{}, mergeRequest, version int, options ...RequestOptionFunc) (*MergeRequestDiffVersion, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/versions/%d", PathEscape(project), mergeRequest, version)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	v := new(MergeRequestDiffVersion)
	resp, err := s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}

	return v, resp, nil
}

// SubscribeToMergeRequest subscribes the authenticated user to the given merge
// request to receive notifications. If the user is already subscribed to the
// merge request, the status code 304 is returned.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#subscribe-to-a-merge-request
func (s *MergeRequestsService) SubscribeToMergeRequest(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/subscribe", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// UnsubscribeFromMergeRequest unsubscribes the authenticated user from the
// given merge request to not receive notifications from that merge request.
// If the user is not subscribed to the merge request, status code 304 is
// returned.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#unsubscribe-from-a-merge-request
func (s *MergeRequestsService) UnsubscribeFromMergeRequest(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/unsubscribe", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	m := new(MergeRequest)
	resp, err := s.client.Do(req, m)
	if err != nil {
		return nil, resp, err
	}

	return m, resp, nil
}

// CreateTodo manually creates a todo for the current user on a merge request.
// If there already exists a todo for the user on that merge request,
// status code 304 is returned.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#create-a-to-do-item
func (s *MergeRequestsService) CreateTodo(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*Todo, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/todo", PathEscape(project), mergeRequest)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	t := new(Todo)
	resp, err := s.client.Do(req, t)
	if err != nil {
		return nil, resp, err
	}

	return t, resp, nil
}

// SetTimeEstimate sets the time estimate for a single project merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#set-a-time-estimate-for-a-merge-request
func (s *MergeRequestsService) SetTimeEstimate(pid interface{}, mergeRequest int, opt *SetTimeEstimateOptions, options ...RequestOptionFunc) (*TimeStats, *Response, error) {
	return s.timeStats.setTimeEstimate(pid, "merge_requests", mergeRequest, opt, options...)
}

// ResetTimeEstimate resets the time estimate for a single project merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#reset-the-time-estimate-for-a-merge-request
func (s *MergeRequestsService) ResetTimeEstimate(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*TimeStats, *Response, error) {
	return s.timeStats.resetTimeEstimate(pid, "merge_requests", mergeRequest, options...)
}

// AddSpentTime adds spent time for a single project merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#add-spent-time-for-a-merge-request
func (s *MergeRequestsService) AddSpentTime(pid interface{}, mergeRequest int, opt *AddSpentTimeOptions, options ...RequestOptionFunc) (*TimeStats, *Response, error) {
	return s.timeStats.addSpentTime(pid, "merge_requests", mergeRequest, opt, options...)
}

// ResetSpentTime resets the spent time for a single project merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#reset-spent-time-for-a-merge-request
func (s *MergeRequestsService) ResetSpentTime(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*TimeStats, *Response, error) {
	return s.timeStats.resetSpentTime(pid, "merge_requests", mergeRequest, options...)
}

// GetTimeSpent gets the spent time for a single project merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_requests.html#get-time-tracking-stats
func (s *MergeRequestsService) GetTimeSpent(pid interface{}, mergeRequest int, options ...RequestOptionFunc) (*TimeStats, *Response, error) {
	return s.timeStats.getTimeSpent(pid, "merge_requests", mergeRequest, options...)
}
