package buildkite

import (
	"errors"
	"fmt"
)

// ClusterQueuesService handles communication with cluster queue related
// methods of the Buildkite API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/clusters#cluster-queues
type ClusterQueuesService struct {
	client *Client
}

type ClusterQueue struct {
	ID                 *string         `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID          *string         `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Key                *string         `json:"key,omitempty" yaml:"key,omitempty"`
	Description        *string         `json:"description,omitempty" yaml:"description,omitempty"`
	URL                *string         `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL             *string         `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	ClusterURL         *string         `json:"cluster_url,omitempty" yaml:"cluster_url,omitempty"`
	DispatchPaused     *bool           `json:"dispatch_paused,omitempty" yaml:"dispatch_paused,omitempty"`
	DispatchPausedBy   *ClusterCreator `json:"dispatch_paused_by,omitempty" yaml:"dispatch_paused_by,omitempty"`
	DispatchPausedAt   *Timestamp      `json:"dispatch_paused_at,omitempty" yaml:"dispatch_paused_at,omitempty"`
	DispatchPausedNote *string         `json:"dispatch_paused_note,omitempty" yaml:"dispatch_paused_note,omitempty"`
	CreatedAt          *Timestamp      `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy          *ClusterCreator `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

type ClusterQueueCreate struct {
	Key         *string `json:"key,omitempty" yaml:"key,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
}

type ClusterQueueUpdate struct {
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
}

type ClusterQueuePause struct {
	Note *string `json:"dispatch_paused_note,omitempty" yaml:"dispatch_paused_note,omitempty"`
}

type ClusterQueuesListOptions struct {
	ListOptions
}

func (cqs *ClusterQueuesService) List(org, clusterID string, opt *ClusterQueuesListOptions) ([]ClusterQueue, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues", org, clusterID)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := cqs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	queues := new([]ClusterQueue)

	resp, err := cqs.client.Do(req, queues)

	if err != nil {
		return nil, resp, err
	}

	return *queues, resp, err
}

func (cqs *ClusterQueuesService) Get(org, clusterID, queueID string) (*ClusterQueue, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues/%s", org, clusterID, queueID)

	req, err := cqs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	queue := new(ClusterQueue)

	resp, err := cqs.client.Do(req, queue)

	if err != nil {
		return nil, resp, err
	}

	return queue, resp, err
}

func (cqs *ClusterQueuesService) Create(org, clusterID string, qc *ClusterQueueCreate) (*ClusterQueue, *Response, error) {

	if qc == nil {
		return nil, nil, errors.New("ClusterQueueCreate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues", org, clusterID)

	req, err := cqs.client.NewRequest("POST", u, qc)

	if err != nil {
		return nil, nil, err
	}

	queue := new(ClusterQueue)

	resp, err := cqs.client.Do(req, queue)

	if err != nil {
		return nil, resp, err
	}

	return queue, resp, err
}

func (cqs *ClusterQueuesService) Update(org, clusterID, queueID string, qu *ClusterQueueUpdate) (*Response, error) {

	if qu == nil {
		return nil, errors.New("ClusterQueueUpdate struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues/%s", org, clusterID, queueID)

	req, err := cqs.client.NewRequest("PATCH", u, qu)

	if err != nil {
		return nil, err
	}

	resp, err := cqs.client.Do(req, qu)

	if err != nil {
		return resp, err
	}

	return resp, err
}

func (cqs *ClusterQueuesService) Delete(org, clusterID, queueID string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues/%s", org, clusterID, queueID)

	req, err := cqs.client.NewRequest("DELETE", u, nil)

	if err != nil {
		return nil, err
	}

	return cqs.client.Do(req, nil)
}

func (cqs *ClusterQueuesService) Pause(org, clusterID, queueID string, qp *ClusterQueuePause) (*Response, error) {

	if qp == nil {
		return nil, errors.New("ClusterQueuePause struct instance must not be nil")
	}

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues/%s/pause_dispatch", org, clusterID, queueID)

	req, err := cqs.client.NewRequest("POST", u, qp)

	if err != nil {
		return nil, err
	}

	return cqs.client.Do(req, nil)
}

func (cqs *ClusterQueuesService) Resume(org, clusterID, queueID string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/clusters/%s/queues/%s/resume_dispatch", org, clusterID, queueID)

	req, err := cqs.client.NewRequest("POST", u, nil)

	if err != nil {
		return nil, err
	}

	return cqs.client.Do(req, nil)
}
