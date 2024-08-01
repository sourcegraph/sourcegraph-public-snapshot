package gqltestutil

import (
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ScheduleRepositoryPermissionsSync schedules a permissions syncing request for
// the given repository.
func (c *Client) ScheduleRepositoryPermissionsSync(id string) error {
	const query = `
mutation ScheduleRepositoryPermissionsSync($repository: ID!) {
	scheduleRepositoryPermissionsSync(repository: $repository) {
		alwaysNil
	}
}
`
	variables := map[string]any{
		"repository": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// ScheduleUserPermissionsSync schedules a permissions syncing request for
// the given user.
func (c *Client) ScheduleUserPermissionsSync(id string) error {
	const query = `
mutation ScheduleUserPermissionsSync($user: ID!) {
	scheduleUserPermissionsSync(user: $user) {
		alwaysNil
	}
}
`
	variables := map[string]any{
		"user": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// UserPermissionsInfo returns permissions information of the given
// user.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) UserPermissionsInfo(name string) (*PermissionsInfo, error) {
	const query = `
query UserPermissionsInfo($name: String!) {
	user(username: $name) {
		permissionsInfo {
			syncedAt
			updatedAt
			permissions
			unrestricted
		}
	}
}
`
	variables := map[string]any{
		"name": name,
	}
	var resp struct {
		Data struct {
			User struct {
				*PermissionsInfo `json:"permissionsInfo"`
			} `json:"user"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.User.PermissionsInfo, nil
}

type BitbucketProjectPermsSyncArgs struct {
	ProjectKey      string
	CodeHost        string
	UserPermissions []types.UserPermission
	Unrestricted    *bool
}

type UserPermission struct {
	BindID     string `json:"bindID"`
	Permission string `json:"permission"`
}

// SetRepositoryPermissionsForBitbucketProject requests to set repo permissions for given Bitbucket Project and users.
func (c *Client) SetRepositoryPermissionsForBitbucketProject(args BitbucketProjectPermsSyncArgs) error {
	const query = `
mutation SetRepositoryPermissionsForBitbucketProject($projectKey: String!, $codeHost: ID!, $userPermissions: [UserPermissionInput!]!, $unrestricted: Boolean) {
	setRepositoryPermissionsForBitbucketProject(
		projectKey: $projectKey
		codeHost: $codeHost
		userPermissions: $userPermissions
		unrestricted: $unrestricted
	) {
		alwaysNil
	}
}
`
	variables := map[string]any{
		"projectKey":      args.ProjectKey,
		"codeHost":        graphql.ID(args.CodeHost),
		"userPermissions": args.UserPermissions,
		"unrestricted":    args.Unrestricted,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// GetLastBitbucketProjectPermissionJob returns a status of the most recent
// BitbucketProjectPermissionJob for given projectKey
func (c *Client) GetLastBitbucketProjectPermissionJob(projectKey string) (state string, failureMessage string, err error) {
	const query = `
query BitbucketProjectPermissionJobs($projectKeys: [String!], $status: String, $count: Int) {
	bitbucketProjectPermissionJobs(projectKeys: $projectKeys, status: $status, count: $count) {
		totalCount,
   		nodes {
			State
			FailureMessage
   		}
	}
}
`
	variables := map[string]any{
		"projectKeys": []string{projectKey},
	}
	var resp struct {
		Data struct {
			Jobs struct {
				TotalCount int `json:"totalCount"`
				Nodes      []struct {
					State          string `json:"state"`
					FailureMessage string `json:"failureMessage"`
				} `json:"nodes"`
			} `json:"bitbucketProjectPermissionJobs"`
		} `json:"data"`
	}
	err = c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", "", errors.Wrap(err, "request GraphQL")
	}

	if resp.Data.Jobs.TotalCount < 1 {
		return "", "", nil
	} else {
		job := resp.Data.Jobs.Nodes[0]
		return job.State, job.FailureMessage, nil
	}
}

// UsersWithPendingPermissions returns bind IDs of users with pending permissions
func (c *Client) UsersWithPendingPermissions() ([]string, error) {
	const query = `
query {
	usersWithPendingPermissions
}
`
	var resp struct {
		Data struct {
			UsersWithPendingPermissions []string `json:"usersWithPendingPermissions"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.UsersWithPendingPermissions, nil
}
