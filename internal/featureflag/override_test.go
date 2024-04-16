package featureflag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestOverrides(t *testing.T) {
	setupRedisTest(t)

	mockStore := NewMockStore()
	mockStore.GetUserFlagsFunc.SetDefaultHook(func(_ context.Context, uid int32) (map[string]bool, error) {
		if uid != 123 {
			return nil, errors.New("BOOM")
		}
		return map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"user":       true,
		}, nil
	})
	mockStore.GetAnonymousUserFlagsFunc.SetDefaultHook(func(_ context.Context, anonUID string) (map[string]bool, error) {
		if anonUID != "123" {
			return nil, errors.New("BOOM")
		}
		return map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"anon":       true,
		}, nil
	})
	mockStore.GetGlobalFeatureFlagsFunc.SetDefaultHook(func(_ context.Context) (map[string]bool, error) {
		return map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"global":     true,
		}, nil
	})

	handler := Middleware(mockStore, http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(FromContext(r.Context()).flags)
	})))

	cases := []struct {
		Name     string
		Header   []string
		RawQuery string
		Want     map[string]bool
	}{{
		Name: "unset",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": false,
		},
	}, {
		Name:   "header-true",
		Header: []string{"feat-false"},
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": true,
		},
	}, {
		Name:   "header-false",
		Header: []string{"-feat-true"},
		Want: map[string]bool{
			"feat-true":  false,
			"feat-false": false,
		},
	}, {
		Name:   "header-multiple",
		Header: []string{"feat-false", "-new-false", "new-true"},
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": true,
			"new-false":  false,
			"new-true":   true,
		},
	}, {
		Name:     "query-true",
		RawQuery: "feat=feat-false",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": true,
		},
	}, {
		Name:     "query-false",
		RawQuery: "feat=-feat-true",
		Want: map[string]bool{
			"feat-true":  false,
			"feat-false": false,
		},
	}, {
		Name:     "query-multiple",
		RawQuery: "feat=feat-false&feat=-new-false&feat=new-true",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": true,
			"new-false":  false,
			"new-true":   true,
		},
	}, {
		Name:     "prefer-header",
		Header:   []string{"header"},
		RawQuery: "feat=query",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"header":     true,
		},
	}, {
		Name:   "header-flat-map-values",
		Header: []string{"  feat-false", "-new-false,new-true", " a b,c,  d "},
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": true,
			"new-false":  false,
			"new-true":   true,
			"a":          true,
			"b":          true,
			"c":          true,
			"d":          true,
		},
	}, {
		Name:     "empty-param",
		RawQuery: "feat=",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": false,
		},
	}, {
		Name:     "query-client-only-unset",
		RawQuery: "feat=~foo,bar,-baz",
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"bar":        true,
			"baz":        false,
		},
	}, {
		Name:   "header-client-only-unset",
		Header: []string{"~foo", "bar", "-baz"},
		Want: map[string]bool{
			"feat-true":  true,
			"feat-false": false,
			"bar":        true,
			"baz":        false,
		},
	}}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			run := func(act *actor.Actor) map[string]bool {
				t.Helper()

				req, err := http.NewRequest(http.MethodGet, "/test?"+tc.RawQuery, nil)
				if err != nil {
					t.Fatal(err)
				}
				req = req.WithContext(actor.WithActor(context.Background(), act))
				if tc.Header != nil {
					req.Header["X-Sourcegraph-Override-Feature"] = tc.Header
				}

				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)

				var flags map[string]bool
				err = json.NewDecoder(w.Result().Body).Decode(&flags)
				if err != nil {
					t.Fatal(err)
				}
				return flags
			}

			got := run(actor.FromUser(123))
			want := flagsWithKey(tc.Want, "user")
			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("unexpected flags for user (-want, +got):\n%s", d)
			}

			got = run(actor.FromAnonymousUser("123"))
			want = flagsWithKey(tc.Want, "anon")
			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("unexpected flags for anonymous (-want, +got):\n%s", d)
			}

			got = run(&actor.Actor{})
			want = flagsWithKey(tc.Want, "global")
			if d := cmp.Diff(want, got); d != "" {
				t.Errorf("unexpected flags for global (-want, +got):\n%s", d)
			}
		})
	}
}

func flagsWithKey(flags map[string]bool, key string) map[string]bool {
	m := map[string]bool{key: true}
	for k, v := range flags {
		m[k] = v
	}
	return m
}
