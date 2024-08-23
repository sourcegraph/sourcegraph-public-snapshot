package buildkite

import "fmt"

// TestsService handles communication with test related
// methods of the Buildkite Test Analytics API.
//
// Buildkite API docs: https://buildkite.com/docs/apis/rest-api/analytics/tests
type TestsService struct {
	client *Client
}

type Test struct {
	ID       *string `json:"id,omitempty" yaml:"id,omitempty"`
	URL      *string `json:"url,omitempty" yaml:"url,omitempty"`
	WebURL   *string `json:"web_url,omitempty" yaml:"web_url,omitempty"`
	Scope    *string `json:"scope,omitempty" yaml:"scope,omitempty"`
	Name     *string `json:"name,omitempty" yaml:"name,omitempty"`
	Location *string `json:"location,omitempty" yaml:"location,omitempty"`
	FileName *string `json:"file_name,omitempty" yaml:"file_name,omitempty"`
}

func (ts *TestsService) Get(org, slug, testID string) (*Test, *Response, error) {

	u := fmt.Sprintf("v2/analytics/organizations/%s/suites/%s/tests/%s", org, slug, testID)

	req, err := ts.client.NewRequest("GET", u, nil)

	if err != nil {
		return nil, nil, err
	}

	test := new(Test)

	resp, err := ts.client.Do(req, test)

	if err != nil {
		return nil, resp, err
	}

	return test, resp, err
}
