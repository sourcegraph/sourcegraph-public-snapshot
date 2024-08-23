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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// ProjectsService handles communication with the repositories related methods
// of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html
type ProjectsService struct {
	client *Client
}

// Project represents a GitLab project.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html
type Project struct {
	ID                                        int                        `json:"id"`
	Description                               string                     `json:"description"`
	DefaultBranch                             string                     `json:"default_branch"`
	Public                                    bool                       `json:"public"`
	Visibility                                VisibilityValue            `json:"visibility"`
	SSHURLToRepo                              string                     `json:"ssh_url_to_repo"`
	HTTPURLToRepo                             string                     `json:"http_url_to_repo"`
	WebURL                                    string                     `json:"web_url"`
	ReadmeURL                                 string                     `json:"readme_url"`
	TagList                                   []string                   `json:"tag_list"`
	Topics                                    []string                   `json:"topics"`
	Owner                                     *User                      `json:"owner"`
	Name                                      string                     `json:"name"`
	NameWithNamespace                         string                     `json:"name_with_namespace"`
	Path                                      string                     `json:"path"`
	PathWithNamespace                         string                     `json:"path_with_namespace"`
	IssuesEnabled                             bool                       `json:"issues_enabled"`
	OpenIssuesCount                           int                        `json:"open_issues_count"`
	MergeRequestsEnabled                      bool                       `json:"merge_requests_enabled"`
	ApprovalsBeforeMerge                      int                        `json:"approvals_before_merge"`
	JobsEnabled                               bool                       `json:"jobs_enabled"`
	WikiEnabled                               bool                       `json:"wiki_enabled"`
	SnippetsEnabled                           bool                       `json:"snippets_enabled"`
	ResolveOutdatedDiffDiscussions            bool                       `json:"resolve_outdated_diff_discussions"`
	ContainerExpirationPolicy                 *ContainerExpirationPolicy `json:"container_expiration_policy,omitempty"`
	ContainerRegistryEnabled                  bool                       `json:"container_registry_enabled"`
	ContainerRegistryAccessLevel              AccessControlValue         `json:"container_registry_access_level"`
	ContainerRegistryImagePrefix              string                     `json:"container_registry_image_prefix,omitempty"`
	CreatedAt                                 *time.Time                 `json:"created_at,omitempty"`
	LastActivityAt                            *time.Time                 `json:"last_activity_at,omitempty"`
	CreatorID                                 int                        `json:"creator_id"`
	Namespace                                 *ProjectNamespace          `json:"namespace"`
	Permissions                               *Permissions               `json:"permissions"`
	MarkedForDeletionAt                       *ISOTime                   `json:"marked_for_deletion_at"`
	EmptyRepo                                 bool                       `json:"empty_repo"`
	Archived                                  bool                       `json:"archived"`
	AvatarURL                                 string                     `json:"avatar_url"`
	LicenseURL                                string                     `json:"license_url"`
	License                                   *ProjectLicense            `json:"license"`
	SharedRunnersEnabled                      bool                       `json:"shared_runners_enabled"`
	GroupRunnersEnabled                       bool                       `json:"group_runners_enabled"`
	RunnerTokenExpirationInterval             int                        `json:"runner_token_expiration_interval"`
	ForksCount                                int                        `json:"forks_count"`
	StarCount                                 int                        `json:"star_count"`
	RunnersToken                              string                     `json:"runners_token"`
	AllowMergeOnSkippedPipeline               bool                       `json:"allow_merge_on_skipped_pipeline"`
	OnlyAllowMergeIfPipelineSucceeds          bool                       `json:"only_allow_merge_if_pipeline_succeeds"`
	OnlyAllowMergeIfAllDiscussionsAreResolved bool                       `json:"only_allow_merge_if_all_discussions_are_resolved"`
	RemoveSourceBranchAfterMerge              bool                       `json:"remove_source_branch_after_merge"`
	PrintingMergeRequestLinkEnabled           bool                       `json:"printing_merge_request_link_enabled"`
	LFSEnabled                                bool                       `json:"lfs_enabled"`
	RepositoryStorage                         string                     `json:"repository_storage"`
	RequestAccessEnabled                      bool                       `json:"request_access_enabled"`
	MergeMethod                               MergeMethodValue           `json:"merge_method"`
	CanCreateMergeRequestIn                   bool                       `json:"can_create_merge_request_in"`
	ForkedFromProject                         *ForkParent                `json:"forked_from_project"`
	Mirror                                    bool                       `json:"mirror"`
	MirrorUserID                              int                        `json:"mirror_user_id"`
	MirrorTriggerBuilds                       bool                       `json:"mirror_trigger_builds"`
	OnlyMirrorProtectedBranches               bool                       `json:"only_mirror_protected_branches"`
	MirrorOverwritesDivergedBranches          bool                       `json:"mirror_overwrites_diverged_branches"`
	PackagesEnabled                           bool                       `json:"packages_enabled"`
	ServiceDeskEnabled                        bool                       `json:"service_desk_enabled"`
	ServiceDeskAddress                        string                     `json:"service_desk_address"`
	IssuesAccessLevel                         AccessControlValue         `json:"issues_access_level"`
	ReleasesAccessLevel                       AccessControlValue         `json:"releases_access_level,omitempty"`
	RepositoryAccessLevel                     AccessControlValue         `json:"repository_access_level"`
	MergeRequestsAccessLevel                  AccessControlValue         `json:"merge_requests_access_level"`
	ForkingAccessLevel                        AccessControlValue         `json:"forking_access_level"`
	WikiAccessLevel                           AccessControlValue         `json:"wiki_access_level"`
	BuildsAccessLevel                         AccessControlValue         `json:"builds_access_level"`
	SnippetsAccessLevel                       AccessControlValue         `json:"snippets_access_level"`
	PagesAccessLevel                          AccessControlValue         `json:"pages_access_level"`
	OperationsAccessLevel                     AccessControlValue         `json:"operations_access_level"`
	AnalyticsAccessLevel                      AccessControlValue         `json:"analytics_access_level"`
	EnvironmentsAccessLevel                   AccessControlValue         `json:"environments_access_level"`
	FeatureFlagsAccessLevel                   AccessControlValue         `json:"feature_flags_access_level"`
	InfrastructureAccessLevel                 AccessControlValue         `json:"infrastructure_access_level"`
	MonitorAccessLevel                        AccessControlValue         `json:"monitor_access_level"`
	AutocloseReferencedIssues                 bool                       `json:"autoclose_referenced_issues"`
	SuggestionCommitMessage                   string                     `json:"suggestion_commit_message"`
	SquashOption                              SquashOptionValue          `json:"squash_option"`
	EnforceAuthChecksOnUploads                bool                       `json:"enforce_auth_checks_on_uploads,omitempty"`
	SharedWithGroups                          []struct {
		GroupID          int    `json:"group_id"`
		GroupName        string `json:"group_name"`
		GroupFullPath    string `json:"group_full_path"`
		GroupAccessLevel int    `json:"group_access_level"`
	} `json:"shared_with_groups"`
	Statistics                               *Statistics        `json:"statistics"`
	Links                                    *Links             `json:"_links,omitempty"`
	ImportURL                                string             `json:"import_url"`
	ImportType                               string             `json:"import_type"`
	ImportStatus                             string             `json:"import_status"`
	ImportError                              string             `json:"import_error"`
	CIDefaultGitDepth                        int                `json:"ci_default_git_depth"`
	CIForwardDeploymentEnabled               bool               `json:"ci_forward_deployment_enabled"`
	CISeperateCache                          bool               `json:"ci_separated_caches"`
	CIJobTokenScopeEnabled                   bool               `json:"ci_job_token_scope_enabled"`
	CIOptInJWT                               bool               `json:"ci_opt_in_jwt"`
	CIAllowForkPipelinesToRunInParentProject bool               `json:"ci_allow_fork_pipelines_to_run_in_parent_project"`
	PublicJobs                               bool               `json:"public_jobs"`
	BuildTimeout                             int                `json:"build_timeout"`
	AutoCancelPendingPipelines               string             `json:"auto_cancel_pending_pipelines"`
	CIConfigPath                             string             `json:"ci_config_path"`
	CustomAttributes                         []*CustomAttribute `json:"custom_attributes"`
	ComplianceFrameworks                     []string           `json:"compliance_frameworks"`
	BuildCoverageRegex                       string             `json:"build_coverage_regex"`
	IssuesTemplate                           string             `json:"issues_template"`
	MergeRequestsTemplate                    string             `json:"merge_requests_template"`
	IssueBranchTemplate                      string             `json:"issue_branch_template"`
	KeepLatestArtifact                       bool               `json:"keep_latest_artifact"`
	MergePipelinesEnabled                    bool               `json:"merge_pipelines_enabled"`
	MergeTrainsEnabled                       bool               `json:"merge_trains_enabled"`
	RestrictUserDefinedVariables             bool               `json:"restrict_user_defined_variables"`
	MergeCommitTemplate                      string             `json:"merge_commit_template"`
	SquashCommitTemplate                     string             `json:"squash_commit_template"`
	AutoDevopsDeployStrategy                 string             `json:"auto_devops_deploy_strategy"`
	AutoDevopsEnabled                        bool               `json:"auto_devops_enabled"`
	BuildGitStrategy                         string             `json:"build_git_strategy"`
	EmailsDisabled                           bool               `json:"emails_disabled"`
	ExternalAuthorizationClassificationLabel string             `json:"external_authorization_classification_label"`
	RequirementsEnabled                      bool               `json:"requirements_enabled"`
	RequirementsAccessLevel                  AccessControlValue `json:"requirements_access_level"`
	SecurityAndComplianceEnabled             bool               `json:"security_and_compliance_enabled"`
	SecurityAndComplianceAccessLevel         AccessControlValue `json:"security_and_compliance_access_level"`
	MergeRequestDefaultTargetSelf            bool               `json:"mr_default_target_self"`

	// Deprecated: This parameter has been renamed to PublicJobs in GitLab 9.0.
	PublicBuilds bool `json:"public_builds"`
}

