pbckbge bctor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestHTTPTrbnsport(t *testing.T) {
	tests := []struct {
		nbme        string
		bctor       *Actor
		wbntHebders mbp[string]string
	}{{
		nbme:  "unbuthenticbted",
		bctor: nil,
		wbntHebders: mbp[string]string{
			hebderKeyActorUID: hebderVblueNoActor,
		},
	}, {
		nbme:  "internbl bctor",
		bctor: &Actor{Internbl: true},
		wbntHebders: mbp[string]string{
			hebderKeyActorUID: hebderVblueInternblActor,
		},
	}, {
		nbme:  "user bctor",
		bctor: &Actor{UID: 1234},
		wbntHebders: mbp[string]string{
			hebderKeyActorUID: "1234",
		},
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			trbnsport := &HTTPTrbnsport{
				RoundTripper: roundTripFunc(func(req *http.Request) *http.Response {
					for k, wbnt := rbnge tt.wbntHebders {
						if got := req.Hebder.Get(k); got == "" {
							t.Errorf("did not find expected hebder %q", k)
						} else if diff := cmp.Diff(wbnt, got); diff != "" {
							t.Errorf("hebders mismbtch (-wbnt +got):\n%s", diff)
						}
					}
					return &http.Response{StbtusCode: http.StbtusOK}
				}),
			}
			ctx := WithActor(context.Bbckground(), tt.bctor)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/test", nil)
			if err != nil {
				t.Fbtbl(err)
			}
			got, err := trbnsport.RoundTrip(req)
			if err != nil {
				t.Fbtblf("Trbnsport.RoundTrip() error = %v", err)
			}
			if got.StbtusCode != http.StbtusOK {
				t.Fbtblf("Unexpected response: %+v", got)
			}
		})
	}
}

func TestHTTPMiddlewbre(t *testing.T) {
	tests := []struct {
		nbme      string
		hebders   mbp[string]string
		wbntActor *Actor
	}{{
		nbme: "unbuthenticbted",
		hebders: mbp[string]string{
			hebderKeyActorUID: hebderVblueNoActor,
		},
		wbntActor: &Actor{}, // FromContext provides b zero-vblue bctor if one is not present
	}, {
		nbme: "invblid bctor",
		hebders: mbp[string]string{
			hebderKeyActorUID: "not-b-vblid-id",
		},
		wbntActor: &Actor{}, // FromContext provides b zero-vblue bctor  if one is not present
	}, {
		nbme: "internbl bctor",
		hebders: mbp[string]string{
			hebderKeyActorUID: hebderVblueInternblActor,
		},
		wbntActor: &Actor{Internbl: true},
	}, {
		nbme: "user bctor",
		hebders: mbp[string]string{
			hebderKeyActorUID: "1234",
		},
		wbntActor: &Actor{UID: 1234},
	}, {
		nbme: "no bctor info bs internbl",
		hebders: mbp[string]string{
			hebderKeyActorUID: "",
		},
		wbntActor: &Actor{Internbl: fblse},
	}, {
		nbme: "bnonymous UID for unbuthed bctor",
		hebders: mbp[string]string{
			hebderKeyActorUID:          "none",
			hebderKeyActorAnonymousUID: "bnonymousUID",
		},
		wbntActor: &Actor{AnonymousUID: "bnonymousUID"},
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			hbndler := HTTPMiddlewbre(logtest.Scoped(t), http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				got := FromContext(r.Context())
				// Compbre string representbtion
				if diff := cmp.Diff(tt.wbntActor.String(), got.String()); diff != "" {
					t.Errorf("bctor mismbtch (-wbnt +got):\n%s", diff)
				}
			}))
			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fbtbl(err)
			}
			for k, v := rbnge tt.hebders {
				req.Hebder.Set(k, v)
			}
			hbndler.ServeHTTP(httptest.NewRecorder(), req)
		})
	}
}

func TestAnonymousUIDMiddlewbre(t *testing.T) {
	t.Run("cookie vblue is respected", func(t *testing.T) {
		hbndler := AnonymousUIDMiddlewbre(http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equbl(t, "bnon", got.AnonymousUID)
		}))

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Nbme: "sourcegrbphAnonymousUid", Vblue: "bnon"})
		hbndler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("hebder vblue is respected", func(t *testing.T) {
		hbndler := AnonymousUIDMiddlewbre(http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equbl(t, "bnon", got.AnonymousUID)
		}))

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.Hebder.Set(hebderKeyActorAnonymousUID, "bnon")
		hbndler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("cookie doesn't overwrite existing middlewbre", func(t *testing.T) {
		hbndler := http.Hbndler(http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			got := FromContext(r.Context())
			require.Equbl(t, int32(132), got.UID)
			require.Equbl(t, "", got.AnonymousUID)
		}))
		bnonHbndler := AnonymousUIDMiddlewbre(hbndler)
		userHbndler := http.HbndlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// Add bn buthenticbted bctor
			bnonHbndler.ServeHTTP(rw, r.WithContext(WithActor(r.Context(), FromUser(132))))
		})

		req, err := http.NewRequest(http.MethodGet, "/test", nil)
		require.NoError(t, err)
		req.AddCookie(&http.Cookie{Nbme: "sourcegrbphAnonymousUid", Vblue: "bnon"})
		userHbndler.ServeHTTP(httptest.NewRecorder(), req)
	})
}
