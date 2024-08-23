package buildkite

import (
	"errors"
	"fmt"
)

// TestSuitesService handles communication with the test suite related
// methods of the Buildkite Test Analytics API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/analytics/suites
type TestSuitesService struct {
	client *Client
}

type TestSuiteCreate struct {
	Name          string   `json:"name" yaml:"name"`
	DefaultBranch string   `json:"default_branch,omitempty" yaml:"default_branch,omitempty"`
	ShowAPIToken  bool     `json:"show_api_token,omitempty" yaml:"show_api_token,omitempty"`
	TeamUUIDs     []string `json:"team_ids,omitempty" yaml:"team_ids,omitempty"`
}

type TestSuite struct {
	ID            *string `json:"id,omitempty" yaml:"id,omitempty"`
	GraphQLID     *string `json:"graphql_id,omitempty" yaml:"graphql_id,omitempty"`
	Slug          *string `json:"slug,omitempty" yaml:"slug,omitempty"`
	Name          *string `json:"name,omitempty" yaml:"name,omitempty"`
	URL           *string `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL        *string `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	DefaultBranch *string `json:"default_branch,omitempty" yaml:"default_branch,omitempty"`
}

type TestSuiteListOptions struct {
	ListOptions
}

func (tss *TestSuitesService) List(org string, opt *TestSuiteListOptions) ([]TestSuite, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites", org)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := tss.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	testSuites := new([]TestSuite)

	resp, err := tss.client.Do(req, testSuites)

	if err != nil {
		return nil, resp, err
	}

	return *testSuites, resp, err
}

func (tss *TestSuitesService) Get(org, slug string) (*TestSuite, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s", org, slug)

	req, err := tss.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	testSuite := new(TestSuite)

	resp, err := tss.client.Do(req, testSuite)

	if err != nil {
		return nil, resp, err
	}

	return testSuite, resp, err
}

func (tss *TestSuitesService) Create(org string, ts *TestSuiteCreate) (*TestSuite, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites", org)

	req, err := tss.client.NewRequest("POST", u, ts)

	if err != nil {
		return nil, nil, err
	}

	testSuite := new(TestSuite)
	resp, err := tss.client.Do(req, testSuite)

	if err != nil {
		return nil, resp, err
	}

	return testSuite, resp, err
}

func (tss *TestSuitesService) Update(org string, ts *TestSuite) (*Response, error) {

	if ts == nil {
		return nil, errors.New("Test suite must not be nil")
	}

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s", org, *ts.Slug)

	req, err := tss.client.NewRequest("PATCH", u, ts)

	if err != nil {
		return nil, nil
	}

	resp, err := tss.client.Do(req, ts)

	if err != nil {
		return resp, err
	}

	return resp, err
}

func (tss *TestSuitesService) Delete(org, slug string) (*Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s", org, slug)

	req, err := tss.client.NewRequest("DELETE", u, nil)

	if err != nil {
		return nil, err
	}

	return tss.client.Do(req, nil)
}