// BasicProject included in other service responses (such as todos).
type BasicProject struct {
	ID                int        `json:"id"`
	Description       string     `json:"description"`
	Name              string     `json:"name"`
	NameWithNamespace string     `json:"name_with_namespace"`
	Path              string     `json:"path"`
	PathWithNamespace string     `json:"path_with_namespace"`
	CreatedAt         *time.Time `json:"created_at"`
}

// ContainerExpirationPolicy represents the container expiration policy.
type ContainerExpirationPolicy struct {
	Cadence         string     `json:"cadence"`
	KeepN           int        `json:"keep_n"`
	OlderThan       string     `json:"older_than"`
	NameRegex       string     `json:"name_regex"`
	NameRegexDelete string     `json:"name_regex_delete"`
	NameRegexKeep   string     `json:"name_regex_keep"`
	Enabled         bool       `json:"enabled"`
	NextRunAt       *time.Time `json:"next_run_at"`
}

// ForkParent represents the parent project when this is a fork.
type ForkParent struct {
	HTTPURLToRepo     string `json:"http_url_to_repo"`
	ID                int    `json:"id"`
	Name              string `json:"name"`
	NameWithNamespace string `json:"name_with_namespace"`
	Path              string `json:"path"`
	PathWithNamespace string `json:"path_with_namespace"`
	WebURL            string `json:"web_url"`
}

// GroupAccess represents group access.
type GroupAccess struct {
	AccessLevel       AccessLevelValue       `json:"access_level"`
	NotificationLevel NotificationLevelValue `json:"notification_level"`
}

// Links represents a project web links for self, issues, merge_requests,
// repo_branches, labels, events, members.
type Links struct {
	Self          string `json:"self"`
	Issues        string `json:"issues"`
	MergeRequests string `json:"merge_requests"`
	RepoBranches  string `json:"repo_branches"`
	Labels        string `json:"labels"`
	Events        string `json:"events"`
	Members       string `json:"members"`
	ClusterAgents string `json:"cluster_agents"`
}

// Permissions represents permissions.
type Permissions struct {
	ProjectAccess *ProjectAccess `json:"project_access"`
	GroupAccess   *GroupAccess   `json:"group_access"`
}

// ProjectAccess represents project access.
type ProjectAccess struct {
	AccessLevel       AccessLevelValue       `json:"access_level"`
	NotificationLevel NotificationLevelValue `json:"notification_level"`
}

// ProjectLicense represent the license for a project.
type ProjectLicense struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	HTMLURL   string `json:"html_url"`
	SourceURL string `json:"source_url"`
}

// ProjectNamespace represents a project namespace.
type ProjectNamespace struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Kind      string `json:"kind"`
	FullPath  string `json:"full_path"`
	ParentID  int    `json:"parent_id"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// Repository represents a repository.
type Repository struct {
	Name              string          `json:"name"`
	Description       string          `json:"description"`
	WebURL            string          `json:"web_url"`
	AvatarURL         string          `json:"avatar_url"`
	GitSSHURL         string          `json:"git_ssh_url"`
	GitHTTPURL        string          `json:"git_http_url"`
	Namespace         string          `json:"namespace"`
	Visibility        VisibilityValue `json:"visibility"`
	PathWithNamespace string          `json:"path_with_namespace"`
	DefaultBranch     string          `json:"default_branch"`
	Homepage          string          `json:"homepage"`
	URL               string          `json:"url"`
	SSHURL            string          `json:"ssh_url"`
	HTTPURL           string          `json:"http_url"`
}

// Statistics represents a statistics record for a group or project.
type Statistics struct {
	CommitCount           int64 `json:"commit_count"`
	StorageSize           int64 `json:"storage_size"`
	RepositorySize        int64 `json:"repository_size"`
	WikiSize              int64 `json:"wiki_size"`
	LFSObjectsSize        int64 `json:"lfs_objects_size"`
	JobArtifactsSize      int64 `json:"job_artifacts_size"`
	PipelineArtifactsSize int64 `json:"pipeline_artifacts_size"`
	PackagesSize          int64 `json:"packages_size"`
	SnippetsSize          int64 `json:"snippets_size"`
	UploadsSize           int64 `json:"uploads_size"`
}

func (s Project) String() string {
	return Stringify(s)
}

// ProjectApprovalRule represents a GitLab project approval rule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#get-project-level-rules
type ProjectApprovalRule struct {
	ID                            int                `json:"id"`
	Name                          string             `json:"name"`
	RuleType                      string             `json:"rule_type"`
	EligibleApprovers             []*BasicUser       `json:"eligible_approvers"`
	ApprovalsRequired             int                `json:"approvals_required"`
	Users                         []*BasicUser       `json:"users"`
	Groups                        []*Group           `json:"groups"`
	ContainsHiddenGroups          bool               `json:"contains_hidden_groups"`
	ProtectedBranches             []*ProtectedBranch `json:"protected_branches"`
	AppliesToAllProtectedBranches bool               `json:"applies_to_all_protected_branches"`
}

func (s ProjectApprovalRule) String() string {
	return Stringify(s)
}

// ListProjectsOptions represents the available ListProjects() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#list-all-projects
type ListProjectsOptions struct {
	ListOptions
	Archived                 *bool             `url:"archived,omitempty" json:"archived,omitempty"`
	IDAfter                  *int              `url:"id_after,omitempty" json:"id_after,omitempty"`
	IDBefore                 *int              `url:"id_before,omitempty" json:"id_before,omitempty"`
	Imported                 *bool             `url:"imported,omitempty" json:"imported,omitempty"`
	LastActivityAfter        *time.Time        `url:"last_activity_after,omitempty" json:"last_activity_after,omitempty"`
	LastActivityBefore       *time.Time        `url:"last_activity_before,omitempty" json:"last_activity_before,omitempty"`
	Membership               *bool             `url:"membership,omitempty" json:"membership,omitempty"`
	MinAccessLevel           *AccessLevelValue `url:"min_access_level,omitempty" json:"min_access_level,omitempty"`
	OrderBy                  *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Owned                    *bool             `url:"owned,omitempty" json:"owned,omitempty"`
	RepositoryChecksumFailed *bool             `url:"repository_checksum_failed,omitempty" json:"repository_checksum_failed,omitempty"`
	RepositoryStorage        *string           `url:"repository_storage,omitempty" json:"repository_storage,omitempty"`
	Search                   *string           `url:"search,omitempty" json:"search,omitempty"`
	SearchNamespaces         *bool             `url:"search_namespaces,omitempty" json:"search_namespaces,omitempty"`
	Simple                   *bool             `url:"simple,omitempty" json:"simple,omitempty"`
	Sort                     *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Starred                  *bool             `url:"starred,omitempty" json:"starred,omitempty"`
	Statistics               *bool             `url:"statistics,omitempty" json:"statistics,omitempty"`
	Topic                    *string           `url:"topic,omitempty" json:"topic,omitempty"`
	Visibility               *VisibilityValue  `url:"visibility,omitempty" json:"visibility,omitempty"`
	WikiChecksumFailed       *bool             `url:"wiki_checksum_failed,omitempty" json:"wiki_checksum_failed,omitempty"`
	WithCustomAttributes     *bool             `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
	WithIssuesEnabled        *bool             `url:"with_issues_enabled,omitempty" json:"with_issues_enabled,omitempty"`
	WithMergeRequestsEnabled *bool             `url:"with_merge_requests_enabled,omitempty" json:"with_merge_requests_enabled,omitempty"`
	WithProgrammingLanguage  *string           `url:"with_programming_language,omitempty" json:"with_programming_language,omitempty"`
}

