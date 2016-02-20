package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func getUser(ctx context.Context, r *http.Request) (*sourcegraph.User, *sourcegraph.UserSpec, error) {
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
		return nil, nil, &errcode.HTTPErr{Status: http.StatusNotFound, Err: fmt.Errorf("user account is disabled")}
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
