package gqltestutil

import "github.com/sourcegraph/sourcegraph/lib/errors"

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
