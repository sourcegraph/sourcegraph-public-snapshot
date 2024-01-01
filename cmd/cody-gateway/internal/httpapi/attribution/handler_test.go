package attribution_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/httpapi"
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
	handler, err := httpapi.NewHandler(nil, nil, nil, nil, authr, nil, &httpapi.Config{})
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/v1/attribution", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer sgs_faketoken")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if got, want := w.Code, http.StatusOK; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected OK, got %d", got)
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
	handler, err := httpapi.NewHandler(nil, nil, nil, nil, authr, nil, &httpapi.Config{})
	require.NoError(t, err)
	req, err := http.NewRequest("POST", "/v1/attribution", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer sgs_faketoken")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Error(w.Body.String())
		t.Fatalf("expected unauthorized, got %d", got)
	}
}