// ListProjects gets a list of projects accessible by the authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#list-all-projects
func (s *ProjectsService) ListProjects(opt *ListProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "projects", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*Project
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ListUserProjects gets a list of projects for the given user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-user-projects
func (s *ProjectsService) ListUserProjects(uid interface{}, opt *ListProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	user, err := parseID(uid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("users/%s/projects", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*Project
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ListUserStarredProjects gets a list of projects starred by the given user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-projects-starred-by-a-user
func (s *ProjectsService) ListUserStarredProjects(uid interface{}, opt *ListProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	user, err := parseID(uid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("users/%s/starred_projects", user)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*Project
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ProjectUser represents a GitLab project user.
type ProjectUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

// ListProjectUserOptions represents the available ListProjectsUsers() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#get-project-users
type ListProjectUserOptions struct {
	ListOptions
	Search *string `url:"search,omitempty" json:"search,omitempty"`
}

// ListProjectsUsers gets a list of users for the given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-project-users
func (s *ProjectsService) ListProjectsUsers(pid interface{}, opt *ListProjectUserOptions, options ...RequestOptionFunc) ([]*ProjectUser, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/users", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*ProjectUser
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ProjectGroup represents a GitLab project group.
type ProjectGroup struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
	FullName  string `json:"full_name"`
	FullPath  string `json:"full_path"`
}

// ListProjectGroupOptions represents the available ListProjectsGroups() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#list-a-projects-groups
type ListProjectGroupOptions struct {
	ListOptions
	Search               *string           `url:"search,omitempty" json:"search,omitempty"`
	SharedMinAccessLevel *AccessLevelValue `url:"shared_min_access_level,omitempty" json:"shared_min_access_level,omitempty"`
	SharedVisiableOnly   *bool             `url:"shared_visible_only,omitempty" json:"shared_visible_only,omitempty"`
	SkipGroups           *[]int            `url:"skip_groups,omitempty" json:"skip_groups,omitempty"`
	WithShared           *bool             `url:"with_shared,omitempty" json:"with_shared,omitempty"`
}

// ListProjectsGroups gets a list of groups for the given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-a-projects-groups
func (s *ProjectsService) ListProjectsGroups(pid interface{}, opt *ListProjectGroupOptions, options ...RequestOptionFunc) ([]*ProjectGroup, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/groups", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*ProjectGroup
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ProjectLanguages is a map of strings because the response is arbitrary
//
// Gitlab API docs: https://docs.gitlab.com/ee/api/projects.html#languages
type ProjectLanguages map[string]float32

// GetProjectLanguages gets a list of languages used by the project
//
// GitLab API docs:  https://docs.gitlab.com/ee/api/projects.html#languages
func (s *ProjectsService) GetProjectLanguages(pid interface{}, options ...RequestOptionFunc) (*ProjectLanguages, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/languages", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(ProjectLanguages)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// GetProjectOptions represents the available GetProject() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#get-single-project
type GetProjectOptions struct {
	License              *bool `url:"license,omitempty" json:"license,omitempty"`
	Statistics           *bool `url:"statistics,omitempty" json:"statistics,omitempty"`
	WithCustomAttributes *bool `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
}

// GetProject gets a specific project, identified by project ID or
// NAMESPACE/PROJECT_NAME, which is owned by the authenticated user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-single-project
func (s *ProjectsService) GetProject(pid interface{}, opt *GetProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// CreateProjectOptions represents the available CreateProject() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
type CreateProjectOptions struct {
	AllowMergeOnSkippedPipeline               *bool                                `url:"allow_merge_on_skipped_pipeline,omitempty" json:"allow_merge_on_skipped_pipeline,omitempty"`
	OnlyAllowMergeIfAllStatusChecksPassed     *bool                                `url:"only_allow_merge_if_all_status_checks_passed,omitempty" json:"only_allow_merge_if_all_status_checks_passed,omitempty"`
	AnalyticsAccessLevel                      *AccessControlValue                  `url:"analytics_access_level,omitempty" json:"analytics_access_level,omitempty"`
	ApprovalsBeforeMerge                      *int                                 `url:"approvals_before_merge,omitempty" json:"approvals_before_merge,omitempty"`
	AutoCancelPendingPipelines                *string                              `url:"auto_cancel_pending_pipelines,omitempty" json:"auto_cancel_pending_pipelines,omitempty"`
	AutoDevopsDeployStrategy                  *string                              `url:"auto_devops_deploy_strategy,omitempty" json:"auto_devops_deploy_strategy,omitempty"`
	AutoDevopsEnabled                         *bool                                `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	AutocloseReferencedIssues                 *bool                                `url:"autoclose_referenced_issues,omitempty" json:"autoclose_referenced_issues,omitempty"`
	Avatar                                    *ProjectAvatar                       `url:"-" json:"-"`
	BuildCoverageRegex                        *string                              `url:"build_coverage_regex,omitempty" json:"build_coverage_regex,omitempty"`
	BuildGitStrategy                          *string                              `url:"build_git_strategy,omitempty" json:"build_git_strategy,omitempty"`
	BuildTimeout                              *int                                 `url:"build_timeout,omitempty" json:"build_timeout,omitempty"`
	BuildsAccessLevel                         *AccessControlValue                  `url:"builds_access_level,omitempty" json:"builds_access_level,omitempty"`
	CIConfigPath                              *string                              `url:"ci_config_path,omitempty" json:"ci_config_path,omitempty"`
	ContainerExpirationPolicyAttributes       *ContainerExpirationPolicyAttributes `url:"container_expiration_policy_attributes,omitempty" json:"container_expiration_policy_attributes,omitempty"`
	ContainerRegistryAccessLevel              *AccessControlValue                  `url:"container_registry_access_level,omitempty" json:"container_registry_access_level,omitempty"`
	DefaultBranch                             *string                              `url:"default_branch,omitempty" json:"default_branch,omitempty"`
	Description                               *string                              `url:"description,omitempty" json:"description,omitempty"`
	EmailsDisabled                            *bool                                `url:"emails_disabled,omitempty" json:"emails_disabled,omitempty"`
	EnforceAuthChecksOnUploads                *bool                                `url:"enforce_auth_checks_on_uploads,omitempty" json:"enforce_auth_checks_on_uploads,omitempty"`
	ExternalAuthorizationClassificationLabel  *string                              `url:"external_authorization_classification_label,omitempty" json:"external_authorization_classification_label,omitempty"`
	ForkingAccessLevel                        *AccessControlValue                  `url:"forking_access_level,omitempty" json:"forking_access_level,omitempty"`
	GroupWithProjectTemplatesID               *int                                 `url:"group_with_project_templates_id,omitempty" json:"group_with_project_templates_id,omitempty"`
	ImportURL                                 *string                              `url:"import_url,omitempty" json:"import_url,omitempty"`
	InitializeWithReadme                      *bool                                `url:"initialize_with_readme,omitempty" json:"initialize_with_readme,omitempty"`
	IssuesAccessLevel                         *AccessControlValue                  `url:"issues_access_level,omitempty" json:"issues_access_level,omitempty"`
	IssueBranchTemplate                       *string                              `url:"issue_branch_template,omitempty" json:"issue_branch_template,omitempty"`
	LFSEnabled                                *bool                                `url:"lfs_enabled,omitempty" json:"lfs_enabled,omitempty"`
	MergeCommitTemplate                       *string                              `url:"merge_commit_template,omitempty" json:"merge_commit_template,omitempty"`
	MergeMethod                               *MergeMethodValue                    `url:"merge_method,omitempty" json:"merge_method,omitempty"`
	MergePipelinesEnabled                     *bool                                `url:"merge_pipelines_enabled,omitempty" json:"merge_pipelines_enabled,omitempty"`
	MergeRequestsAccessLevel                  *AccessControlValue                  `url:"merge_requests_access_level,omitempty" json:"merge_requests_access_level,omitempty"`
	MergeTrainsEnabled                        *bool                                `url:"merge_trains_enabled,omitempty" json:"merge_trains_enabled,omitempty"`
	Mirror                                    *bool                                `url:"mirror,omitempty" json:"mirror,omitempty"`
	MirrorTriggerBuilds                       *bool                                `url:"mirror_trigger_builds,omitempty" json:"mirror_trigger_builds,omitempty"`
	Name                                      *string                              `url:"name,omitempty" json:"name,omitempty"`
	NamespaceID                               *int                                 `url:"namespace_id,omitempty" json:"namespace_id,omitempty"`
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool                                `url:"only_allow_merge_if_all_discussions_are_resolved,omitempty" json:"only_allow_merge_if_all_discussions_are_resolved,omitempty"`
	OnlyAllowMergeIfPipelineSucceeds          *bool                                `url:"only_allow_merge_if_pipeline_succeeds,omitempty" json:"only_allow_merge_if_pipeline_succeeds,omitempty"`
	OperationsAccessLevel                     *AccessControlValue                  `url:"operations_access_level,omitempty" json:"operations_access_level,omitempty"`
	PackagesEnabled                           *bool                                `url:"packages_enabled,omitempty" json:"packages_enabled,omitempty"`
	PagesAccessLevel                          *AccessControlValue                  `url:"pages_access_level,omitempty" json:"pages_access_level,omitempty"`
	Path                                      *string                              `url:"path,omitempty" json:"path,omitempty"`
	PublicBuilds                              *bool                                `url:"public_builds,omitempty" json:"public_builds,omitempty"`
	ReleasesAccessLevel                       *AccessControlValue                  `url:"releases_access_level,omitempty" json:"releases_access_level,omitempty"`
	EnvironmentsAccessLevel                   *AccessControlValue                  `url:"environments_access_level,omitempty" json:"environments_access_level,omitempty"`
	FeatureFlagsAccessLevel                   *AccessControlValue                  `url:"feature_flags_access_level,omitempty" json:"feature_flags_access_level,omitempty"`
	InfrastructureAccessLevel                 *AccessControlValue                  `url:"infrastructure_access_level,omitempty" json:"infrastructure_access_level,omitempty"`
	MonitorAccessLevel                        *AccessControlValue                  `url:"monitor_access_level,omitempty" json:"monitor_access_level,omitempty"`
	RemoveSourceBranchAfterMerge              *bool                                `url:"remove_source_branch_after_merge,omitempty" json:"remove_source_branch_after_merge,omitempty"`
	PrintingMergeRequestLinkEnabled           *bool                                `url:"printing_merge_request_link_enabled,omitempty" json:"printing_merge_request_link_enabled,omitempty"`
	RepositoryAccessLevel                     *AccessControlValue                  `url:"repository_access_level,omitempty" json:"repository_access_level,omitempty"`
	RepositoryStorage                         *string                              `url:"repository_storage,omitempty" json:"repository_storage,omitempty"`
	RequestAccessEnabled                      *bool                                `url:"request_access_enabled,omitempty" json:"request_access_enabled,omitempty"`
	RequirementsAccessLevel                   *AccessControlValue                  `url:"requirements_access_level,omitempty" json:"requirements_access_level,omitempty"`
	ResolveOutdatedDiffDiscussions            *bool                                `url:"resolve_outdated_diff_discussions,omitempty" json:"resolve_outdated_diff_discussions,omitempty"`
	SecurityAndComplianceAccessLevel          *AccessControlValue                  `url:"security_and_compliance_access_level,omitempty" json:"security_and_compliance_access_level,omitempty"`
	SharedRunnersEnabled                      *bool                                `url:"shared_runners_enabled,omitempty" json:"shared_runners_enabled,omitempty"`
	GroupRunnersEnabled                       *bool                                `url:"group_runners_enabled,omitempty" json:"group_runners_enabled,omitempty"`
	ShowDefaultAwardEmojis                    *bool                                `url:"show_default_award_emojis,omitempty" json:"show_default_award_emojis,omitempty"`
	SnippetsAccessLevel                       *AccessControlValue                  `url:"snippets_access_level,omitempty" json:"snippets_access_level,omitempty"`
	SquashCommitTemplate                      *string                              `url:"squash_commit_template,omitempty" json:"squash_commit_template,omitempty"`
	SquashOption                              *SquashOptionValue                   `url:"squash_option,omitempty" json:"squash_option,omitempty"`
	SuggestionCommitMessage                   *string                              `url:"suggestion_commit_message,omitempty" json:"suggestion_commit_message,omitempty"`
	TemplateName                              *string                              `url:"template_name,omitempty" json:"template_name,omitempty"`
	TemplateProjectID                         *int                                 `url:"template_project_id,omitempty" json:"template_project_id,omitempty"`
	Topics                                    *[]string                            `url:"topics,omitempty" json:"topics,omitempty"`
	UseCustomTemplate                         *bool                                `url:"use_custom_template,omitempty" json:"use_custom_template,omitempty"`
	Visibility                                *VisibilityValue                     `url:"visibility,omitempty" json:"visibility,omitempty"`
	WikiAccessLevel                           *AccessControlValue                  `url:"wiki_access_level,omitempty" json:"wiki_access_level,omitempty"`

	// Deprecated: No longer supported in recent versions.
	CIForwardDeploymentEnabled *bool `url:"ci_forward_deployment_enabled,omitempty" json:"ci_forward_deployment_enabled,omitempty"`
	// Deprecated: Use ContainerRegistryAccessLevel instead.
	ContainerRegistryEnabled *bool `url:"container_registry_enabled,omitempty" json:"container_registry_enabled,omitempty"`
	// Deprecated: Use IssuesAccessLevel instead.
	IssuesEnabled *bool `url:"issues_enabled,omitempty" json:"issues_enabled,omitempty"`
	// Deprecated: No longer supported in recent versions.
	IssuesTemplate *string `url:"issues_template,omitempty" json:"issues_template,omitempty"`
	// Deprecated: Use BuildsAccessLevel instead.
	JobsEnabled *bool `url:"jobs_enabled,omitempty" json:"jobs_enabled,omitempty"`
	// Deprecated: Use MergeRequestsAccessLevel instead.
	MergeRequestsEnabled *bool `url:"merge_requests_enabled,omitempty" json:"merge_requests_enabled,omitempty"`
	// Deprecated: No longer supported in recent versions.
	MergeRequestsTemplate *string `url:"merge_requests_template,omitempty" json:"merge_requests_template,omitempty"`
	// Deprecated: No longer supported in recent versions.
	ServiceDeskEnabled *bool `url:"service_desk_enabled,omitempty" json:"service_desk_enabled,omitempty"`
	// Deprecated: Use SnippetsAccessLevel instead.
	SnippetsEnabled *bool `url:"snippets_enabled,omitempty" json:"snippets_enabled,omitempty"`
	// Deprecated: Use Topics instead. (Deprecated in GitLab 14.0)
	TagList *[]string `url:"tag_list,omitempty" json:"tag_list,omitempty"`
	// Deprecated: Use WikiAccessLevel instead.
	WikiEnabled *bool `url:"wiki_enabled,omitempty" json:"wiki_enabled,omitempty"`
}

// ContainerExpirationPolicyAttributes represents the available container
// expiration policy attributes.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
type ContainerExpirationPolicyAttributes struct {
	Cadence         *string `url:"cadence,omitempty" json:"cadence,omitempty"`
	KeepN           *int    `url:"keep_n,omitempty" json:"keep_n,omitempty"`
	OlderThan       *string `url:"older_than,omitempty" json:"older_than,omitempty"`
	NameRegexDelete *string `url:"name_regex_delete,omitempty" json:"name_regex_delete,omitempty"`
	NameRegexKeep   *string `url:"name_regex_keep,omitempty" json:"name_regex_keep,omitempty"`
	Enabled         *bool   `url:"enabled,omitempty" json:"enabled,omitempty"`

	// Deprecated: Is replaced by NameRegexDelete and is internally hardwired to its value.
	NameRegex *string `url:"name_regex,omitempty" json:"name_regex,omitempty"`
}

// ProjectAvatar represents a GitLab project avatar.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
type ProjectAvatar struct {
	Filename string
	Image    io.Reader
}

// MarshalJSON implements the json.Marshaler interface.
func (a *ProjectAvatar) MarshalJSON() ([]byte, error) {
	if a.Filename == "" && a.Image == nil {
		return []byte(`""`), nil
	}
	type alias ProjectAvatar
	return json.Marshal((*alias)(a))
}

// CreateProject creates a new project owned by the authenticated user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#create-project
func (s *ProjectsService) CreateProject(opt *CreateProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	if opt.ContainerExpirationPolicyAttributes != nil {
		// This is needed to satisfy the API. Should be deleted
		// when NameRegex is removed (it's now deprecated).
		opt.ContainerExpirationPolicyAttributes.NameRegex = opt.ContainerExpirationPolicyAttributes.NameRegexDelete
	}

	var err error
	var req *retryablehttp.Request

	if opt.Avatar == nil {
		req, err = s.client.NewRequest(http.MethodPost, "projects", opt, options)
	} else {
		req, err = s.client.UploadRequest(
			http.MethodPost,
			"projects",
			opt.Avatar.Image,
			opt.Avatar.Filename,
			UploadAvatar,
			opt,
			options,
		)
	}
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// CreateProjectForUserOptions represents the available CreateProjectForUser()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#create-project-for-user
type CreateProjectForUserOptions CreateProjectOptions

// CreateProjectForUser creates a new project owned by the specified user.
// Available only for admins.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#create-project-for-user
func (s *ProjectsService) CreateProjectForUser(user int, opt *CreateProjectForUserOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	if opt.ContainerExpirationPolicyAttributes != nil {
		// This is needed to satisfy the API. Should be deleted
		// when NameRegex is removed (it's now deprecated).
		opt.ContainerExpirationPolicyAttributes.NameRegex = opt.ContainerExpirationPolicyAttributes.NameRegexDelete
	}

	var err error
	var req *retryablehttp.Request
	u := fmt.Sprintf("projects/user/%d", user)

	if opt.Avatar == nil {
		req, err = s.client.NewRequest(http.MethodPost, u, opt, options)
	} else {
		req, err = s.client.UploadRequest(
			http.MethodPost,
			u,
			opt.Avatar.Image,
			opt.Avatar.Filename,
			UploadAvatar,
			opt,
			options,
		)
	}
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// EditProjectOptions represents the available EditProject() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#edit-project
type EditProjectOptions struct {
	AllowMergeOnSkippedPipeline               *bool                                `url:"allow_merge_on_skipped_pipeline,omitempty" json:"allow_merge_on_skipped_pipeline,omitempty"`
	OnlyAllowMergeIfAllStatusChecksPassed     *bool                                `url:"only_allow_merge_if_all_status_checks_passed,omitempty" json:"only_allow_merge_if_all_status_checks_passed,omitempty"`
	AnalyticsAccessLevel                      *AccessControlValue                  `url:"analytics_access_level,omitempty" json:"analytics_access_level,omitempty"`
	ApprovalsBeforeMerge                      *int                                 `url:"approvals_before_merge,omitempty" json:"approvals_before_merge,omitempty"`
	AutoCancelPendingPipelines                *string                              `url:"auto_cancel_pending_pipelines,omitempty" json:"auto_cancel_pending_pipelines,omitempty"`
	AutoDevopsDeployStrategy                  *string                              `url:"auto_devops_deploy_strategy,omitempty" json:"auto_devops_deploy_strategy,omitempty"`
	AutoDevopsEnabled                         *bool                                `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	AutocloseReferencedIssues                 *bool                                `url:"autoclose_referenced_issues,omitempty" json:"autoclose_referenced_issues,omitempty"`
	Avatar                                    *ProjectAvatar                       `url:"-" json:"avatar,omitempty"`
	BuildCoverageRegex                        *string                              `url:"build_coverage_regex,omitempty" json:"build_coverage_regex,omitempty"`
	BuildGitStrategy                          *string                              `url:"build_git_strategy,omitempty" json:"build_git_strategy,omitempty"`
	BuildTimeout                              *int                                 `url:"build_timeout,omitempty" json:"build_timeout,omitempty"`
	BuildsAccessLevel                         *AccessControlValue                  `url:"builds_access_level,omitempty" json:"builds_access_level,omitempty"`
	CIConfigPath                              *string                              `url:"ci_config_path,omitempty" json:"ci_config_path,omitempty"`
	CIDefaultGitDepth                         *int                                 `url:"ci_default_git_depth,omitempty" json:"ci_default_git_depth,omitempty"`
	CIForwardDeploymentEnabled                *bool                                `url:"ci_forward_deployment_enabled,omitempty" json:"ci_forward_deployment_enabled,omitempty"`
	CISeperateCache                           *bool                                `url:"ci_separated_caches,omitempty" json:"ci_separated_caches,omitempty"`
	ContainerExpirationPolicyAttributes       *ContainerExpirationPolicyAttributes `url:"container_expiration_policy_attributes,omitempty" json:"container_expiration_policy_attributes,omitempty"`
	ContainerRegistryAccessLevel              *AccessControlValue                  `url:"container_registry_access_level,omitempty" json:"container_registry_access_level,omitempty"`
	DefaultBranch                             *string                              `url:"default_branch,omitempty" json:"default_branch,omitempty"`
	Description                               *string                              `url:"description,omitempty" json:"description,omitempty"`
	EmailsDisabled                            *bool                                `url:"emails_disabled,omitempty" json:"emails_disabled,omitempty"`
	EnforceAuthChecksOnUploads                *bool                                `url:"enforce_auth_checks_on_uploads,omitempty" json:"enforce_auth_checks_on_uploads,omitempty"`
	ExternalAuthorizationClassificationLabel  *string                              `url:"external_authorization_classification_label,omitempty" json:"external_authorization_classification_label,omitempty"`
	ForkingAccessLevel                        *AccessControlValue                  `url:"forking_access_level,omitempty" json:"forking_access_level,omitempty"`
	ImportURL                                 *string                              `url:"import_url,omitempty" json:"import_url,omitempty"`
	IssuesAccessLevel                         *AccessControlValue                  `url:"issues_access_level,omitempty" json:"issues_access_level,omitempty"`
	IssueBranchTemplate                       *string                              `url:"issue_branch_template,omitempty" json:"issue_branch_template,omitempty"`
	IssuesTemplate                            *string                              `url:"issues_template,omitempty" json:"issues_template,omitempty"`
	KeepLatestArtifact                        *bool                                `url:"keep_latest_artifact,omitempty" json:"keep_latest_artifact,omitempty"`
	LFSEnabled                                *bool                                `url:"lfs_enabled,omitempty" json:"lfs_enabled,omitempty"`
	MergeCommitTemplate                       *string                              `url:"merge_commit_template,omitempty" json:"merge_commit_template,omitempty"`
	MergeRequestDefaultTargetSelf             *bool                                `url:"mr_default_target_self,omitempty" json:"mr_default_target_self,omitempty"`
	MergeMethod                               *MergeMethodValue                    `url:"merge_method,omitempty" json:"merge_method,omitempty"`
	MergePipelinesEnabled                     *bool                                `url:"merge_pipelines_enabled,omitempty" json:"merge_pipelines_enabled,omitempty"`
	MergeRequestsAccessLevel                  *AccessControlValue                  `url:"merge_requests_access_level,omitempty" json:"merge_requests_access_level,omitempty"`
	MergeRequestsTemplate                     *string                              `url:"merge_requests_template,omitempty" json:"merge_requests_template,omitempty"`
	MergeTrainsEnabled                        *bool                                `url:"merge_trains_enabled,omitempty" json:"merge_trains_enabled,omitempty"`
	Mirror                                    *bool                                `url:"mirror,omitempty" json:"mirror,omitempty"`
	MirrorOverwritesDivergedBranches          *bool                                `url:"mirror_overwrites_diverged_branches,omitempty" json:"mirror_overwrites_diverged_branches,omitempty"`
	MirrorTriggerBuilds                       *bool                                `url:"mirror_trigger_builds,omitempty" json:"mirror_trigger_builds,omitempty"`
	MirrorUserID                              *int                                 `url:"mirror_user_id,omitempty" json:"mirror_user_id,omitempty"`
	Name                                      *string                              `url:"name,omitempty" json:"name,omitempty"`
	OnlyAllowMergeIfAllDiscussionsAreResolved *bool                                `url:"only_allow_merge_if_all_discussions_are_resolved,omitempty" json:"only_allow_merge_if_all_discussions_are_resolved,omitempty"`
	OnlyAllowMergeIfPipelineSucceeds          *bool                                `url:"only_allow_merge_if_pipeline_succeeds,omitempty" json:"only_allow_merge_if_pipeline_succeeds,omitempty"`
	OnlyMirrorProtectedBranches               *bool                                `url:"only_mirror_protected_branches,omitempty" json:"only_mirror_protected_branches,omitempty"`
	OperationsAccessLevel                     *AccessControlValue                  `url:"operations_access_level,omitempty" json:"operations_access_level,omitempty"`
	PackagesEnabled                           *bool                                `url:"packages_enabled,omitempty" json:"packages_enabled,omitempty"`
	PagesAccessLevel                          *AccessControlValue                  `url:"pages_access_level,omitempty" json:"pages_access_level,omitempty"`
	Path                                      *string                              `url:"path,omitempty" json:"path,omitempty"`
	PublicBuilds                              *bool                                `url:"public_builds,omitempty" json:"public_builds,omitempty"`
	ReleasesAccessLevel                       *AccessControlValue                  `url:"releases_access_level,omitempty" json:"releases_access_level,omitempty"`
	EnvironmentsAccessLevel                   *AccessControlValue                  `url:"environments_access_level,omitempty" json:"environments_access_level,omitempty"`
	FeatureFlagsAccessLevel                   *AccessControlValue                  `url:"feature_flags_access_level,omitempty" json:"feature_flags_access_level,omitempty"`
	InfrastructureAccessLevel                 *AccessControlValue                  `url:"infrastructure_access_level,omitempty" json:"infrastructure_access_level,omitempty"`
	MonitorAccessLevel                        *AccessControlValue                  `url:"monitor_access_level,omitempty" json:"monitor_access_level,omitempty"`
	RemoveSourceBranchAfterMerge              *bool                                `url:"remove_source_branch_after_merge,omitempty" json:"remove_source_branch_after_merge,omitempty"`
	PrintingMergeRequestLinkEnabled           *bool                                `url:"printing_merge_request_link_enabled,omitempty" json:"printing_merge_request_link_enabled,omitempty"`
	RepositoryAccessLevel                     *AccessControlValue                  `url:"repository_access_level,omitempty" json:"repository_access_level,omitempty"`
	RepositoryStorage                         *string                              `url:"repository_storage,omitempty" json:"repository_storage,omitempty"`
	RequestAccessEnabled                      *bool                                `url:"request_access_enabled,omitempty" json:"request_access_enabled,omitempty"`
	RequirementsAccessLevel                   *AccessControlValue                  `url:"requirements_access_level,omitempty" json:"requirements_access_level,omitempty"`
	ResolveOutdatedDiffDiscussions            *bool                                `url:"resolve_outdated_diff_discussions,omitempty" json:"resolve_outdated_diff_discussions,omitempty"`
	RestrictUserDefinedVariables              *bool                                `url:"restrict_user_defined_variables,omitempty" json:"restrict_user_defined_variables,omitempty"`
	SecurityAndComplianceAccessLevel          *AccessControlValue                  `url:"security_and_compliance_access_level,omitempty" json:"security_and_compliance_access_level,omitempty"`
	ServiceDeskEnabled                        *bool                                `url:"service_desk_enabled,omitempty" json:"service_desk_enabled,omitempty"`
	SharedRunnersEnabled                      *bool                                `url:"shared_runners_enabled,omitempty" json:"shared_runners_enabled,omitempty"`
	GroupRunnersEnabled                       *bool                                `url:"group_runners_enabled,omitempty" json:"group_runners_enabled,omitempty"`
	ShowDefaultAwardEmojis                    *bool                                `url:"show_default_award_emojis,omitempty" json:"show_default_award_emojis,omitempty"`
	SnippetsAccessLevel                       *AccessControlValue                  `url:"snippets_access_level,omitempty" json:"snippets_access_level,omitempty"`
	SquashCommitTemplate                      *string                              `url:"squash_commit_template,omitempty" json:"squash_commit_template,omitempty"`
	SquashOption                              *SquashOptionValue                   `url:"squash_option,omitempty" json:"squash_option,omitempty"`
	SuggestionCommitMessage                   *string                              `url:"suggestion_commit_message,omitempty" json:"suggestion_commit_message,omitempty"`
	Topics                                    *[]string                            `url:"topics,omitempty" json:"topics,omitempty"`
	Visibility                                *VisibilityValue                     `url:"visibility,omitempty" json:"visibility,omitempty"`
	WikiAccessLevel                           *AccessControlValue                  `url:"wiki_access_level,omitempty" json:"wiki_access_level,omitempty"`

	// Deprecated: Use ContainerRegistryAccessLevel instead.
	ContainerRegistryEnabled *bool `url:"container_registry_enabled,omitempty" json:"container_registry_enabled,omitempty"`
	// Deprecated: Use IssuesAccessLevel instead.
	IssuesEnabled *bool `url:"issues_enabled,omitempty" json:"issues_enabled,omitempty"`
	// Deprecated: Use BuildsAccessLevel instead.
	JobsEnabled *bool `url:"jobs_enabled,omitempty" json:"jobs_enabled,omitempty"`
	// Deprecated: Use MergeRequestsAccessLevel instead.
	MergeRequestsEnabled *bool `url:"merge_requests_enabled,omitempty" json:"merge_requests_enabled,omitempty"`
	// Deprecated: Use SnippetsAccessLevel instead.
	SnippetsEnabled *bool `url:"snippets_enabled,omitempty" json:"snippets_enabled,omitempty"`
	// Deprecated: Use Topics instead. (Deprecated in GitLab 14.0)
	TagList *[]string `url:"tag_list,omitempty" json:"tag_list,omitempty"`
	// Deprecated: Use WikiAccessLevel instead.
	WikiEnabled *bool `url:"wiki_enabled,omitempty" json:"wiki_enabled,omitempty"`
}

// EditProject updates an existing project.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#edit-project
func (s *ProjectsService) EditProject(pid interface{}, opt *EditProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	if opt.ContainerExpirationPolicyAttributes != nil {
		// This is needed to satisfy the API. Should be deleted
		// when NameRegex is removed (it's now deprecated).
		opt.ContainerExpirationPolicyAttributes.NameRegex = opt.ContainerExpirationPolicyAttributes.NameRegexDelete
	}

	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	var req *retryablehttp.Request

	if opt.Avatar == nil || (opt.Avatar.Filename == "" && opt.Avatar.Image == nil) {
		req, err = s.client.NewRequest(http.MethodPut, u, opt, options)
	} else {
		req, err = s.client.UploadRequest(
			http.MethodPut,
			u,
			opt.Avatar.Image,
			opt.Avatar.Filename,
			UploadAvatar,
			opt,
			options,
		)
	}
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ForkProjectOptions represents the available ForkProject() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#fork-project
type ForkProjectOptions struct {
	Description                   *string          `url:"description,omitempty" json:"description,omitempty"`
	MergeRequestDefaultTargetSelf *bool            `url:"mr_default_target_self,omitempty" json:"mr_default_target_self,omitempty"`
	Name                          *string          `url:"name,omitempty" json:"name,omitempty"`
	NamespaceID                   *int             `url:"namespace_id,omitempty" json:"namespace_id,omitempty"`
	NamespacePath                 *string          `url:"namespace_path,omitempty" json:"namespace_path,omitempty"`
	Path                          *string          `url:"path,omitempty" json:"path,omitempty"`
	Visibility                    *VisibilityValue `url:"visibility,omitempty" json:"visibility,omitempty"`

	// Deprecated: This parameter has been split into NamespaceID and NamespacePath.
	Namespace *string `url:"namespace,omitempty" json:"namespace,omitempty"`
}

// ForkProject forks a project into the user namespace of the authenticated
// user.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#fork-project
func (s *ProjectsService) ForkProject(pid interface{}, opt *ForkProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/fork", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// StarProject stars a given the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#star-a-project
func (s *ProjectsService) StarProject(pid interface{}, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/star", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// UnstarProject unstars a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#unstar-a-project
func (s *ProjectsService) UnstarProject(pid interface{}, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/unstar", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ArchiveProject archives the project if the user is either admin or the
// project owner of this project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#archive-a-project
func (s *ProjectsService) ArchiveProject(pid interface{}, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/archive", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// UnarchiveProject unarchives the project if the user is either admin or
// the project owner of this project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#unarchive-a-project
func (s *ProjectsService) UnarchiveProject(pid interface{}, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/unarchive", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// DeleteProject removes a project including all associated resources
// (issues, merge requests etc.)
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#delete-project
func (s *ProjectsService) DeleteProject(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ShareWithGroupOptions represents options to share project with groups
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#share-project-with-group
type ShareWithGroupOptions struct {
	ExpiresAt   *string           `url:"expires_at" json:"expires_at"`
	GroupAccess *AccessLevelValue `url:"group_access" json:"group_access"`
	GroupID     *int              `url:"group_id" json:"group_id"`
}

// ShareProjectWithGroup allows to share a project with a group.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#share-project-with-group
func (s *ProjectsService) ShareProjectWithGroup(pid interface{}, opt *ShareWithGroupOptions, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/share", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteSharedProjectFromGroup allows to unshare a project from a group.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#delete-a-shared-project-link-within-a-group
func (s *ProjectsService) DeleteSharedProjectFromGroup(pid interface{}, groupID int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/share/%d", PathEscape(project), groupID)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ProjectMember represents a project member.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/members.html#list-all-members-of-a-group-or-project
type ProjectMember struct {
	ID          int              `json:"id"`
	Username    string           `json:"username"`
	Email       string           `json:"email"`
	Name        string           `json:"name"`
	State       string           `json:"state"`
	CreatedAt   *time.Time       `json:"created_at"`
	ExpiresAt   *ISOTime         `json:"expires_at"`
	AccessLevel AccessLevelValue `json:"access_level"`
	WebURL      string           `json:"web_url"`
	AvatarURL   string           `json:"avatar_url"`
}

// ProjectHook represents a project hook.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-project-hooks
type ProjectHook struct {
	ID                       int        `json:"id"`
	URL                      string     `json:"url"`
	ConfidentialNoteEvents   bool       `json:"confidential_note_events"`
	ProjectID                int        `json:"project_id"`
	PushEvents               bool       `json:"push_events"`
	PushEventsBranchFilter   string     `json:"push_events_branch_filter"`
	IssuesEvents             bool       `json:"issues_events"`
	ConfidentialIssuesEvents bool       `json:"confidential_issues_events"`
	MergeRequestsEvents      bool       `json:"merge_requests_events"`
	TagPushEvents            bool       `json:"tag_push_events"`
	NoteEvents               bool       `json:"note_events"`
	JobEvents                bool       `json:"job_events"`
	PipelineEvents           bool       `json:"pipeline_events"`
	WikiPageEvents           bool       `json:"wiki_page_events"`
	DeploymentEvents         bool       `json:"deployment_events"`
	ReleasesEvents           bool       `json:"releases_events"`
	EnableSSLVerification    bool       `json:"enable_ssl_verification"`
	CreatedAt                *time.Time `json:"created_at"`
}

// ListProjectHooksOptions represents the available ListProjectHooks() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#list-project-hooks
type ListProjectHooksOptions ListOptions

// ListProjectHooks gets a list of project hooks.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-project-hooks
func (s *ProjectsService) ListProjectHooks(pid interface{}, opt *ListProjectHooksOptions, options ...RequestOptionFunc) ([]*ProjectHook, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/hooks", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ph []*ProjectHook
	resp, err := s.client.Do(req, &ph)
	if err != nil {
		return nil, resp, err
	}

	return ph, resp, nil
}

// GetProjectHook gets a specific hook for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-project-hook
func (s *ProjectsService) GetProjectHook(pid interface{}, hook int, options ...RequestOptionFunc) (*ProjectHook, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/hooks/%d", PathEscape(project), hook)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ph := new(ProjectHook)
	resp, err := s.client.Do(req, ph)
	if err != nil {
		return nil, resp, err
	}

	return ph, resp, nil
}

// AddProjectHookOptions represents the available AddProjectHook() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#add-project-hook
type AddProjectHookOptions struct {
	ConfidentialIssuesEvents *bool   `url:"confidential_issues_events,omitempty" json:"confidential_issues_events,omitempty"`
	ConfidentialNoteEvents   *bool   `url:"confidential_note_events,omitempty" json:"confidential_note_events,omitempty"`
	DeploymentEvents         *bool   `url:"deployment_events,omitempty" json:"deployment_events,omitempty"`
	EnableSSLVerification    *bool   `url:"enable_ssl_verification,omitempty" json:"enable_ssl_verification,omitempty"`
	IssuesEvents             *bool   `url:"issues_events,omitempty" json:"issues_events,omitempty"`
	JobEvents                *bool   `url:"job_events,omitempty" json:"job_events,omitempty"`
	MergeRequestsEvents      *bool   `url:"merge_requests_events,omitempty" json:"merge_requests_events,omitempty"`
	NoteEvents               *bool   `url:"note_events,omitempty" json:"note_events,omitempty"`
	PipelineEvents           *bool   `url:"pipeline_events,omitempty" json:"pipeline_events,omitempty"`
	PushEvents               *bool   `url:"push_events,omitempty" json:"push_events,omitempty"`
	PushEventsBranchFilter   *string `url:"push_events_branch_filter,omitempty" json:"push_events_branch_filter,omitempty"`
	ReleasesEvents           *bool   `url:"releases_events,omitempty" json:"releases_events,omitempty"`
	TagPushEvents            *bool   `url:"tag_push_events,omitempty" json:"tag_push_events,omitempty"`
	Token                    *string `url:"token,omitempty" json:"token,omitempty"`
	URL                      *string `url:"url,omitempty" json:"url,omitempty"`
	WikiPageEvents           *bool   `url:"wiki_page_events,omitempty" json:"wiki_page_events,omitempty"`
}

// AddProjectHook adds a hook to a specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#add-project-hook
func (s *ProjectsService) AddProjectHook(pid interface{}, opt *AddProjectHookOptions, options ...RequestOptionFunc) (*ProjectHook, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/hooks", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	ph := new(ProjectHook)
	resp, err := s.client.Do(req, ph)
	if err != nil {
		return nil, resp, err
	}

	return ph, resp, nil
}

// EditProjectHookOptions represents the available EditProjectHook() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#edit-project-hook
type EditProjectHookOptions struct {
	ConfidentialIssuesEvents *bool   `url:"confidential_issues_events,omitempty" json:"confidential_issues_events,omitempty"`
	ConfidentialNoteEvents   *bool   `url:"confidential_note_events,omitempty" json:"confidential_note_events,omitempty"`
	DeploymentEvents         *bool   `url:"deployment_events,omitempty" json:"deployment_events,omitempty"`
	EnableSSLVerification    *bool   `url:"enable_ssl_verification,omitempty" json:"enable_ssl_verification,omitempty"`
	IssuesEvents             *bool   `url:"issues_events,omitempty" json:"issues_events,omitempty"`
	JobEvents                *bool   `url:"job_events,omitempty" json:"job_events,omitempty"`
	MergeRequestsEvents      *bool   `url:"merge_requests_events,omitempty" json:"merge_requests_events,omitempty"`
	NoteEvents               *bool   `url:"note_events,omitempty" json:"note_events,omitempty"`
	PipelineEvents           *bool   `url:"pipeline_events,omitempty" json:"pipeline_events,omitempty"`
	PushEvents               *bool   `url:"push_events,omitempty" json:"push_events,omitempty"`
	PushEventsBranchFilter   *string `url:"push_events_branch_filter,omitempty" json:"push_events_branch_filter,omitempty"`
	ReleasesEvents           *bool   `url:"releases_events,omitempty" json:"releases_events,omitempty"`
	TagPushEvents            *bool   `url:"tag_push_events,omitempty" json:"tag_push_events,omitempty"`
	Token                    *string `url:"token,omitempty" json:"token,omitempty"`
	URL                      *string `url:"url,omitempty" json:"url,omitempty"`
	WikiPageEvents           *bool   `url:"wiki_page_events,omitempty" json:"wiki_page_events,omitempty"`
}

// EditProjectHook edits a hook for a specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#edit-project-hook
func (s *ProjectsService) EditProjectHook(pid interface{}, hook int, opt *EditProjectHookOptions, options ...RequestOptionFunc) (*ProjectHook, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/hooks/%d", PathEscape(project), hook)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	ph := new(ProjectHook)
	resp, err := s.client.Do(req, ph)
	if err != nil {
		return nil, resp, err
	}

	return ph, resp, nil
}

// DeleteProjectHook removes a hook from a project. This is an idempotent
// method and can be called multiple times. Either the hook is available or not.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#delete-project-hook
func (s *ProjectsService) DeleteProjectHook(pid interface{}, hook int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/hooks/%d", PathEscape(project), hook)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ProjectForkRelation represents a project fork relationship.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#admin-fork-relation
type ProjectForkRelation struct {
	ID                  int        `json:"id"`
	ForkedToProjectID   int        `json:"forked_to_project_id"`
	ForkedFromProjectID int        `json:"forked_from_project_id"`
	CreatedAt           *time.Time `json:"created_at"`
	UpdatedAt           *time.Time `json:"updated_at"`
}

// CreateProjectForkRelation creates a forked from/to relation between
// existing projects.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#create-a-forked-fromto-relation-between-existing-projects.
func (s *ProjectsService) CreateProjectForkRelation(pid interface{}, fork int, options ...RequestOptionFunc) (*ProjectForkRelation, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/fork/%d", PathEscape(project), fork)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pfr := new(ProjectForkRelation)
	resp, err := s.client.Do(req, pfr)
	if err != nil {
		return nil, resp, err
	}

	return pfr, resp, nil
}

// DeleteProjectForkRelation deletes an existing forked from relationship.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#delete-an-existing-forked-from-relationship
func (s *ProjectsService) DeleteProjectForkRelation(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/fork", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ProjectFile represents an uploaded project file.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#upload-a-file
type ProjectFile struct {
	Alt      string `json:"alt"`
	URL      string `json:"url"`
	Markdown string `json:"markdown"`
}

// UploadFile uploads a file.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#upload-a-file
func (s *ProjectsService) UploadFile(pid interface{}, content io.Reader, filename string, options ...RequestOptionFunc) (*ProjectFile, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/uploads", PathEscape(project))

	req, err := s.client.UploadRequest(
		http.MethodPost,
		u,
		content,
		filename,
		UploadFile,
		nil,
		options,
	)
	if err != nil {
		return nil, nil, err
	}

	pf := new(ProjectFile)
	resp, err := s.client.Do(req, pf)
	if err != nil {
		return nil, resp, err
	}

	return pf, resp, nil
}

// UploadAvatar uploads an avatar.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#upload-a-project-avatar
func (s *ProjectsService) UploadAvatar(pid interface{}, avatar io.Reader, filename string, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s", PathEscape(project))

	req, err := s.client.UploadRequest(
		http.MethodPut,
		u,
		avatar,
		filename,
		UploadAvatar,
		nil,
		options,
	)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// ListProjectForks gets a list of project forks.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#list-forks-of-a-project
func (s *ProjectsService) ListProjectForks(pid interface{}, opt *ListProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/forks", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var forks []*Project
	resp, err := s.client.Do(req, &forks)
	if err != nil {
		return nil, resp, err
	}

	return forks, resp, nil
}

// ProjectPushRules represents a project push rule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#push-rules
type ProjectPushRules struct {
	ID                         int        `json:"id"`
	ProjectID                  int        `json:"project_id"`
	CommitMessageRegex         string     `json:"commit_message_regex"`
	CommitMessageNegativeRegex string     `json:"commit_message_negative_regex"`
	BranchNameRegex            string     `json:"branch_name_regex"`
	DenyDeleteTag              bool       `json:"deny_delete_tag"`
	CreatedAt                  *time.Time `json:"created_at"`
	MemberCheck                bool       `json:"member_check"`
	PreventSecrets             bool       `json:"prevent_secrets"`
	AuthorEmailRegex           string     `json:"author_email_regex"`
	FileNameRegex              string     `json:"file_name_regex"`
	MaxFileSize                int        `json:"max_file_size"`
	CommitCommitterCheck       bool       `json:"commit_committer_check"`
	RejectUnsignedCommits      bool       `json:"reject_unsigned_commits"`
}

// GetProjectPushRules gets the push rules of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-project-push-rules
func (s *ProjectsService) GetProjectPushRules(pid interface{}, options ...RequestOptionFunc) (*ProjectPushRules, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/push_rule", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ppr := new(ProjectPushRules)
	resp, err := s.client.Do(req, ppr)
	if err != nil {
		return nil, resp, err
	}

	return ppr, resp, nil
}

// AddProjectPushRuleOptions represents the available AddProjectPushRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#add-project-push-rule
type AddProjectPushRuleOptions struct {
	AuthorEmailRegex           *string `url:"author_email_regex,omitempty" json:"author_email_regex,omitempty"`
	BranchNameRegex            *string `url:"branch_name_regex,omitempty" json:"branch_name_regex,omitempty"`
	CommitCommitterCheck       *bool   `url:"commit_committer_check,omitempty" json:"commit_committer_check,omitempty"`
	CommitMessageNegativeRegex *string `url:"commit_message_negative_regex,omitempty" json:"commit_message_negative_regex,omitempty"`
	CommitMessageRegex         *string `url:"commit_message_regex,omitempty" json:"commit_message_regex,omitempty"`
	DenyDeleteTag              *bool   `url:"deny_delete_tag,omitempty" json:"deny_delete_tag,omitempty"`
	FileNameRegex              *string `url:"file_name_regex,omitempty" json:"file_name_regex,omitempty"`
	MaxFileSize                *int    `url:"max_file_size,omitempty" json:"max_file_size,omitempty"`
	MemberCheck                *bool   `url:"member_check,omitempty" json:"member_check,omitempty"`
	PreventSecrets             *bool   `url:"prevent_secrets,omitempty" json:"prevent_secrets,omitempty"`
	RejectUnsignedCommits      *bool   `url:"reject_unsigned_commits,omitempty" json:"reject_unsigned_commits,omitempty"`
}

// AddProjectPushRule adds a push rule to a specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#add-project-push-rule
func (s *ProjectsService) AddProjectPushRule(pid interface{}, opt *AddProjectPushRuleOptions, options ...RequestOptionFunc) (*ProjectPushRules, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/push_rule", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	ppr := new(ProjectPushRules)
	resp, err := s.client.Do(req, ppr)
	if err != nil {
		return nil, resp, err
	}

	return ppr, resp, nil
}

// EditProjectPushRuleOptions represents the available EditProjectPushRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#edit-project-push-rule
type EditProjectPushRuleOptions struct {
	AuthorEmailRegex           *string `url:"author_email_regex,omitempty" json:"author_email_regex,omitempty"`
	BranchNameRegex            *string `url:"branch_name_regex,omitempty" json:"branch_name_regex,omitempty"`
	CommitCommitterCheck       *bool   `url:"commit_committer_check,omitempty" json:"commit_committer_check,omitempty"`
	CommitMessageNegativeRegex *string `url:"commit_message_negative_regex,omitempty" json:"commit_message_negative_regex,omitempty"`
	CommitMessageRegex         *string `url:"commit_message_regex,omitempty" json:"commit_message_regex,omitempty"`
	DenyDeleteTag              *bool   `url:"deny_delete_tag,omitempty" json:"deny_delete_tag,omitempty"`
	FileNameRegex              *string `url:"file_name_regex,omitempty" json:"file_name_regex,omitempty"`
	MaxFileSize                *int    `url:"max_file_size,omitempty" json:"max_file_size,omitempty"`
	MemberCheck                *bool   `url:"member_check,omitempty" json:"member_check,omitempty"`
	PreventSecrets             *bool   `url:"prevent_secrets,omitempty" json:"prevent_secrets,omitempty"`
	RejectUnsignedCommits      *bool   `url:"reject_unsigned_commits,omitempty" json:"reject_unsigned_commits,omitempty"`
}

// EditProjectPushRule edits a push rule for a specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#edit-project-push-rule
func (s *ProjectsService) EditProjectPushRule(pid interface{}, opt *EditProjectPushRuleOptions, options ...RequestOptionFunc) (*ProjectPushRules, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/push_rule", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	ppr := new(ProjectPushRules)
	resp, err := s.client.Do(req, ppr)
	if err != nil {
		return nil, resp, err
	}

	return ppr, resp, nil
}

// DeleteProjectPushRule removes a push rule from a project. This is an
// idempotent method and can be called multiple times. Either the push rule is
// available or not.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#delete-project-push-rule
func (s *ProjectsService) DeleteProjectPushRule(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/push_rule", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ProjectApprovals represents GitLab project level merge request approvals.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#project-level-mr-approvals
type ProjectApprovals struct {
	Approvers                                 []*MergeRequestApproverUser  `json:"approvers"`
	ApproverGroups                            []*MergeRequestApproverGroup `json:"approver_groups"`
	ApprovalsBeforeMerge                      int                          `json:"approvals_before_merge"`
	ResetApprovalsOnPush                      bool                         `json:"reset_approvals_on_push"`
	DisableOverridingApproversPerMergeRequest bool                         `json:"disable_overriding_approvers_per_merge_request"`
	MergeRequestsAuthorApproval               bool                         `json:"merge_requests_author_approval"`
	MergeRequestsDisableCommittersApproval    bool                         `json:"merge_requests_disable_committers_approval"`
	RequirePasswordToApprove                  bool                         `json:"require_password_to_approve"`
	SelectiveCodeOwnerRemovals                bool                         `json:"selective_code_owner_removals,omitempty"`
}

// GetApprovalConfiguration get the approval configuration for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#get-configuration
func (s *ProjectsService) GetApprovalConfiguration(pid interface{}, options ...RequestOptionFunc) (*ProjectApprovals, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approvals", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pa := new(ProjectApprovals)
	resp, err := s.client.Do(req, pa)
	if err != nil {
		return nil, resp, err
	}

	return pa, resp, nil
}

// ChangeApprovalConfigurationOptions represents the available
// ApprovalConfiguration() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#change-configuration
type ChangeApprovalConfigurationOptions struct {
	ApprovalsBeforeMerge                      *int  `url:"approvals_before_merge,omitempty" json:"approvals_before_merge,omitempty"`
	DisableOverridingApproversPerMergeRequest *bool `url:"disable_overriding_approvers_per_merge_request,omitempty" json:"disable_overriding_approvers_per_merge_request,omitempty"`
	MergeRequestsAuthorApproval               *bool `url:"merge_requests_author_approval,omitempty" json:"merge_requests_author_approval,omitempty"`
	MergeRequestsDisableCommittersApproval    *bool `url:"merge_requests_disable_committers_approval,omitempty" json:"merge_requests_disable_committers_approval,omitempty"`
	RequirePasswordToApprove                  *bool `url:"require_password_to_approve,omitempty" json:"require_password_to_approve,omitempty"`
	ResetApprovalsOnPush                      *bool `url:"reset_approvals_on_push,omitempty" json:"reset_approvals_on_push,omitempty"`
	SelectiveCodeOwnerRemovals                *bool `url:"selective_code_owner_removals,omitempty" json:"selective_code_owner_removals,omitempty"`
}

// ChangeApprovalConfiguration updates the approval configuration for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#change-configuration
func (s *ProjectsService) ChangeApprovalConfiguration(pid interface{}, opt *ChangeApprovalConfigurationOptions, options ...RequestOptionFunc) (*ProjectApprovals, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approvals", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pa := new(ProjectApprovals)
	resp, err := s.client.Do(req, pa)
	if err != nil {
		return nil, resp, err
	}

	return pa, resp, nil
}

// GetProjectApprovalRulesListsOptions represents the available GetProjectApprovalRules() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/merge_request_approvals.html#get-project-level-rules
type GetProjectApprovalRulesListsOptions ListOptions

// GetProjectApprovalRules looks up the list of project level approver rules.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#get-project-level-rules
func (s *ProjectsService) GetProjectApprovalRules(pid interface{}, opt *GetProjectApprovalRulesListsOptions, options ...RequestOptionFunc) ([]*ProjectApprovalRule, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approval_rules", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var par []*ProjectApprovalRule
	resp, err := s.client.Do(req, &par)
	if err != nil {
		return nil, resp, err
	}

	return par, resp, nil
}

// GetProjectApprovalRule gets the project level approvers.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#get-a-single-project-level-rule
func (s *ProjectsService) GetProjectApprovalRule(pid interface{}, ruleID int, options ...RequestOptionFunc) (*ProjectApprovalRule, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approval_rules/%d", PathEscape(project), ruleID)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	par := new(ProjectApprovalRule)
	resp, err := s.client.Do(req, &par)
	if err != nil {
		return nil, resp, err
	}

	return par, resp, nil
}

// CreateProjectLevelRuleOptions represents the available CreateProjectApprovalRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#create-project-level-rule
type CreateProjectLevelRuleOptions struct {
	Name                          *string `url:"name,omitempty" json:"name,omitempty"`
	ApprovalsRequired             *int    `url:"approvals_required,omitempty" json:"approvals_required,omitempty"`
	RuleType                      *string `url:"rule_type,omitempty" json:"rule_type,omitempty"`
	UserIDs                       *[]int  `url:"user_ids,omitempty" json:"user_ids,omitempty"`
	GroupIDs                      *[]int  `url:"group_ids,omitempty" json:"group_ids,omitempty"`
	ProtectedBranchIDs            *[]int  `url:"protected_branch_ids,omitempty" json:"protected_branch_ids,omitempty"`
	AppliesToAllProtectedBranches *bool   `url:"applies_to_all_protected_branches,omitempty" json:"applies_to_all_protected_branches,omitempty"`
}

// CreateProjectApprovalRule creates a new project-level approval rule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#create-project-level-rule
func (s *ProjectsService) CreateProjectApprovalRule(pid interface{}, opt *CreateProjectLevelRuleOptions, options ...RequestOptionFunc) (*ProjectApprovalRule, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approval_rules", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	par := new(ProjectApprovalRule)
	resp, err := s.client.Do(req, &par)
	if err != nil {
		return nil, resp, err
	}

	return par, resp, nil
}

// UpdateProjectLevelRuleOptions represents the available UpdateProjectApprovalRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#update-project-level-rule
type UpdateProjectLevelRuleOptions struct {
	Name                          *string `url:"name,omitempty" json:"name,omitempty"`
	ApprovalsRequired             *int    `url:"approvals_required,omitempty" json:"approvals_required,omitempty"`
	UserIDs                       *[]int  `url:"user_ids,omitempty" json:"user_ids,omitempty"`
	GroupIDs                      *[]int  `url:"group_ids,omitempty" json:"group_ids,omitempty"`
	ProtectedBranchIDs            *[]int  `url:"protected_branch_ids,omitempty" json:"protected_branch_ids,omitempty"`
	AppliesToAllProtectedBranches *bool   `url:"applies_to_all_protected_branches,omitempty" json:"applies_to_all_protected_branches,omitempty"`
}

// UpdateProjectApprovalRule updates an existing approval rule with new options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#update-project-level-rule
func (s *ProjectsService) UpdateProjectApprovalRule(pid interface{}, approvalRule int, opt *UpdateProjectLevelRuleOptions, options ...RequestOptionFunc) (*ProjectApprovalRule, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approval_rules/%d", PathEscape(project), approvalRule)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	par := new(ProjectApprovalRule)
	resp, err := s.client.Do(req, &par)
	if err != nil {
		return nil, resp, err
	}

	return par, resp, nil
}

// DeleteProjectApprovalRule deletes a project-level approval rule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#delete-project-level-rule
func (s *ProjectsService) DeleteProjectApprovalRule(pid interface{}, approvalRule int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/approval_rules/%d", PathEscape(project), approvalRule)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ChangeAllowedApproversOptions represents the available ChangeAllowedApprovers()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#change-allowed-approvers
type ChangeAllowedApproversOptions struct {
	ApproverGroupIDs *[]int `url:"approver_group_ids,omitempty" json:"approver_group_ids,omitempty"`
	ApproverIDs      *[]int `url:"approver_ids,omitempty" json:"approver_ids,omitempty"`
}

// ChangeAllowedApprovers updates the list of approvers and approver groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/merge_request_approvals.html#change-allowed-approvers
func (s *ProjectsService) ChangeAllowedApprovers(pid interface{}, opt *ChangeAllowedApproversOptions, options ...RequestOptionFunc) (*ProjectApprovals, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/approvers", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pa := new(ProjectApprovals)
	resp, err := s.client.Do(req, pa)
	if err != nil {
		return nil, resp, err
	}

	return pa, resp, nil
}

// ProjectPullMirrorDetails represent the details of the configuration pull
// mirror and its update status.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-a-projects-pull-mirror-details
type ProjectPullMirrorDetails struct {
	ID                     int        `json:"id"`
	LastError              string     `json:"last_error"`
	LastSuccessfulUpdateAt *time.Time `json:"last_successful_update_at"`
	LastUpdateAt           *time.Time `json:"last_update_at"`
	LastUpdateStartedAt    *time.Time `json:"last_update_started_at"`
	UpdateStatus           string     `json:"update_status"`
	URL                    string     `json:"url"`
}

// GetProjectPullMirrorDetails returns the pull mirror details.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-a-projects-pull-mirror-details
func (s *ProjectsService) GetProjectPullMirrorDetails(pid interface{}, options ...RequestOptionFunc) (*ProjectPullMirrorDetails, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/mirror/pull", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	pmd := new(ProjectPullMirrorDetails)
	resp, err := s.client.Do(req, pmd)
	if err != nil {
		return nil, resp, err
	}

	return pmd, resp, nil
}

// StartMirroringProject start the pull mirroring process for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#start-the-pull-mirroring-process-for-a-project
func (s *ProjectsService) StartMirroringProject(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/mirror/pull", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// TransferProjectOptions represents the available TransferProject() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#transfer-a-project-to-a-new-namespace
type TransferProjectOptions struct {
	Namespace interface{} `url:"namespace,omitempty" json:"namespace,omitempty"`
}

// TransferProject transfer a project into the specified namespace
//
// GitLab API docs: https://docs.gitlab.com/ee/api/projects.html#transfer-a-project-to-a-new-namespace
func (s *ProjectsService) TransferProject(pid interface{}, opt *TransferProjectOptions, options ...RequestOptionFunc) (*Project, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/transfer", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(Project)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// StartHousekeepingProject start the Housekeeping task for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#start-the-housekeeping-task-for-a-project
func (s *ProjectsService) StartHousekeepingProject(pid interface{}, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/housekeeping", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// GetRepositoryStorage Get the path to repository storage.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/projects.html#get-the-path-to-repository-storage
type ProjectReposityStorage struct {
	ProjectID         int        `json:"project_id"`
	DiskPath          string     `json:"disk_path"`
	CreatedAt         *time.Time `json:"created_at"`
	RepositoryStorage string     `json:"repository_storage"`
}

func (s *ProjectsService) GetRepositoryStorage(pid interface{}, options ...RequestOptionFunc) (*ProjectReposityStorage, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/storage", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	prs := new(ProjectReposityStorage)
	resp, err := s.client.Do(req, prs)
	if err != nil {
		return nil, resp, err
	}

	return prs, resp, nil
}
