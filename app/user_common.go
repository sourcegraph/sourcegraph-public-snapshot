package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func getUser(r *http.Request) (*sourcegraph.User, *sourcegraph.UserSpec, error) {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)
	v := mux.Vars(r)

	spec, err := sourcegraph.ParseUserSpec(v["User"])
	if err != nil {
		return nil, nil, err
	}

	p, err := apiclient.Users.Get(ctx, &spec)
	if err != nil {
		return nil, nil, err
	}

	if p.Disabled {
		return nil, nil, &handlerutil.HTTPErr{Status: http.StatusNotFound, Err: fmt.Errorf("user account is disabled")}
	}

	spec.UID = int32(p.UID)

	return p, &spec, nil
}

func personLabel(loginOrEmail string) string {
	if strings.Contains(loginOrEmail, "@") {
		user, _, err := util.SplitEmail(loginOrEmail)
		if err != nil {
			user = "unknown"
		}
		return user + "@…"
	}
	return loginOrEmail
}

func userMetaDescription(p *sourcegraph.User) string {
	var desc string
	if p.Name == "" {
		desc = p.Login
	} else {
		desc += fmt.Sprintf("%s (%s)", p.Name, p.Login)
	}
	desc += " on Sourcegraph"
	return desc
}

func userStat(p *sourcegraph.User, statType string) int {
	// TODO(sqs): this is a stub to make templates and go code
	// compile, it does not actually work - we need to reimplement user
	// stats for this to work.
	return 0
}
