package gqltestutil

import "github.com/sourcegraph/sourcegraph/lib/errors"

// CreateUser creates a new user with the given username and email.
// It returns the GraphQL node ID of newly created user.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) CreateUser(username, email string) (string, error) {
	const query = `
mutation CreateUser($username: String!, $email: String) {
	createUser(username: $username, email: $email) {
		user {
			id
		}
	}
}
`
	variables := map[string]any{
		"username": username,
		"email":    email,
	}
	var resp struct {
		Data struct {
			CreateUser struct {
				User struct {
					ID string `json:"id"`
				} `json:"user"`
			} `json:"createUser"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateUser.User.ID, nil
}

// DeleteUser deletes a user by given GraphQL node ID.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) DeleteUser(id string, hard bool) error {
	const query = `
mutation DeleteUser($user: ID!, $hard: Boolean) {
	deleteUser(user: $user, hard: $hard) {
		alwaysNil
	}
}
`
	variables := map[string]any{
		"user": id,
		"hard": hard,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// SetUserEmailVerified sets the given user's email address verification status.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) SetUserEmailVerified(user, email string, verified bool) error {
	const query = `
mutation setUserEmailVerified($user: ID!, $email: String!, $verified: Boolean!) {
	setUserEmailVerified(user: $user, email: $email, verified: $verified) {
      alwaysNil
	}
}
`
	variables := map[string]any{
		"user":     user,
		"email":    email,
		"verified": verified,
	}
	var resp struct {
		Data struct {
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	return nil
}

// UserOrganizations returns organizations name the given user belongs to.
func (c *Client) UserOrganizations(username string) ([]string, error) {
	const query = `
query User($username: String) {
	user(username: $username) {
		organizations {
			nodes {
				name
			}
		}
	}
}
`
	variables := map[string]any{
		"username": username,
	}
	var resp struct {
		Data struct {
			User struct {
				Organizations struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"organizations"`
			} `json:"user"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	names := make([]string, 0, len(resp.Data.User.Organizations.Nodes))
	for _, node := range resp.Data.User.Organizations.Nodes {
		names = append(names, node.Name)
	}
	return names, nil
}

// UserPermissions returns repo permissions that given user has.
func (c *Client) UserPermissions(username string) ([]string, error) {
	const query = `
query User($username: String) {
	user(username: $username) {
		permissionsInfo {
			permissions
		}
	}
}
`
	variables := map[string]any{
		"username": username,
	}
	var resp struct {
		Data struct {
			User struct {
				PermissionsInfo struct {
					Permissions []string `json:"permissions"`
				} `json:"permissionsInfo"`
			} `json:"user"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.User.PermissionsInfo.Permissions, nil
}
