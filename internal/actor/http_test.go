package actor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestHTTPTransport(t *testing.T) {
	tests := []struct {
		name        string
		actor       *Actor
		wantHeaders map[string]string
	}{{
		name:  "unauthenticated",
		actor: nil,
		wantHeaders: map[string]string{
			headerKeyActorUID: headerValueNoActor,
		},
	}, {
		name:  "internal actor",
		actor: &Actor{Internal: true},
		wantHeaders: map[string]string{
			headerKeyActorUID: headerValueInternalActor,
		},
	}, {
		name:  "user actor",
		actor: &Actor{UID: 1234},
		wantHeaders: map[string]string{
			headerKeyActorUID: "1234",
		},
	}, {
		name:  "user actor with anonymous UID",
		actor: &Actor{UID: 1234, AnonymousUID: "foobar"},
		wantHeaders: map[string]string{
			headerKeyActorUID:          "1234",
			headerKeyActorAnonymousUID: "foobar",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &HTTPTransport{
				RoundTripper: roundTripFunc(func(req *http.Request) *http.Response {
					for k, want := range tt.wantHeaders {
						if got := req.Header.Get(k); got == "" {
							t.Errorf("did not find expected header %q", k)
						} else if diff := cmp.Diff(want, got); diff != "" {
							t.Errorf("headers mismatch (-want +got):\n%s", diff)
						}
					}
					return &http.Response{StatusCode: http.StatusOK}
				}),
			}
			ctx := WithActor(context.Background(), tt.actor)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			got, err := transport.RoundTrip(req)
			if err != nil {
				t.Fatalf("Transport.RoundTrip() error = %v", err)
			}
			if got.StatusCode != http.StatusOK {
				t.Fatalf("Unexpected response: %+v", got)
			}
		})
	}
}

func TestHTTPMiddleware(t *testing.T) {
	tests := []struct {
		name      string
		headers   map[string]string
		wantActor *Actor
	}{{
		name: "unauthenticated",
		headers: map[string]string{
			headerKeyActorUID: headerValueNoActor,
		},
		wantActor: &Actor{}, // FromContext provides a zero-value actor if one is not present
	}, {
		name: "invalid actor",
		headers: map[string]string{
			headerKeyActorUID: "not-a-valid-id",
		},
		wantActor: &Actor{}, // FromContext provides a zero-value actor  if one is not present
	}, {
		name: "internal actor",
		headers: map[string]string{
			headerKeyActorUID: headerValueInternalActor,
		},
		wantActor: &Actor{Internal: true},
	}, {
		name: "user actor",
		headers: map[string]string{
			headerKeyActorUID: "1234",
		},
		wantActor: &Actor{UID: 1234},
	}, {
		name: "no actor info as internal",
		headers: map[string]string{
			headerKeyActorUID: "",
		},
		wantActor: &Actor{Internal: false},
	}, {
		name: "anonymous UID for unauthed actor",
		headers: map[string]string{
			headerKeyActorUID:          "none",
			headerKeyActorAnonymousUID: "anonymousUID",
		},
		wantActor: &Actor{AnonymousUID: "anonymousUID"},
	}, {
		name: "anonymous UID for authed actor",
		headers: map[string]string{
			headerKeyActorUID:          "123",
			headerKeyActorAnonymousUID: "anonymousUID",
		},
		wantActor: &Actor{UID: 123, AnonymousUID: "anonymousUID"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HTTPMiddleware(logtest.Scoped(t), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				got := FromContext(r.Context())
				// Compare string representation
				if diff := cmp.Diff(tt.wantActor.String(), got.String()); diff != "" {
					t.Errorf("actor mismatch (-want +got):\n%s", diff)
				}
			}))
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			handler.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}

func TestAnonymousUIDMiddleware(t *testing.T) {
	t.Run("cookie value is respected", func(t *testing.T) {
		handler := AnonymousUIDMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equal(t, "anon", got.AnonymousUID)
		}))

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "sourcegraphAnonymousUid", Value: "anon"})
		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("header value is respected", func(t *testing.T) {
		handler := AnonymousUIDMiddleware(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equal(t, "anon", got.AnonymousUID)
		}))

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.Header.Set(headerKeyActorAnonymousUID, "anon")
		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("cookie doesn't overwrite existing middleware", func(t *testing.T) {
		handler := http.Handler(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equal(t, int32(132), got.UID)
			require.Equal(t, "anon", got.AnonymousUID) // preserve the anonymous UID!
		}))
		anonHandler := AnonymousUIDMiddleware(handler)
		userHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// Add an authenticated actor
			anonHandler.ServeHTTP(rw, r.WithContext(WithActor(r.Context(), FromUser(132))))
		})

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Name: "sourcegraphAnonymousUid", Value: "anon"})
		userHandler.ServeHTTP(httptest.NewRecorder(), req)
	})
}
