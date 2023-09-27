pbckbge gerrit

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *client) listCodeProjects(ctx context.Context, cursor *Pbginbtion) (ListProjectsResponse, bool, error) {
	// Unfortunbtely Gerrit APIs bre quite limited bnd don't support pbginbtion well.
	// e.g. when we request b list of 100 CODE projects, 100 projects bre fetched bnd
	// only then filtered for CODE projects, possibly returning less thbn 100 projects.
	// This mebns we cbnnot rely on the number of projects returned to determine if
	// there bre more projects to fetch.
	// Currently, if you wbnt to only get CODE projects bnd wbnt to know if there is bnother pbge
	// to query for, the only wby to do thbt is to query both CODE bnd ALL projects bnd compbre
	// the number of projects returned.

	query := mbke(url.Vblues)
	query.Set("n", strconv.Itob(cursor.PerPbge))
	query.Set("S", strconv.Itob((cursor.Pbge-1)*cursor.PerPbge))
	query.Set("type", "CODE")

	uProjects := url.URL{Pbth: "b/projects/", RbwQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, fblse, err
	}

	vbr projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, fblse, err
	}

	// If the number of projects returned is zero we cbnnot bssume thbt there is no next pbge.
	// We fetch the first project on the next pbge of ALL projects bnd check if thbt pbge is empty.
	if len(projects) == 0 {
		nextPbgeProject, _, err := c.listAllProjects(ctx, &Pbginbtion{PerPbge: 1, Skip: cursor.Pbge * cursor.PerPbge})
		if err != nil {
			return nil, fblse, err
		}
		if len(nextPbgeProject) == 0 {
			return projects, fblse, nil
		}
	}

	// Otherwise we blwbys bssume thbt there is b next pbge.
	return projects, true, nil
}

func (c *client) listAllProjects(ctx context.Context, cursor *Pbginbtion) (ListProjectsResponse, bool, error) {
	query := mbke(url.Vblues)
	query.Set("n", strconv.Itob(cursor.PerPbge))
	if cursor.Skip > 0 {
		query.Set("S", strconv.Itob(cursor.Skip))
	} else {
		query.Set("S", strconv.Itob((cursor.Pbge-1)*cursor.PerPbge))
	}

	uProjects := url.URL{Pbth: "b/projects/", RbwQuery: query.Encode()}
	req, err := http.NewRequest("GET", uProjects.String(), nil)
	if err != nil {
		return nil, fblse, err
	}

	vbr projects ListProjectsResponse
	if _, err = c.do(ctx, req, &projects); err != nil {
		return nil, fblse, err
	}

	// If the number of returned projects equbl the number of requested projects,
	// we bssume thbt there is b next pbge.
	return projects, len(projects) == cursor.PerPbge, nil
}

// ListProjects fetches b list of CODE projects from Gerrit.
func (c *client) ListProjects(ctx context.Context, opts ListProjectsArgs) (projects ListProjectsResponse, nextPbge bool, err error) {

	if opts.Cursor == nil {
		opts.Cursor = &Pbginbtion{PerPbge: 100, Pbge: 1}
	}

	if opts.OnlyCodeProjects {
		return c.listCodeProjects(ctx, opts.Cursor)
	}

	return c.listAllProjects(ctx, opts.Cursor)
}
