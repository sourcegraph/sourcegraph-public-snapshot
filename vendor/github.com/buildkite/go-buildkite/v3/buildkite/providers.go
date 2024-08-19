package buildkite

import "encoding/json"

// Provider represents a source code provider. It is read-only, but settings may be written using Pipeline.ProviderSettings.
type Provider struct {
	ID         string           `json:"id" yaml:"id"`
	WebhookURL *string          `json:"webhook_url" yaml:"webhook_url"`
	Settings   ProviderSettings `json:"settings" yaml:"settings"`
}

// UnmarshalJSON decodes the Provider, choosing the type of the Settings from the ID.
func (p *Provider) UnmarshalJSON(data []byte) error {
	type provider Provider
	var v struct {
		provider
		Settings json.RawMessage `json:"settings" yaml:"settings"`
	}

	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	*p = Provider(v.provider)

	var settings ProviderSettings
	switch v.ID {
	case "bitbucket":
		settings = &BitbucketSettings{}
	case "github":
		settings = &GitHubSettings{}
	case "github_enterprise":
		settings = &GitHubEnterpriseSettings{}
	case "gitlab":
		settings = &GitLabSettings{}
	default:
		return nil
	}

	err = json.Unmarshal(v.Settings, settings)
	if err != nil {
		return err
	}
	p.Settings = settings

	return nil
}

// ProviderSettings represents the sum type of the settings for different source code providers.
type ProviderSettings interface {
	isProviderSettings()
}

// BitbucketSettings are settings for pipelines building from Bitbucket repositories.
type BitbucketSettings struct {
	BuildPullRequests                       *bool   `json:"build_pull_requests,omitempty" yaml:"build_pull_requests,omitempty"`
	BuildBranches                           *bool   `json:"build_branches,omitempty" yaml:"build_branches,omitempty"`
	PullRequestBranchFilterEnabled          *bool   `json:"pull_request_branch_filter_enabled,omitempty" yaml:"pull_request_branch_filter_enabled,omitempty"`
	PullRequestBranchFilterConfiguration    *string `json:"pull_request_branch_filter_configuration,omitempty" yaml:"pull_request_branch_filter_configuration,omitempty"`
	SkipPullRequestBuildsForExistingCommits *bool   `json:"skip_pull_request_builds_for_existing_commits,omitempty" yaml:"skip_pull_request_builds_for_existing_commits,omitempty"`
	BuildTags                               *bool   `json:"build_tags,omitempty" yaml:"build_tags,omitempty"`
	PublishCommitStatus                     *bool   `json:"publish_commit_status,omitempty" yaml:"publish_commit_status,omitempty"`
	PublishCommitStatusPerStep              *bool   `json:"publish_commit_status_per_step,omitempty" yaml:"publish_commit_status_per_step,omitempty"`

	// Read-only
	Repository *string `json:"repository,omitempty" yaml:"repository,omitempty"`
}

func (s *BitbucketSettings) isProviderSettings() {}

// GitHubSettings are settings for pipelines building from GitHub repositories.
type GitHubSettings struct {
	TriggerMode                             *string `json:"trigger_mode,omitempty" yaml:"trigger_mode,omitempty"`
	BuildPullRequests                       *bool   `json:"build_pull_requests,omitempty" yaml:"build_pull_requests,omitempty"`
	BuildBranches                           *bool   `json:"build_branches,omitempty" yaml:"build_branches,omitempty"`
	PullRequestBranchFilterEnabled          *bool   `json:"pull_request_branch_filter_enabled,omitempty" yaml:"pull_request_branch_filter_enabled,omitempty"`
	PullRequestBranchFilterConfiguration    *string `json:"pull_request_branch_filter_configuration,omitempty" yaml:"pull_request_branch_filter_configuration,omitempty"`
	SkipPullRequestBuildsForExistingCommits *bool   `json:"skip_pull_request_builds_for_existing_commits,omitempty" yaml:"skip_pull_request_builds_for_existing_commits,omitempty"`
	BuildPullRequestForks                   *bool   `json:"build_pull_request_forks,omitempty" yaml:"build_pull_request_forks,omitempty"`
	PrefixPullRequestForkBranchNames        *bool   `json:"prefix_pull_request_fork_branch_names,omitempty" yaml:"prefix_pull_request_fork_branch_names,omitempty"`
	BuildTags                               *bool   `json:"build_tags,omitempty" yaml:"build_tags,omitempty"`
	PublishCommitStatus                     *bool   `json:"publish_commit_status,omitempty" yaml:"publish_commit_status,omitempty"`
	PublishCommitStatusPerStep              *bool   `json:"publish_commit_status_per_step,omitempty" yaml:"publish_commit_status_per_step,omitempty"`
	FilterEnabled                           *bool   `json:"filter_enabled,omitempty" yaml:"filter_enabled,omitempty"`
	FilterCondition                         *string `json:"filter_condition,omitempty" yaml:"filter_condition,omitempty"`
	SeparatePullRequestStatuses             *bool   `json:"separate_pull_request_statuses,omitempty" yaml:"separate_pull_request_statuses,omitempty"`
	PublishBlockedAsPending                 *bool   `json:"publish_blocked_as_pending,omitempty" yaml:"publish_blocked_as_pending,omitempty"`

	// Read-only
	Repository *string `json:"repository,omitempty" yaml:"repository,omitempty"`
}

func (s *GitHubSettings) isProviderSettings() {}

// GitHubEnterpriseSettings are settings for pipelines building from GitHub Enterprise repositories.
type GitHubEnterpriseSettings struct {
	TriggerMode                             *string `json:"trigger_mode,omitempty" yaml:"trigger_mode,omitempty"`
	BuildPullRequests                       *bool   `json:"build_pull_requests,omitempty" yaml:"build_pull_requests,omitempty"`
	BuildBranches                           *bool   `json:"build_branches,omitempty" yaml:"build_branches,omitempty"`
	PullRequestBranchFilterEnabled          *bool   `json:"pull_request_branch_filter_enabled,omitempty" yaml:"pull_request_branch_filter_enabled,omitempty"`
	PullRequestBranchFilterConfiguration    *string `json:"pull_request_branch_filter_configuration,omitempty" yaml:"pull_request_branch_filter_configuration,omitempty"`
	SkipPullRequestBuildsForExistingCommits *bool   `json:"skip_pull_request_builds_for_existing_commits,omitempty" yaml:"skip_pull_request_builds_for_existing_commits,omitempty"`
	BuildTags                               *bool   `json:"build_tags,omitempty" yaml:"build_tags,omitempty"`
	PublishCommitStatus                     *bool   `json:"publish_commit_status,omitempty" yaml:"publish_commit_status,omitempty"`
	PublishCommitStatusPerStep              *bool   `json:"publish_commit_status_per_step,omitempty" yaml:"publish_commit_status_per_step,omitempty"`

	// Read-only
	Repository *string `json:"repository,omitempty" yaml:"repository,omitempty"`
}

func (s *GitHubEnterpriseSettings) isProviderSettings() {}

// GitLabSettings are settings for pipelines building from GitLab repositories.
type GitLabSettings struct {
	// Read-only
	Repository *string `json:"repository,omitempty" yaml:"repository,omitempty"`
}

func (s *GitLabSettings) isProviderSettings() {}
