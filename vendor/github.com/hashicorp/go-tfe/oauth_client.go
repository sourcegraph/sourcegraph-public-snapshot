// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfe

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Compile-time proof of interface implementation.
var _ OAuthClients = (*oAuthClients)(nil)

// OAuthClients describes all the OAuth client related methods that the
// Terraform Enterprise API supports.
//
// TFE API docs:
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/oauth-clients
type OAuthClients interface {
	// List all the OAuth clients for a given organization.
	List(ctx context.Context, organization string, options *OAuthClientListOptions) (*OAuthClientList, error)

	// Create an OAuth client to connect an organization and a VCS provider.
	Create(ctx context.Context, organization string, options OAuthClientCreateOptions) (*OAuthClient, error)

	// Read an OAuth client by its ID.
	Read(ctx context.Context, oAuthClientID string) (*OAuthClient, error)

	// Update an existing OAuth client by its ID.
	Update(ctx context.Context, oAuthClientID string, options OAuthClientUpdateOptions) (*OAuthClient, error)

	// Delete an OAuth client by its ID.
	Delete(ctx context.Context, oAuthClientID string) error
}

// oAuthClients implements OAuthClients.
type oAuthClients struct {
	client *Client
}

// ServiceProviderType represents a VCS type.
type ServiceProviderType string

// List of available VCS types.
const (
	ServiceProviderAzureDevOpsServer   ServiceProviderType = "ado_server"
	ServiceProviderAzureDevOpsServices ServiceProviderType = "ado_services"
	ServiceProviderBitbucket           ServiceProviderType = "bitbucket_hosted"
	// Bitbucket Server v5.4.0 and above
	ServiceProviderBitbucketServer ServiceProviderType = "bitbucket_server"
	// Bitbucket Server v5.3.0 and below
	ServiceProviderBitbucketServerLegacy ServiceProviderType = "bitbucket_server_legacy"
	ServiceProviderGithub                ServiceProviderType = "github"
	ServiceProviderGithubEE              ServiceProviderType = "github_enterprise"
	ServiceProviderGitlab                ServiceProviderType = "gitlab_hosted"
	ServiceProviderGitlabCE              ServiceProviderType = "gitlab_community_edition"
	ServiceProviderGitlabEE              ServiceProviderType = "gitlab_enterprise_edition"
)

// OAuthClientList represents a list of OAuth clients.
type OAuthClientList struct {
	*Pagination
	Items []*OAuthClient
}

// OAuthClient represents a connection between an organization and a VCS
// provider.
type OAuthClient struct {
	ID                  string              `jsonapi:"primary,oauth-clients"`
	APIURL              string              `jsonapi:"attr,api-url"`
	CallbackURL         string              `jsonapi:"attr,callback-url"`
	ConnectPath         string              `jsonapi:"attr,connect-path"`
	CreatedAt           time.Time           `jsonapi:"attr,created-at,iso8601"`
	HTTPURL             string              `jsonapi:"attr,http-url"`
	Key                 string              `jsonapi:"attr,key"`
	RSAPublicKey        string              `jsonapi:"attr,rsa-public-key"`
	Name                *string             `jsonapi:"attr,name"`
	Secret              string              `jsonapi:"attr,secret"`
	ServiceProvider     ServiceProviderType `jsonapi:"attr,service-provider"`
	ServiceProviderName string              `jsonapi:"attr,service-provider-display-name"`

	// Relations
	Organization *Organization `jsonapi:"relation,organization"`
	OAuthTokens  []*OAuthToken `jsonapi:"relation,oauth-tokens"`
}

// A list of relations to include
type OAuthClientIncludeOpt string

const OauthClientOauthTokens OAuthClientIncludeOpt = "oauth_tokens"

// OAuthClientListOptions represents the options for listing
// OAuth clients.
type OAuthClientListOptions struct {
	ListOptions

	Include []OAuthClientIncludeOpt `url:"include,omitempty"`
}

// OAuthClientCreateOptions represents the options for creating an OAuth client.
type OAuthClientCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,oauth-clients"`

	// A display name for the OAuth Client.
	Name *string `jsonapi:"attr,name"`

	// Required: The base URL of your VCS provider's API.
	APIURL *string `jsonapi:"attr,api-url"`

	// Required: The homepage of your VCS provider.
	HTTPURL *string `jsonapi:"attr,http-url"`

	// Optional: The OAuth Client key.
	Key *string `jsonapi:"attr,key,omitempty"`

	// Optional: The token string you were given by your VCS provider.
	OAuthToken *string `jsonapi:"attr,oauth-token-string,omitempty"`

	// Optional: Private key associated with this vcs provider - only available for ado_server
	PrivateKey *string `jsonapi:"attr,private-key,omitempty"`

	// Optional: Secret key associated with this vcs provider - only available for ado_server
	Secret *string `jsonapi:"attr,secret,omitempty"`

	// Optional: RSAPublicKey the text of the SSH public key associated with your BitBucket
	// Server Application Link.
	RSAPublicKey *string `jsonapi:"attr,rsa-public-key,omitempty"`

	// Required: The VCS provider being connected with.
	ServiceProvider *ServiceProviderType `jsonapi:"attr,service-provider"`
}

