package buildkite

import (
	"fmt"
)

// AgentsService handles communication with the agent related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api/agents
type AgentsService struct {
	client *Client
}

// Agent represents a buildkite build agent.
type Agent struct {
	ID                *string    `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID         *string    `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	URL               *string    `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL            *string    `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	Name              *string    `json:"name,omitempty" yaml:"name,omitempty"`
	ConnectedState    *string    `json:"connection_state,omitempty" yaml:"connection_state,omitempty"`
	AgentToken        *string    `json:"access_token,omitempty" yaml:"access_token,omitempty"`
	Hostname          *string    `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	IPAddress         *string    `json:"ip_address,omitempty" yaml:"ip_address,omitempty"`
	UserAgent         *string    `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`
	Version           *string    `json:"version,omitempty" yaml:"version,omitempty"`
	CreatedAt         *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	LastJobFinishedAt *Timestamp `json:"last_job_finished_at,omitempty" yaml:"last_job_finished_at,omitempty"`
	Priority          *int       `json:"priority,omitempty" yaml:"priority,omitempty"`
	Metadata          []string   `json:"meta_data,omitempty" yaml:"meta_data,omitempty"`

	// the user that created the agent
	Creator *User `json:"creator,omitempty" yaml:"creator,omitempty"`

	Job *Job `json:"job,omitempty" yaml:"job,omitempty"`
}

// AgentListOptions specifies the optional parameters to the
// AgentService.List method.
type AgentListOptions struct {
	// Filters the results by the given agent name
	Name string `url:"name,omitempty"`

	// Filters the results by the given hostname
	Hostname string `url:"hostname,omitempty"`

	// Filters the results by the given exact version number
	Version string `url:"version,omitempty"`

	ListOptions
}

// List the agents for a given orginisation.
//
// buildkite API docs: https://buildkite.com/docs/api/agents#list-agents
func (as *AgentsService) List(org string, opt *AgentListOptions) ([]Agent, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/agents", org)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := as.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	agents := new([]Agent)
	resp, err := as.client.Do(req, agents)
	if err != nil {
		return nil, resp, err
	}

	return *agents, resp, err
}

// Get fetches an agent.
//
// buildkite API docs: https://buildkite.com/docs/api/agents#get-an-agent
func (as *AgentsService) Get(org string, id string) (*Agent, *Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/agents/%s", org, id)

	req, err := as.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	agent := new(Agent)
	resp, err := as.client.Do(req, agent)
	if err != nil {
		return nil, resp, err
	}

	return agent, resp, err
}

// Create a new buildkite agent.
//
// buildkite API docs: https://buildkite.com/docs/api/agents#create-an-agent
func (as *AgentsService) Create(org string, agent *Agent) (*Agent, *Response, error) {

	var u string

	u = fmt.Sprintf("v2/organizations/%s/agents", org)

	req, err := as.client.NewRequest("POST", u, agent)
	if err != nil {
		return nil, nil, err
	}

	a := new(Agent)
	resp, err := as.client.Do(req, a)
	if err != nil {
		return nil, resp, err
	}

	return a, resp, err
}

// Delete an agent.
//
// buildkite API docs: https://buildkite.com/docs/api/agents#delete-an-agent
func (as *AgentsService) Delete(org string, id string) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/agents/%s", org, id)

	req, err := as.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return nil, err
	}

	return as.client.Do(req, nil)
}

// Stop an agent.
//
// buildkite API docs: https://buildkite.com/docs/apis/rest-api/agents#stop-an-agent
func (as *AgentsService) Stop(org string, id string, force bool) (*Response, error) {

	u := fmt.Sprintf("v2/organizations/%s/agents/%s/stop", org, id)

	var body = struct {
		Force bool `json:"force"`
	}{force}

	req, err := as.client.NewRequest("PUT", u, body)
	if err != nil {
		return nil, err
	}

	return as.client.Do(req, nil)
}
