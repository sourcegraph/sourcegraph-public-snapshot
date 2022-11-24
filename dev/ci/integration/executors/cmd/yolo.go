package main

// This is a YOLOFILE, it's only there to hack around, @jhchabran will fix it once we POC'ed the whole thing.
//
// TODO: @jhchabran: extract and refactor these things from the codeintel-qa runner and make it available for import here

import (
	"context"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

var (
	SourcegraphEndpoint    = env.Get("SOURCEGRAPH_BASE_URL", "http://127.0.0.1:7080", "Sourcegraph frontend endpoint")
	SourcegraphAccessToken = env.Get("SOURCEGRAPH_SUDO_TOKEN", "123", "Sourcegraph access token with sudo privileges")
)

var (
	client         *gqltestutil.Client
	requestWriter  = &requestResponseWriter{}
	responseWriter = &requestResponseWriter{}
)

func InitializeGraphQLClient() (err error) {
	client, err = gqltestutil.NewClient(SourcegraphEndpoint, requestWriter.Write, responseWriter.Write)
	return err
}

func GraphQLClient() *gqltestutil.Client {
	return client
}

func LastRequestResponsePair() (string, string) {
	return requestWriter.Last(), responseWriter.Last()
}

type requestResponseWriter struct {
	payloads []string
}

func (w *requestResponseWriter) Write(payload []byte) {
	w.payloads = append(w.payloads, string(payload))
}

func (w *requestResponseWriter) Last() string {
	if len(w.payloads) == 0 {
		return ""
	}

	return w.payloads[len(w.payloads)-1]
}

func queryGraphQL(_ context.Context, logger log.Logger, queryName, query string, variables map[string]any, target any) error {
	if err := GraphQLClient().GraphQL(SourcegraphAccessToken, query, variables, target); err != nil {
		logger.Warn("Failed GQL request", log.String("name", queryName), log.String("query", query), log.Error(err))
		return err
	}
	logger.Info("Successful GQL request", log.String("name", queryName), log.String("query", query))
	return nil
}
