package buildkite

import (
	"errors"
	"fmt"
)

// ClustersService handles communication with cluster related
// methods of the Buildkite API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/clusters#clusters
type ClustersService struct {
	client *Client
}

type ClusterCreate struct {
	Name        string  `json:"name,omitempty" yaml:"name,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Emoji       *string `json:"emoji,omitempty" yaml:"emoji,omitempty"`
	Color       *string `json:"color,omitempty" yaml:"color,omitempty"`
}

type ClusterUpdate struct {
	Name           *string `json:"name,omitempty" yaml:"name,omitempty"`
	Description    *string `json:"description,omitempty" yaml:"description,omitempty"`
	Emoji          *string `json:"emoji,omitempty" yaml:"emoji,omitempty"`
	Color          *string `json:"color,omitempty" yaml:"color,omitempty"`
	DefaultQueueID *string `json:"default_queue_id,omitempty" yaml:"default_queue_id,omitempty"`
}

type Cluster struct {
	ID              *string         `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID       *string         `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	DefaultQueueID  *string         `json:"default_queue_id,omitempty" yaml:"default_queue_id,omitempty"`
	Name            *string         `json:"name,omitempty" yaml:"name,omitempty"`
	Description     *string         `json:"description,omitempty" yaml:"description,omitempty"`
	Emoji           *string         `json:"emoji,omitempty" yaml:"emoji,omitempty"`
	Color           *string         `json:"color,omitempty" yaml:"color,omitempty"`
	URL             *string         `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL          *string         `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	QueuesURL       *string         `json:"queues_url,omitempty" yaml:"queues_url,omitempty"`
	DefaultQueueURL *string         `json:"default_queue_url,omitempty" yaml:"default_queue_url,omitempty"`
	CreatedAt       *Timestamp      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy       *ClusterCreator `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

type ClusterCreator struct {
	ID        *string    `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID *string    `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Name      *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Email     *string    `json:"email,omitempty" yaml:"email,omitempty"`
	AvatarURL *string    `json:"avatar_url,omitempty" yaml:"avatar_url,omitempty"`
	CreatedAt *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

type ClustersListOptions struct {
	ListOptions
}

func (cs *ClustersService) List(org string, opt *ClustersListOptions) ([]Cluster, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters", org)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := cs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	clusters := new([]Cluster)

	resp, err := cs.client.Do(req, clusters)

	if err != nil {
		return nil, resp, err
	}

	return *clusters, resp, err
}

func (cs *ClustersService) Get(org, id string) (*Cluster, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s", org, id)

	req, err := cs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	cluster := new(Cluster)

	resp, err := cs.client.Do(req, cluster)

	if err != nil {
		return nil, resp, err
	}

	return cluster, resp, err
}

func (cs *ClustersService) Create(org string, cc *ClusterCreate) (*Cluster, *Response, error) {

	if cc == nil {
		return nil, nil, errors.New("ClusterCreate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters", org)

	req, err := cs.client.NewRequest("POST", u, cc)

	if err != nil {
		return nil, nil, err
	}

	cluster := new(Cluster)

	resp, err := cs.client.Do(req, cluster)

	if err != nil {
		return nil, resp, err
	}

	return cluster, resp, err
}

func (cs *ClustersService) Update(org, id string, cu *ClusterUpdate) (*Response, error) {

	if cu == nil {
		return nil, errors.New("ClusterUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s", org, id)

	req, err := cs.client.NewRequest("PATCH", u, cu)

	if err != nil {
		return nil, nil
	}

	resp, err := cs.client.Do(req, cu)

	if err != nil {
		return resp, err
	}

	return resp, err
}

func (cs *ClustersService) Delete(org, id string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s", org, id)

	req, err := cs.client.NewRequest("DELETE", u, nil)

	if err != nil {
		return nil, err
	}

	return cs.client.Do(req, nil)
}
