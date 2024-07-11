// Package attribution_test implements a component test for OSS Attribution
// feature of Cody Enterprise, focusing on the gateway component.
//
// ┌───────────┐         ┌────────────┐      ┌─────────┐         ┌────────┐
// │           │ GraphQL │            │ REST │         │ GraphQL │        │
// │ Extension ├────────►│ Enterprise ├─────►│ Gateway ├────────►│ Dotcom │
// │           │         │  instance  │      │         │         │ search │
// └───────────┘         └────────────┘      └─────────┘         └────────┘
//
// !                                  └─── scope of this test ───┘
// Please see RFC 862 for more detailed design consideration and feature scoping:
// https://docs.google.com/document/d/1zSxFDQPxZcn5b6yKx40etpJayoibVzj_Gnugzln1weI/view
package attribution_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
)

type fakeActorSource struct {
	name codygatewayactor.ActorSource
}

func (s fakeActorSource) Name() string {
	return string(s.name)
}
func (s fakeActorSource) Get(context.Context, string) (*actor.Actor, error) {
	return &actor.Actor{Source: s, AccessEnabled: true}, nil
}

// fakeGraphQL is used as test double for dotcom GraphQL API search request.
// The test runs via HTTP layer exercising also GraphQL code-gen.
type fakeGraphQL struct {
	t        *testing.T     // For cleanup and error handling.
	response map[string]any // Nest as deeply as needed.
	url      string         // For client.
}

func (s *fakeGraphQL) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s.response); err != nil {
		s.t.Fatalf("fakeGraphQL.ServeHTTP: %s", err)
	}
}

func runFakeGraphQL(t *testing.T) *fakeGraphQL {
	h := &fakeGraphQL{t: t}
	s := httptest.NewServer(h)
	t.Cleanup(s.Close)
	h.url = s.URL
	return h
}

// request creates an attribution search request to the gateway.
func request(t *testing.T) *http.Request {
	requestBody, err := json.Marshal(&codygateway.AttributionRequest{
		Snippet: strings.Join([]string{
			"for n != 1 {",
			"  if n % 2 == 0 {",
			"    n = n/2",
			"  } else {",
			"    n = 3n+1",
			"  }",
			"}",
		}, "\n"),
		Limit: 2,
	})
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/v1/attribution", bytes.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer sgs_faketoken")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestSuccess(t *testing.T) {
	logger := logtest.Scoped(t)
	ps := fakeActorSource{
		name: codygatewayactor.ActorSourceEnterpriseSubscription,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(ps),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{EnableAttributionSearch: true}
	fakeDotcom := runFakeGraphQL(t)
	fakeDotcom.response = map[string]any{
		"data": map[string]any{
			"snippetAttribution": map[string]any{
				"nodes": []map[string]any{
					{"repositoryName": "github.com/sourcegraph/sourcegraph"},
					{"repositoryName": "github.com/sourcegraph/cody"},
				},
				"totalCount": 2,
				"limitHit":   true,
			},
		},
	}
	dotcomClient := dotcom.NewClient(fakeDotcom.url, "fake auth token", "random", "dev")
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, dotcomClient)
	require.NoError(t, err)
	r := request(t)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected OK, got %d", got)
	}
	var gotResponseBody codygateway.AttributionResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&gotResponseBody))
	wantResponseBody := &codygateway.AttributionResponse{
		Repositories: []codygateway.AttributionRepository{
			{Name: "github.com/sourcegraph/sourcegraph"},
			{Name: "github.com/sourcegraph/cody"},
		},
		TotalCount: 2,
		LimitHit:   true,
	}
	if diff := cmp.Diff(wantResponseBody, &gotResponseBody); diff != "" {
		t.Fatalf("unespected response (-want+got):\n%s", diff)
	}
}

// dummyDotComGraphQLApi is the smallest plumbing to wire up nil graphQL client.
type dummyDotComGraphQLApi struct{}

func (g dummyDotComGraphQLApi) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	return nil
}

func TestFailsForDotcomUsers(t *testing.T) {
	logger := logtest.Scoped(t)
	dotCom := fakeActorSource{
		name: codygatewayactor.ActorSourceDotcomUser,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(dotCom),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{EnableAttributionSearch: true}
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, dummyDotComGraphQLApi{})
	require.NoError(t, err)
	r := request(t)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected unauthorized, got %d", got)
	}
}

func TestUnavailableIfConfigDisabled(t *testing.T) {
	logger := logtest.Scoped(t)
	dotCom := fakeActorSource{
		name: codygatewayactor.ActorSourceEnterpriseSubscription,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(dotCom),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{}
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, dummyDotComGraphQLApi{})
	require.NoError(t, err)
	r := request(t)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusServiceUnavailable; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected unauthorized, got %d", got)
	}
}
