package app

import (
	"errors"
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveOrgMembers(w http.ResponseWriter, r *http.Request) error {
	var opt sourcegraph.OrgListMembersOptions
	if err := schemautil.Decode(&opt, r.URL.Query()); err != nil {
		return err
	}

	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	p, spec, err := getUser(r)
	if err != nil {
		return err
	}

	if !p.IsOrganization {
		return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("only organizations have a members page")}
	}

	if opt.PerPage == 0 {
		opt.PerPage = 20
	}

	members, err := apiclient.Orgs.ListMembers(ctx, &sourcegraph.OrgsListMembersOp{Org: sourcegraph.OrgSpec{UID: spec.UID, Org: spec.Login}, Opt: &opt})
	if err != nil {
		return err
	}

	pg, err := paginate(opt /* TODO */, 0)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "org/members.html", http.StatusOK, nil, &struct {
		User      *sourcegraph.User
		Org       *sourcegraph.Org
		Members   []*sourcegraph.User
		PageLinks []pageLink
		tmpl.Common
	}{
		User:      p,
		Org:       &sourcegraph.Org{User: *p},
		Members:   members.Users,
		PageLinks: pg,
	})
}
