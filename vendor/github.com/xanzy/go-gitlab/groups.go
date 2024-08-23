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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// GroupsService handles communication with the group related methods of
// the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html
type GroupsService struct {
	client *Client
}

// Group represents a GitLab group.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html
type Group struct {
	ID                      int                        `json:"id"`
	Name                    string                     `json:"name"`
	Path                    string                     `json:"path"`
	Description             string                     `json:"description"`
	MembershipLock          bool                       `json:"membership_lock"`
	Visibility              VisibilityValue            `json:"visibility"`
	LFSEnabled              bool                       `json:"lfs_enabled"`
	DefaultBranchProtection int                        `json:"default_branch_protection"`
	AvatarURL               string                     `json:"avatar_url"`
	WebURL                  string                     `json:"web_url"`
	RequestAccessEnabled    bool                       `json:"request_access_enabled"`
	FullName                string                     `json:"full_name"`
	FullPath                string                     `json:"full_path"`
	FileTemplateProjectID   int                        `json:"file_template_project_id"`
	ParentID                int                        `json:"parent_id"`
	Projects                []*Project                 `json:"projects"`
	Statistics              *Statistics                `json:"statistics"`
	CustomAttributes        []*CustomAttribute         `json:"custom_attributes"`
	ShareWithGroupLock      bool                       `json:"share_with_group_lock"`
	RequireTwoFactorAuth    bool                       `json:"require_two_factor_authentication"`
	TwoFactorGracePeriod    int                        `json:"two_factor_grace_period"`
	ProjectCreationLevel    ProjectCreationLevelValue  `json:"project_creation_level"`
	AutoDevopsEnabled       bool                       `json:"auto_devops_enabled"`
	SubGroupCreationLevel   SubGroupCreationLevelValue `json:"subgroup_creation_level"`
	EmailsDisabled          bool                       `json:"emails_disabled"`
	MentionsDisabled        bool                       `json:"mentions_disabled"`
	RunnersToken            string                     `json:"runners_token"`
	SharedProjects          []*Project                 `json:"shared_projects"`
	SharedRunnersEnabled    bool                       `json:"shared_runners_enabled"`
	SharedWithGroups        []struct {
		GroupID          int      `json:"group_id"`
		GroupName        string   `json:"group_name"`
		GroupFullPath    string   `json:"group_full_path"`
		GroupAccessLevel int      `json:"group_access_level"`
		ExpiresAt        *ISOTime `json:"expires_at"`
	} `json:"shared_with_groups"`
	LDAPCN                         string           `json:"ldap_cn"`
	LDAPAccess                     AccessLevelValue `json:"ldap_access"`
	LDAPGroupLinks                 []*LDAPGroupLink `json:"ldap_group_links"`
	SAMLGroupLinks                 []*SAMLGroupLink `json:"saml_group_links"`
	SharedRunnersMinutesLimit      int              `json:"shared_runners_minutes_limit"`
	ExtraSharedRunnersMinutesLimit int              `json:"extra_shared_runners_minutes_limit"`
	PreventForkingOutsideGroup     bool             `json:"prevent_forking_outside_group"`
	MarkedForDeletionOn            *ISOTime         `json:"marked_for_deletion_on"`
	CreatedAt                      *time.Time       `json:"created_at"`
	IPRestrictionRanges            string           `json:"ip_restriction_ranges"`
}

// GroupAvatar represents a GitLab group avatar.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html
type GroupAvatar struct {
	Filename string
	Image    io.Reader
}

// MarshalJSON implements the json.Marshaler interface.
func (a *GroupAvatar) MarshalJSON() ([]byte, error) {
	if a.Filename == "" && a.Image == nil {
		return []byte(`""`), nil
	}
	type alias GroupAvatar
	return json.Marshal((*alias)(a))
}

