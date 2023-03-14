package azuredevops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *client) GetProject(ctx context.Context, org, project string) (Project, error) {
	reqURL := url.URL{Path: fmt.Sprintf("%s/_apis/projects/%s", org, project)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Project{}, err
	}

	var p Project
	_, err = c.do(ctx, req, "", &p)
	return p, err
}

func (c *client) ListAuthorizedUserProjects(ctx context.Context, org string) ([]Project, error) {
	projects := []Project{}

	continuationToken := ""
	queryParams := make(url.Values)

	reqURL := url.URL{Path: fmt.Sprintf("%s/_apis/projects", org)}

	for {
		if continuationToken != "" {
			queryParams.Set("continuationToken", continuationToken)
		}

		reqURL.RawQuery = queryParams.Encode()

		req, err := http.NewRequest("GET", reqURL.String(), nil)
		if err != nil {
			return nil, err
		}

		var response ListAuthorizedUserProjectsResponse
		continuationToken, err := c.do(ctx, req, "", &response)
		if err != nil {
			return nil, err
		}

		projects = append(projects, response.Value...)

		if continuationToken == "" {
			break
		}
	}
	return projects, nil
}
