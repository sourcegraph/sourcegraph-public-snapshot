package buildkite

import "fmt"

// TestRunsService handles communication with test run related
// methods of the Buildkite Test Analytics API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/analytics/runs
type TestRunsService struct {
	client *Client
}

type TestRun struct {
	ID        *string    `json:"id,omitempty" yaml:"id,omitempty"`
	URL       *string    `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL    *string    `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	Branch    *string    `json:"branch,omitempty" yaml:"branch,omitempty"`
	CommitSHA *string    `json:"commit_sha,omitempty" yaml:"commit_sha,omitempty"`
	CreatedAt *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty`
}

type TestRunsListOptions struct {
	ListOptions
}

func (trs *TestRunsService) List(org, slug string, opt *TestRunsListOptions) ([]TestRun, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s/runs", org, slug)

	u, err := addOptions(u, opt)

	if err != nil {
		return nil, nil, err
	}

	req, err := trs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	testRuns := new([]TestRun)

	resp, err := trs.client.Do(req, testRuns)

	if err != nil {
		return nil, resp, err
	}

	return *testRuns, resp, err
}

func (trs *TestRunsService) Get(org, slug, runID string) (*TestRun, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s/runs/%s", org, slug, runID)

	req, err := trs.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	testRun := new(TestRun)

	resp, err := trs.client.Do(req, testRun)

	if err != nil {
		return nil, resp, err
	}

	return testRun, resp, err
}