// LDAPGroupLink represents a GitLab LDAP group link.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#ldap-group-links
type LDAPGroupLink struct {
	CN          string           `json:"cn"`
	Filter      string           `json:"filter"`
	GroupAccess AccessLevelValue `json:"group_access"`
	Provider    string           `json:"provider"`
}

// SAMLGroupLink represents a GitLab SAML group link.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#saml-group-links
type SAMLGroupLink struct {
	Name        string           `json:"name"`
	AccessLevel AccessLevelValue `json:"access_level"`
}

// ListGroupsOptions represents the available ListGroups() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#list-groups
type ListGroupsOptions struct {
	ListOptions
	AllAvailable         *bool             `url:"all_available,omitempty" json:"all_available,omitempty"`
	MinAccessLevel       *AccessLevelValue `url:"min_access_level,omitempty" json:"min_access_level,omitempty"`
	OrderBy              *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Owned                *bool             `url:"owned,omitempty" json:"owned,omitempty"`
	Search               *string           `url:"search,omitempty" json:"search,omitempty"`
	SkipGroups           *[]int            `url:"skip_groups,omitempty" del:"," json:"skip_groups,omitempty"`
	Sort                 *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Statistics           *bool             `url:"statistics,omitempty" json:"statistics,omitempty"`
	TopLevelOnly         *bool             `url:"top_level_only,omitempty" json:"top_level_only,omitempty"`
	WithCustomAttributes *bool             `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
}

// ListGroups gets a list of groups (as user: my groups, as admin: all groups).
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-groups
func (s *GroupsService) ListGroups(opt *ListGroupsOptions, options ...RequestOptionFunc) ([]*Group, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "groups", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*Group
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// ListSubGroupsOptions represents the available ListSubGroups() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-subgroups
type ListSubGroupsOptions ListGroupsOptions

// ListSubGroups gets a list of subgroups for a given group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-subgroups
func (s *GroupsService) ListSubGroups(gid interface{}, opt *ListSubGroupsOptions, options ...RequestOptionFunc) ([]*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/subgroups", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*Group
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// ListDescendantGroupsOptions represents the available ListDescendantGroups()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-descendant-groups
type ListDescendantGroupsOptions ListGroupsOptions

// ListDescendantGroups gets a list of subgroups for a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-descendant-groups
func (s *GroupsService) ListDescendantGroups(gid interface{}, opt *ListDescendantGroupsOptions, options ...RequestOptionFunc) ([]*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/descendant_groups", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*Group
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// ListGroupProjectsOptions represents the available ListGroup() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects
type ListGroupProjectsOptions struct {
	ListOptions
	Archived                 *bool             `url:"archived,omitempty" json:"archived,omitempty"`
	IncludeSubGroups         *bool             `url:"include_subgroups,omitempty" json:"include_subgroups,omitempty"`
	MinAccessLevel           *AccessLevelValue `url:"min_access_level,omitempty" json:"min_access_level,omitempty"`
	OrderBy                  *string           `url:"order_by,omitempty" json:"order_by,omitempty"`
	Owned                    *bool             `url:"owned,omitempty" json:"owned,omitempty"`
	Search                   *string           `url:"search,omitempty" json:"search,omitempty"`
	Simple                   *bool             `url:"simple,omitempty" json:"simple,omitempty"`
	Sort                     *string           `url:"sort,omitempty" json:"sort,omitempty"`
	Starred                  *bool             `url:"starred,omitempty" json:"starred,omitempty"`
	Topic                    *string           `url:"topic,omitempty" json:"topic,omitempty"`
	Visibility               *VisibilityValue  `url:"visibility,omitempty" json:"visibility,omitempty"`
	WithCustomAttributes     *bool             `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
	WithIssuesEnabled        *bool             `url:"with_issues_enabled,omitempty" json:"with_issues_enabled,omitempty"`
	WithMergeRequestsEnabled *bool             `url:"with_merge_requests_enabled,omitempty" json:"with_merge_requests_enabled,omitempty"`
	WithSecurityReports      *bool             `url:"with_security_reports,omitempty" json:"with_security_reports,omitempty"`
	WithShared               *bool             `url:"with_shared,omitempty" json:"with_shared,omitempty"`
}

