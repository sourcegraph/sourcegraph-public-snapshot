package buildkite

import "fmt"

// FlakyTestsService handles communication with flaky test related
// methods of the Buildkite Test Analytics API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/analytics/flaky-tests
type FlakyTestsService struct {
	client *Client
}

type FlakyTest struct {
	ID                   *string    `json:"id,omitempty" yaml:"id,omitempty"`
	WebURL               *string    `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	Scope                *string    `json:"scope,omitempty" yaml:"scope,omitempty"`
	Name                 *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Location             *string    `json:"location,omitempty" yaml:"location,omitempty"`
	FileName             *string    `json:"file_name,omitempty" yaml:"file_name,omitempty"`
	Instances            *int       `json:"instances,omitempty" yaml:"instances,omitempty"`
	MostRecentInstanceAt *Timestamp `json:"most_recent_instance_at,omitempty" yaml:"most_recent_instance_at,omitempty`
}

type FlakyTestsListOptions struct {
	ListOptions
}

func (fts *FlakyTestsService) List(org, slug string, opt *FlakyTestsListOptions) ([]FlakyTest, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s/flaky-tests", org, slug)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := fts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	flakyTests := new([]FlakyTest)

	resp, err := fts.client.Do(req, flakyTests)

	if err != nil {
		return nil, resp, err
	}

	return *flakyTests, resp, err
}
