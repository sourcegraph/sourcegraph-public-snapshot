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
