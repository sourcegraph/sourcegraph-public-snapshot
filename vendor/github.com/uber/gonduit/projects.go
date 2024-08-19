package gonduit

import (
	"github.com/uber/gonduit/requests"
	"github.com/uber/gonduit/responses"
)

// ProjectQueryMethod is method name on Phabricator API.
const ProjectQueryMethod = "project.query"

// ProjectQuery performs a call to project.query.
func (c *Conn) ProjectQuery(
	req requests.ProjectQueryRequest,
) (*responses.ProjectQueryResponse, error) {
	var res responses.ProjectQueryResponse

	if err := c.Call(ProjectQueryMethod, &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ProjectSearchMethod is method name on Phabricator API.
const ProjectSearchMethod = "project.search"

// ProjectSearch performs a call to project.search.
func (c *Conn) ProjectSearch(
	req requests.ProjectSearchRequest,
) (*responses.ProjectSearchResponse, error) {
	var res responses.ProjectSearchResponse

	if err := c.Call(ProjectSearchMethod, &req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
