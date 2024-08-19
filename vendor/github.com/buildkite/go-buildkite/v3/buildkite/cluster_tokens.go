package buildkite

import (
	"errors"
	"fmt"
)

// ClusterTokensService handles communication with cluster token related
// methods of the Buildkite API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/clusters#cluster-tokens
type ClusterTokensService struct {
	client *Client
}

type ClusterToken struct {
	ID                 *string         `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID          *string         `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Description        *string         `json:"description,omitempty" yaml:"description,omitempty"`
	AllowedIPAddresses *string         `json:"allowed_ip_addresses,omitempty" yaml:"allowed_ip_addresses,omitempty"`
	URL                *string         `json:"url,omitempty" yaml:"url,omitempty"`
	ClusterURL         *string         `json:"cluster_url,omitempty" yaml:"cluster_url,omitempty"`
	CreatedAt          *Timestamp      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy          *ClusterCreator `json:"created_by,omitempty" yaml:"created_by,omitempty"`
	Token              *string         `json:"token,omitempty" yaml:"token,omitempty"`
}

type ClusterTokenCreateUpdate struct {
	Description        *string `json:"description,omitempty" yaml:"description,omitempty"`
	AllowedIPAddresses *string `json:"allowed_ip_addresses,omitempty" yaml:"allowed_ip_addresses,omitempty"`
}

type ClusterTokensListOptions struct {
	ListOptions
}

func (cts *ClusterTokensService) List(org, clusterID string, opt *ClusterTokensListOptions) ([]ClusterToken, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/tokens", org, clusterID)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := cts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	tokens := new([]ClusterToken)

	resp, err := cts.client.Do(req, tokens)

	if err != nil {
		return nil, resp, err
	}

	return *tokens, resp, err
}

func (cts *ClusterTokensService) Get(org, clusterID, tokenID string) (*ClusterToken, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/tokens/%s", org, clusterID, tokenID)

	req, err := cts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	token := new(ClusterToken)

	resp, err := cts.client.Do(req, token)

	if err != nil {
		return nil, resp, err
	}

	return token, resp, err
}

func (cts *ClusterTokensService) Create(org, clusterID string, ctc *ClusterTokenCreateUpdate) (*ClusterToken, *Response, error) {

	if ctc == nil {
		return nil, nil, errors.New("ClusterTokenCreateUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/tokens", org, clusterID)

	req, err := cts.client.NewRequest("POST", u, ctc)

	if err != nil {
		return nil, nil, err
	}

	token := new(ClusterToken)

	resp, err := cts.client.Do(req, token)

	if err != nil {
		return nil, resp, err
	}

	return token, resp, err
}

func (cts *ClusterTokensService) Update(org, clusterID, tokenID string, ctc *ClusterTokenCreateUpdate) (*Response, error) {

	if ctc == nil {
		return nil, errors.New("ClusterTokenCreateUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/tokens/%s", org, clusterID, tokenID)

	req, err := cts.client.NewRequest("PATCH", u, ctc)

	if err != nil {
		return nil, err
	}

	resp, err := cts.client.Do(req, ctc)

	if err != nil {
		return resp, err
	}

	return resp, err
}

func (cts *ClusterTokensService) Delete(org, clusterID, tokenID string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/tokens/%s", org, clusterID, tokenID)

	req, err := cts.client.NewRequest("DELETE", u, nil)

	if err != nil {
		return nil, err
	}

	return cts.client.Do(req, nil)
}
