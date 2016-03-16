package app

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func getUser(ctx context.Context, r *http.Request) (*sourcegraph.User, *sourcegraph.UserSpec, error) {
	ctx, cl := handlerutil.Client(r)
	v := mux.Vars(r)

	spec, err := sourcegraph.ParseUserSpec(v["User"])
	if err != nil {
		return nil, nil, err
	}

	p, err := cl.Users.Get(ctx, &spec)
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
