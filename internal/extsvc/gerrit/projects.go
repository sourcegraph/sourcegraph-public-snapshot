package gerrit

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *client) listCodeProjects(ctx context.Context, cursor *Pagination) (ListProjectsResponse, bool, error) {
	// Unfortunately Gerrit APIs are quite limited and don't support pagination well.
	// e.g. when we request a list of 100 CODE projects, 100 projects are fetched and
	// only then filtered for CODE projects, possibly returning less than 100 projects.
	// This means we cannot rely on the number of projects returned to determine if
	// there are more projects to fetch.
	// Currently, if you want to only get CODE projects and want to know if there is another page
	// to query for, the only way to do that is to query both CODE and ALL projects and compare
	// the number of projects returned.

	query := make(url.Values)
	query.Set("n", strconv.Itoa(cursor.PerPage))
	query.Set("S", strconv.Itoa((cursor.Page-1)*cursor.PerPage))
	query.Set("type", "CODE")

	uProjects := url.URL{Path: "a/projects/", RawQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, false, err
	}

	// If the number of projects returned is zero we cannot assume that there is no next page.
	// We fetch the first project on the next page of ALL projects and check if that page is empty.
	if len(projects) == 0 {
		nextPageProject, _, err := c.listAllProjects(ctx, &Pagination{PerPage: 1, Skip: cursor.Page * cursor.PerPage})
		if err != nil {
			return nil, false, err
		}
		if len(nextPageProject) == 0 {
			return projects, false, nil
		}
	}

	// Otherwise we always assume that there is a next page.
	return projects, true, nil
}

func (c *client) listAllProjects(ctx context.Context, cursor *Pagination) (ListProjectsResponse, bool, error) {
	query := make(url.Values)
	query.Set("n", strconv.Itoa(cursor.PerPage))
	if cursor.Skip > 0 {
		query.Set("S", strconv.Itoa(cursor.Skip))
	} else {
		query.Set("S", strconv.Itoa((cursor.Page-1)*cursor.PerPage))
	}

	uProjects := url.URL{Path: "a/projects/", RawQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, false, err
	}

	var projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, false, err
	}

	// If the number of returned projects equal the number of requested projects,
	// we assume that there is a next page.
	return projects, len(projects) == cursor.PerPage, nil
}

// ListProjects fetches a list of CODE projects from Gerrit.
func (c *client) ListProjects(ctx context.Context, opts ListProjectsArgs) (projects ListProjectsResponse, nextPage bool, err error) {

	if opts.Cursor == nil {
		opts.Cursor = &Pagination{PerPage: 100, Page: 1}
	}

	if opts.OnlyCodeProjects {
		return c.listCodeProjects(ctx, opts.Cursor)
	}

	return c.listAllProjects(ctx, opts.Cursor)
}
