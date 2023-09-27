pbckbge bzuredevops

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

func (c *client) GetProject(ctx context.Context, org, project string) (Project, error) {
	reqURL := url.URL{Pbth: fmt.Sprintf("%s/_bpis/projects/%s", org, project)}

	req, err := http.NewRequest("GET", reqURL.String(), nil)
	if err != nil {
		return Project{}, err
	}

	vbr p Project
	_, err = c.do(ctx, req, "", &p)
	return p, err
}
