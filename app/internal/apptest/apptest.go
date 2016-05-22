// Package apptest contains a simple framework (client, mock helper,
// etc.) for testing the app and app handlers.
//
// It is intended for use in test code only (not main code), but it
// must be exported so it can be used by other packages.
//
// Because package apptest imports app, test code that uses this
// package will probably need to be in a package with the "_test" name
// suffix.
package apptest

import (
	"net/url"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sqs/pbtypes"
)

// New creates a new app handler and returns a client to access it and
// mocks to control its behavior.
func New() (*httptestutil.Client, *httptestutil.MockClients) {
	app.Init()
	mux := app.NewHandler(router.New(mux.NewRouter()))
	c, mock := httptestutil.NewTest(mux)
	mock.Ctx = conf.WithURL(mock.Ctx, &url.URL{Scheme: "http", Host: "example.com", Path: "/"})
	mock.Ctx = sourcegraph.WithGRPCEndpoint(mock.Ctx, &url.URL{Scheme: "http", Host: "grpc.example.com", Path: "/"})
	mock.Ctx = idkey.NewContext(mock.Ctx, &idkey.IDKey{ID: "k"})

	// Convenience mocks.
	mock.Meta.Config_ = func(context.Context, *pbtypes.Void) (*sourcegraph.ServerConfig, error) {
		return &sourcegraph.ServerConfig{}, nil
	}

	mock.Users.Count_ = func(context.Context, *pbtypes.Void) (*sourcegraph.UserCount, error) {
		return &sourcegraph.UserCount{Count: 1}, nil
	}

	return c, mock
}