// OAuthClientUpdateOptions represents the options for updating an OAuth client.
type OAuthClientUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,oauth-clients"`

	// Optional: A display name for the OAuth Client.
	Name *string `jsonapi:"attr,name,omitempty"`

	// Optional: The OAuth Client key.
	Key *string `jsonapi:"attr,key,omitempty"`

	// Optional: Secret key associated with this vcs provider - only available for ado_server
	Secret *string `jsonapi:"attr,secret,omitempty"`

	// Optional: RSAPublicKey the text of the SSH public key associated with your BitBucket
	// Server Application Link.
	RSAPublicKey *string `jsonapi:"attr,rsa-public-key,omitempty"`

	// Optional: The token string you were given by your VCS provider.
	OAuthToken *string `jsonapi:"attr,oauth-token-string,omitempty"`
}

// List all the OAuth clients for a given organization.
func (s *oAuthClients) List(ctx context.Context, organization string, options *OAuthClientListOptions) (*OAuthClientList, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/oauth-clients", url.QueryEscape(organization))
	req, err := s.client.NewRequest("GET", u, options)
	if err != nil {
		return nil, err
	}

	ocl := &OAuthClientList{}
	err = req.Do(ctx, ocl)
	if err != nil {
		return nil, err
	}

	return ocl, nil
}

// Create an OAuth client to connect an organization and a VCS provider.
func (s *oAuthClients) Create(ctx context.Context, organization string, options OAuthClientCreateOptions) (*OAuthClient, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}
	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/oauth-clients", url.QueryEscape(organization))
	req, err := s.client.NewRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	oc := &OAuthClient{}
	err = req.Do(ctx, oc)
	if err != nil {
		return nil, err
	}

	return oc, nil
}

// Read an OAuth client by its ID.
func (s *oAuthClients) Read(ctx context.Context, oAuthClientID string) (*OAuthClient, error) {
	if !validStringID(&oAuthClientID) {
		return nil, ErrInvalidOauthClientID
	}

	u := fmt.Sprintf("oauth-clients/%s", url.QueryEscape(oAuthClientID))
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	oc := &OAuthClient{}
	err = req.Do(ctx, oc)
	if err != nil {
		return nil, err
	}

	return oc, err
}

// Update an OAuth client by its ID.
func (s *oAuthClients) Update(ctx context.Context, oAuthClientID string, options OAuthClientUpdateOptions) (*OAuthClient, error) {
	if !validStringID(&oAuthClientID) {
		return nil, ErrInvalidOauthClientID
	}

	u := fmt.Sprintf("oauth-clients/%s", url.QueryEscape(oAuthClientID))
	req, err := s.client.NewRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	oc := &OAuthClient{}
	err = req.Do(ctx, oc)
	if err != nil {
		return nil, err
	}

	return oc, err
}

// Delete an OAuth client by its ID.
func (s *oAuthClients) Delete(ctx context.Context, oAuthClientID string) error {
	if !validStringID(&oAuthClientID) {
		return ErrInvalidOauthClientID
	}

	u := fmt.Sprintf("oauth-clients/%s", url.QueryEscape(oAuthClientID))
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return req.Do(ctx, nil)
}

func (o OAuthClientCreateOptions) valid() error {
	if !validString(o.APIURL) {
		return ErrRequiredAPIURL
	}
	if !validString(o.HTTPURL) {
		return ErrRequiredHTTPURL
	}
	if o.ServiceProvider == nil {
		return ErrRequiredServiceProvider
	}
	if !validString(o.OAuthToken) && *o.ServiceProvider != *ServiceProvider(ServiceProviderBitbucketServer) {
		return ErrRequiredOauthToken
	}
	if validString(o.PrivateKey) && *o.ServiceProvider != *ServiceProvider(ServiceProviderAzureDevOpsServer) {
		return ErrUnsupportedPrivateKey
	}
	return nil
}

func (o *OAuthClientListOptions) valid() error {
	if o == nil {
		return nil // nothing to validate
	}

	if err := validateOauthClientIncludeParams(o.Include); err != nil {
		return err
	}

	return nil
}

func validateOauthClientIncludeParams(params []OAuthClientIncludeOpt) error {
	for _, p := range params {
		switch p {
		case OauthClientOauthTokens:
			// do nothing
		default:
			return ErrInvalidIncludeValue
		}
	}

	return nil
}
