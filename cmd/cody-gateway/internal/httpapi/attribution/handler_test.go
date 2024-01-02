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
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi/attribution"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type fakeActorSource struct {
	name codygateway.ActorSource
}

func (s fakeActorSource) Name() string {
	return string(s.name)
}
func (s fakeActorSource) Get(context.Context, string) (*actor.Actor, error) {
	return &actor.Actor{Source: s, AccessEnabled: true}, nil
}

type fakeDotComGraphQLApi struct{}

func (g fakeDotComGraphQLApi) MakeRequest(
	ctx context.Context,
	req *graphql.Request,
	resp *graphql.Response,
) error {
	return nil
}

func request(t *testing.T) *http.Request {
	requestBody, err := json.Marshal(&attribution.Request{
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
		name: codygateway.ActorSourceProductSubscription,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(ps),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{EnableAttributionSearch: true}
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, fakeDotComGraphQLApi{})
	require.NoError(t, err)
	r := request(t)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected OK, got %d", got)
	}
	var gotResponseBody attribution.Response
	require.NoError(t, json.NewDecoder(w.Body).Decode(&gotResponseBody))
	wantResponseBody := &attribution.Response{
		TotalCount: 0,
		LimitHit:   false,
	}
	if diff := cmp.Diff(wantResponseBody, &gotResponseBody); diff != "" {
		t.Fatalf("unespected response (-want+got):\n%s", diff)
	}
}

func TestFailsForDotcomUsers(t *testing.T) {
	logger := logtest.Scoped(t)
	dotCom := fakeActorSource{
		name: codygateway.ActorSourceDotcomUser,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(dotCom),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{EnableAttributionSearch: true}
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, fakeDotComGraphQLApi{})
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
		name: codygateway.ActorSourceProductSubscription,
	}
	authr := &auth.Authenticator{
		Sources:     actor.NewSources(dotCom),
		Logger:      logger,
		EventLogger: events.NewStdoutLogger(logger),
	}
	config := &httpapi.Config{}
	handler, err := httpapi.NewHandler(logger, nil, nil, nil, authr, nil, config, fakeDotComGraphQLApi{})
	require.NoError(t, err)
	r := request(t)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if got, want := w.Code, http.StatusServiceUnavailable; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected unauthorized, got %d", got)
	}
}
