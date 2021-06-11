package gqltestutil

import "github.com/cockroachdb/errors"

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
	variables := map[string]interface{}{
		"repository": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
