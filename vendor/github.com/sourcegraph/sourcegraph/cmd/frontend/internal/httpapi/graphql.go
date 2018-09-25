package httpapi

import (
	"errors"
	"net/http"

	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var relayHandler = &relay.Handler{Schema: graphqlbackend.GraphQLSchema}

func serveGraphQL(w http.ResponseWriter, r *http.Request) (err error) {
	if r.Method != "POST" {
		// The URL router should not have routed to this handler if method is not POST, but just in
		// case.
		return errors.New("method must be POST")
	}

	relayHandler.ServeHTTP(w, r)
	return nil
}