// ListGroupProjects get a list of group projects
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-a-groups-projects
func (s *GroupsService) ListGroupProjects(gid interface{}, opt *ListGroupProjectsOptions, options ...RequestOptionFunc) ([]*Project, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/projects", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*Project
	resp, err := s.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// GetGroupOptions represents the available GetGroup() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#details-of-a-group
type GetGroupOptions struct {
	ListOptions
	WithCustomAttributes *bool `url:"with_custom_attributes,omitempty" json:"with_custom_attributes,omitempty"`
	WithProjects         *bool `url:"with_projects,omitempty" json:"with_projects,omitempty"`
}

// GetGroup gets all details of a group.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#details-of-a-group
func (s *GroupsService) GetGroup(gid interface{}, opt *GetGroupOptions, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// DownloadAvatar downloads a group avatar.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#download-a-group-avatar
func (s *GroupsService) DownloadAvatar(gid interface{}, options ...RequestOptionFunc) (*bytes.Reader, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/avatar", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	avatar := new(bytes.Buffer)
	resp, err := s.client.Do(req, avatar)
	if err != nil {
		return nil, resp, err
	}

	return bytes.NewReader(avatar.Bytes()), resp, err
}

// CreateGroupOptions represents the available CreateGroup() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#new-group
type CreateGroupOptions struct {
	Name                           *string                     `url:"name,omitempty" json:"name,omitempty"`
	Path                           *string                     `url:"path,omitempty" json:"path,omitempty"`
	Avatar                         *GroupAvatar                `url:"-" json:"-"`
	Description                    *string                     `url:"description,omitempty" json:"description,omitempty"`
	MembershipLock                 *bool                       `url:"membership_lock,omitempty" json:"membership_lock,omitempty"`
	Visibility                     *VisibilityValue            `url:"visibility,omitempty" json:"visibility,omitempty"`
	ShareWithGroupLock             *bool                       `url:"share_with_group_lock,omitempty" json:"share_with_group_lock,omitempty"`
	RequireTwoFactorAuth           *bool                       `url:"require_two_factor_authentication,omitempty" json:"require_two_factor_authentication,omitempty"`
	TwoFactorGracePeriod           *int                        `url:"two_factor_grace_period,omitempty" json:"two_factor_grace_period,omitempty"`
	ProjectCreationLevel           *ProjectCreationLevelValue  `url:"project_creation_level,omitempty" json:"project_creation_level,omitempty"`
	AutoDevopsEnabled              *bool                       `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	SubGroupCreationLevel          *SubGroupCreationLevelValue `url:"subgroup_creation_level,omitempty" json:"subgroup_creation_level,omitempty"`
	EmailsDisabled                 *bool                       `url:"emails_disabled,omitempty" json:"emails_disabled,omitempty"`
	MentionsDisabled               *bool                       `url:"mentions_disabled,omitempty" json:"mentions_disabled,omitempty"`
	LFSEnabled                     *bool                       `url:"lfs_enabled,omitempty" json:"lfs_enabled,omitempty"`
	DefaultBranchProtection        *int                        `url:"default_branch_protection,omitempty" json:"default_branch_protection"`
	RequestAccessEnabled           *bool                       `url:"request_access_enabled,omitempty" json:"request_access_enabled,omitempty"`
	ParentID                       *int                        `url:"parent_id,omitempty" json:"parent_id,omitempty"`
	SharedRunnersMinutesLimit      *int                        `url:"shared_runners_minutes_limit,omitempty" json:"shared_runners_minutes_limit,omitempty"`
	ExtraSharedRunnersMinutesLimit *int                        `url:"extra_shared_runners_minutes_limit,omitempty" json:"extra_shared_runners_minutes_limit,omitempty"`
	IPRestrictionRanges            *string                     `url:"ip_restriction_ranges,omitempty" json:"ip_restriction_ranges,omitempty"`
}

// CreateGroup creates a new project group. Available only for users who can
// create groups.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#new-group
func (s *GroupsService) CreateGroup(opt *CreateGroupOptions, options ...RequestOptionFunc) (*Group, *Response, error) {
	var err error
	var req *retryablehttp.Request

	if opt.Avatar == nil {
		req, err = s.client.NewRequest(http.MethodPost, "groups", opt, options)
	} else {
		req, err = s.client.UploadRequest(
			http.MethodPost,
			"groups",
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

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// TransferGroup transfers a project to the Group namespace. Available only
// for admin.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#transfer-project-to-group
func (s *GroupsService) TransferGroup(gid interface{}, pid interface{}, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/projects/%s", PathEscape(group), PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// TransferSubGroupOptions represents the available TransferSubGroup() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#transfer-a-group-to-a-new-parent-group--turn-a-subgroup-to-a-top-level-group
type TransferSubGroupOptions struct {
	GroupID *int `url:"group_id,omitempty" json:"group_id,omitempty"`
}

// TransferSubGroup transfers a group to a new parent group or turn a subgroup
// to a top-level group. Available to administrators and users.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#transfer-a-group-to-a-new-parent-group--turn-a-subgroup-to-a-top-level-group
func (s *GroupsService) TransferSubGroup(gid interface{}, opt *TransferSubGroupOptions, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/transfer", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// UpdateGroupOptions represents the available UpdateGroup() options.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#update-group
type UpdateGroupOptions struct {
	Name                                 *string                     `url:"name,omitempty" json:"name,omitempty"`
	Path                                 *string                     `url:"path,omitempty" json:"path,omitempty"`
	Avatar                               *GroupAvatar                `url:"-" json:"avatar,omitempty"`
	Description                          *string                     `url:"description,omitempty" json:"description,omitempty"`
	MembershipLock                       *bool                       `url:"membership_lock,omitempty" json:"membership_lock,omitempty"`
	Visibility                           *VisibilityValue            `url:"visibility,omitempty" json:"visibility,omitempty"`
	ShareWithGroupLock                   *bool                       `url:"share_with_group_lock,omitempty" json:"share_with_group_lock,omitempty"`
	RequireTwoFactorAuth                 *bool                       `url:"require_two_factor_authentication,omitempty" json:"require_two_factor_authentication,omitempty"`
	TwoFactorGracePeriod                 *int                        `url:"two_factor_grace_period,omitempty" json:"two_factor_grace_period,omitempty"`
	ProjectCreationLevel                 *ProjectCreationLevelValue  `url:"project_creation_level,omitempty" json:"project_creation_level,omitempty"`
	AutoDevopsEnabled                    *bool                       `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	SubGroupCreationLevel                *SubGroupCreationLevelValue `url:"subgroup_creation_level,omitempty" json:"subgroup_creation_level,omitempty"`
	EmailsDisabled                       *bool                       `url:"emails_disabled,omitempty" json:"emails_disabled,omitempty"`
	MentionsDisabled                     *bool                       `url:"mentions_disabled,omitempty" json:"mentions_disabled,omitempty"`
	LFSEnabled                           *bool                       `url:"lfs_enabled,omitempty" json:"lfs_enabled,omitempty"`
	RequestAccessEnabled                 *bool                       `url:"request_access_enabled,omitempty" json:"request_access_enabled,omitempty"`
	DefaultBranchProtection              *int                        `url:"default_branch_protection,omitempty" json:"default_branch_protection,omitempty"`
	FileTemplateProjectID                *int                        `url:"file_template_project_id,omitempty" json:"file_template_project_id,omitempty"`
	SharedRunnersMinutesLimit            *int                        `url:"shared_runners_minutes_limit,omitempty" json:"shared_runners_minutes_limit,omitempty"`
	ExtraSharedRunnersMinutesLimit       *int                        `url:"extra_shared_runners_minutes_limit,omitempty" json:"extra_shared_runners_minutes_limit,omitempty"`
	PreventForkingOutsideGroup           *bool                       `url:"prevent_forking_outside_group,omitempty" json:"prevent_forking_outside_group,omitempty"`
	SharedRunnersSetting                 *SharedRunnersSettingValue  `url:"shared_runners_setting,omitempty" json:"shared_runners_setting,omitempty"`
	PreventSharingGroupsOutsideHierarchy *bool                       `url:"prevent_sharing_groups_outside_hierarchy,omitempty" json:"prevent_sharing_groups_outside_hierarchy,omitempty"`
	IPRestrictionRanges                  *string                     `url:"ip_restriction_ranges,omitempty" json:"ip_restriction_ranges,omitempty"`
}

// UpdateGroup updates an existing group; only available to group owners and
// administrators.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#update-group
func (s *GroupsService) UpdateGroup(gid interface{}, opt *UpdateGroupOptions, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s", PathEscape(group))

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

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// UploadAvatar uploads a group avatar.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#upload-a-group-avatar
func (s *GroupsService) UploadAvatar(gid interface{}, avatar io.Reader, filename string, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s", PathEscape(group))

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

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// DeleteGroup removes group with all projects inside.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#remove-group
func (s *GroupsService) DeleteGroup(gid interface{}, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// RestoreGroup restores a previously deleted group
//
// GitLap API docs:
// https://docs.gitlab.com/ee/api/groups.html#restore-group-marked-for-deletion
func (s *GroupsService) RestoreGroup(gid interface{}, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/restore", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// SearchGroup get all groups that match your string in their name or path.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/groups.html#search-for-group
func (s *GroupsService) SearchGroup(query string, options ...RequestOptionFunc) ([]*Group, *Response, error) {
	var q struct {
		Search string `url:"search,omitempty" json:"search,omitempty"`
	}
	q.Search = query

	req, err := s.client.NewRequest(http.MethodGet, "groups", &q, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*Group
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// ListProvisionedUsersOptions represents the available ListProvisionedUsers()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-provisioned-users
type ListProvisionedUsersOptions struct {
	ListOptions
	Username      *string    `url:"username,omitempty" json:"username,omitempty"`
	Search        *string    `url:"search,omitempty" json:"search,omitempty"`
	Active        *bool      `url:"active,omitempty" json:"active,omitempty"`
	Blocked       *bool      `url:"blocked,omitempty" json:"blocked,omitempty"`
	CreatedAfter  *time.Time `url:"created_after,omitempty" json:"created_after,omitempty"`
	CreatedBefore *time.Time `url:"created_before,omitempty" json:"created_before,omitempty"`
}

// ListProvisionedUsers gets a list of users provisioned by the given group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-provisioned-users
func (s *GroupsService) ListProvisionedUsers(gid interface{}, opt *ListProvisionedUsersOptions, options ...RequestOptionFunc) ([]*User, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/provisioned_users", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var us []*User
	resp, err := s.client.Do(req, &us)
	if err != nil {
		return nil, resp, err
	}

	return us, resp, nil
}

// ListGroupLDAPLinks lists the group's LDAP links. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-ldap-group-links
func (s *GroupsService) ListGroupLDAPLinks(gid interface{}, options ...RequestOptionFunc) ([]*LDAPGroupLink, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/ldap_group_links", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var gls []*LDAPGroupLink
	resp, err := s.client.Do(req, &gls)
	if err != nil {
		return nil, resp, err
	}

	return gls, resp, nil
}

// AddGroupLDAPLinkOptions represents the available AddGroupLDAPLink() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-ldap-group-link-with-cn-or-filter
type AddGroupLDAPLinkOptions struct {
	CN          *string           `url:"cn,omitempty" json:"cn,omitempty"`
	Filter      *string           `url:"filter,omitempty" json:"filter,omitempty"`
	GroupAccess *AccessLevelValue `url:"group_access,omitempty" json:"group_access,omitempty"`
	Provider    *string           `url:"provider,omitempty" json:"provider,omitempty"`
}

// DeleteGroupLDAPLinkWithCNOrFilterOptions represents the available DeleteGroupLDAPLinkWithCNOrFilter() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-ldap-group-link-with-cn-or-filter
type DeleteGroupLDAPLinkWithCNOrFilterOptions struct {
	CN       *string `url:"cn,omitempty" json:"cn,omitempty"`
	Filter   *string `url:"filter,omitempty" json:"filter,omitempty"`
	Provider *string `url:"provider,omitempty" json:"provider,omitempty"`
}

// AddGroupLDAPLink creates a new group LDAP link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-ldap-group-link-with-cn-or-filter
func (s *GroupsService) AddGroupLDAPLink(gid interface{}, opt *AddGroupLDAPLinkOptions, options ...RequestOptionFunc) (*LDAPGroupLink, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/ldap_group_links", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	gl := new(LDAPGroupLink)
	resp, err := s.client.Do(req, gl)
	if err != nil {
		return nil, resp, err
	}

	return gl, resp, nil
}

// DeleteGroupLDAPLink deletes a group LDAP link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-ldap-group-link
func (s *GroupsService) DeleteGroupLDAPLink(gid interface{}, cn string, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/ldap_group_links/%s", PathEscape(group), PathEscape(cn))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteGroupLDAPLinkWithCNOrFilter deletes a group LDAP link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-ldap-group-link-with-cn-or-filter
func (s *GroupsService) DeleteGroupLDAPLinkWithCNOrFilter(gid interface{}, opts *DeleteGroupLDAPLinkWithCNOrFilterOptions, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/ldap_group_links", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodDelete, u, opts, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteGroupLDAPLinkForProvider deletes a group LDAP link from a specific
// provider. Available only for users who can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-ldap-group-link
func (s *GroupsService) DeleteGroupLDAPLinkForProvider(gid interface{}, provider, cn string, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf(
		"groups/%s/ldap_group_links/%s/%s",
		PathEscape(group),
		PathEscape(provider),
		PathEscape(cn),
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListGroupSAMLLinks lists the group's SAML links. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#list-saml-group-links
func (s *GroupsService) ListGroupSAMLLinks(gid interface{}, options ...RequestOptionFunc) ([]*SAMLGroupLink, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/saml_group_links", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var gl []*SAMLGroupLink
	resp, err := s.client.Do(req, &gl)
	if err != nil {
		return nil, resp, err
	}

	return gl, resp, nil
}

// GetGroupSAMLLink get a specific group SAML link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#get-saml-group-link
func (s *GroupsService) GetGroupSAMLLink(gid interface{}, samlGroupName string, options ...RequestOptionFunc) (*SAMLGroupLink, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/saml_group_links/%s", PathEscape(group), PathEscape(samlGroupName))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gl := new(SAMLGroupLink)
	resp, err := s.client.Do(req, &gl)
	if err != nil {
		return nil, resp, err
	}

	return gl, resp, nil
}

// AddGroupSAMLLinkOptions represents the available AddGroupSAMLLink() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-saml-group-link
type AddGroupSAMLLinkOptions struct {
	SAMLGroupName *string           `url:"saml_group_name,omitempty" json:"saml_group_name,omitempty"`
	AccessLevel   *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
}

// AddGroupSAMLLink creates a new group SAML link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-saml-group-link
func (s *GroupsService) AddGroupSAMLLink(gid interface{}, opt *AddGroupSAMLLinkOptions, options ...RequestOptionFunc) (*SAMLGroupLink, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/saml_group_links", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	gl := new(SAMLGroupLink)
	resp, err := s.client.Do(req, &gl)
	if err != nil {
		return nil, resp, err
	}

	return gl, resp, nil
}

// DeleteGroupSAMLLink deletes a group SAML link. Available only for users who
// can edit groups.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-saml-group-link
func (s *GroupsService) DeleteGroupSAMLLink(gid interface{}, samlGroupName string, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/saml_group_links/%s", PathEscape(group), PathEscape(samlGroupName))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ShareGroupWithGroupOptions represents the available ShareGroupWithGroup() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#share-groups-with-groups
type ShareGroupWithGroupOptions struct {
	GroupID     *int              `url:"group_id,omitempty" json:"group_id,omitempty"`
	GroupAccess *AccessLevelValue `url:"group_access,omitempty" json:"group_access,omitempty"`
	ExpiresAt   *ISOTime          `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// ShareGroupWithGroup shares a group with another group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#create-a-link-to-share-a-group-with-another-group
func (s *GroupsService) ShareGroupWithGroup(gid interface{}, opt *ShareGroupWithGroupOptions, options ...RequestOptionFunc) (*Group, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/share", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(Group)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}

// UnshareGroupFromGroup unshares a group from another group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-link-sharing-group-with-another-group
func (s *GroupsService) UnshareGroupFromGroup(gid interface{}, groupID int, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/share/%d", PathEscape(group), groupID)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// GroupPushRules represents a group push rule.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#get-group-push-rules
type GroupPushRules struct {
	ID                         int        `json:"id"`
	CreatedAt                  *time.Time `json:"created_at"`
	CommitMessageRegex         string     `json:"commit_message_regex"`
	CommitMessageNegativeRegex string     `json:"commit_message_negative_regex"`
	BranchNameRegex            string     `json:"branch_name_regex"`
	DenyDeleteTag              bool       `json:"deny_delete_tag"`
	MemberCheck                bool       `json:"member_check"`
	PreventSecrets             bool       `json:"prevent_secrets"`
	AuthorEmailRegex           string     `json:"author_email_regex"`
	FileNameRegex              string     `json:"file_name_regex"`
	MaxFileSize                int        `json:"max_file_size"`
	CommitCommitterCheck       bool       `json:"commit_committer_check"`
	RejectUnsignedCommits      bool       `json:"reject_unsigned_commits"`
}

// GetGroupPushRules gets the push rules of a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#get-group-push-rules
func (s *GroupsService) GetGroupPushRules(gid interface{}, options ...RequestOptionFunc) (*GroupPushRules, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/push_rule", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gpr := new(GroupPushRules)
	resp, err := s.client.Do(req, gpr)
	if err != nil {
		return nil, resp, err
	}

	return gpr, resp, nil
}

// AddGroupPushRuleOptions represents the available AddGroupPushRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-group-push-rule
type AddGroupPushRuleOptions struct {
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

// AddGroupPushRule adds push rules to the specified group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#add-group-push-rule
func (s *GroupsService) AddGroupPushRule(gid interface{}, opt *AddGroupPushRuleOptions, options ...RequestOptionFunc) (*GroupPushRules, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/push_rule", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	gpr := new(GroupPushRules)
	resp, err := s.client.Do(req, gpr)
	if err != nil {
		return nil, resp, err
	}

	return gpr, resp, nil
}

// EditGroupPushRuleOptions represents the available EditGroupPushRule()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#edit-group-push-rule
type EditGroupPushRuleOptions struct {
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

// EditGroupPushRule edits a push rule for a specified group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#edit-group-push-rule
func (s *GroupsService) EditGroupPushRule(gid interface{}, opt *EditGroupPushRuleOptions, options ...RequestOptionFunc) (*GroupPushRules, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/push_rule", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	gpr := new(GroupPushRules)
	resp, err := s.client.Do(req, gpr)
	if err != nil {
		return nil, resp, err
	}

	return gpr, resp, nil
}

// DeleteGroupPushRule deletes the push rules of a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/groups.html#delete-group-push-rule
func (s *GroupsService) DeleteGroupPushRule(gid interface{}, options ...RequestOptionFunc) (*Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/push_rule", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
