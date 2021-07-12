package gqltestutil

import (
	"github.com/cockroachdb/errors"
)

// Organization contains basic information of an organization.
type Organization struct {
	ID             string `json:"id"`
	ViewerIsMember bool   `json:"viewerIsMember"`
}

// Organization returns basic information of the given organization.
func (c *Client) Organization(name string) (*Organization, error) {
	const query = `
query Organization($name: String!) {
	organization(name: $name) {
		id
		viewerIsMember
	}
}
`
	variables := map[string]interface{}{
		"name": name,
	}
	var resp struct {
		Data struct {
			*Organization `json:"organization"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Organization, nil
}

// OrganizationMember contains basic information of an organization member.
type OrganizationMember struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// OrganizationMembers returns a list of members of the given organization.
func (c *Client) OrganizationMembers(id string) ([]*OrganizationMember, error) {
	const query = `
query OrganizationMembers($id: ID!) {
	node(id: $id) {
		... on Org {
			members {
				nodes {
					id
					username
				}
			}
		}
	}
}
`
	variables := map[string]interface{}{
		"id": id,
	}
	var resp struct {
		Data struct {
			Node struct {
				Members struct {
					Nodes []*OrganizationMember `json:"nodes"`
				} `json:"members"`
			} `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}
	return resp.Data.Node.Members.Nodes, nil
}

// InviteUserToOrganizationResult contains information of a user invitation to
// an organization.
type InviteUserToOrganizationResult struct {
	SentInvitationEmail bool   `json:"sentInvitationEmail"`
	InvitationURL       string `json:"invitationURL"`
}

// InviteUserToOrganization invites a user to the given organization.
func (c *Client) InviteUserToOrganization(orgID, username string) (*InviteUserToOrganizationResult, error) {
	const query = `
mutation InviteUserToOrganization($organization: ID!, $username: String!) {
	inviteUserToOrganization(organization: $organization, username: $username) {
		... on InviteUserToOrganizationResult {
			sentInvitationEmail
			invitationURL
		}
	}
}
`
	variables := map[string]interface{}{
		"organization": orgID,
		"username":     username,
	}
	var resp struct {
		Data struct {
			*InviteUserToOrganizationResult `json:"inviteUserToOrganization"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}
	return resp.Data.InviteUserToOrganizationResult, nil
}

// CreateOrganization creates a new organization with given name and display name.
// It returns GraphQL node ID of newly created organization.
func (c *Client) CreateOrganization(name, displayName string) (string, error) {
	const query = `
mutation CreateOrganization($name: String!, $displayName: String) {
	createOrganization(name: $name, displayName: $displayName) {
		id
	}
}
`
	variables := map[string]interface{}{
		"name":        name,
		"displayName": displayName,
	}
	var resp struct {
		Data struct {
			CreateOrganization struct {
				ID string `json:"id"`
			} `json:"createOrganization"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}
	return resp.Data.CreateOrganization.ID, nil
}

// UpdateOrganization updates display name of the given organization.
func (c *Client) UpdateOrganization(id, displayName string) error {
	const query = `
mutation UpdateOrganization($id: ID!, $displayName: String) {
	updateOrganization(id: $id, displayName: $displayName) {
		id
	}
}
`
	variables := map[string]interface{}{
		"id":          id,
		"displayName": displayName,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	return nil
}

// DeleteOrganization deletes the organization by given GraphQL node ID.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) DeleteOrganization(id string) error {
	const query = `
mutation DeleteOrganization($organization: ID!) {
	deleteOrganization(organization: $organization) {
		alwaysNil
	}
}
`
	variables := map[string]interface{}{
		"organization": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// RemoveUserFromOrganization removes user from given organization.
func (c *Client) RemoveUserFromOrganization(userID, orgID string) error {
	const query = `
mutation RemoveUserFromOrganization($user: ID!, $organization: ID!) {
	removeUserFromOrganization(user: $user, organization: $organization) {
		alwaysNil
	}
}
`
	variables := map[string]interface{}{
		"user":         userID,
		"organization": orgID,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
