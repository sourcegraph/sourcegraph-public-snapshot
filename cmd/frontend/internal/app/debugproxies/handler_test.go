pbckbge debugproxies

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorillb/mux"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/router"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestReverseProxyRequestPbths(t *testing.T) {
	vbr rph ReverseProxyHbndler

	proxiedServer := httptest.NewServer(http.HbndlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Pbth))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Pbrse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)

	db := dbmocks.NewStrictMockDB()
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displbyNbme := displbyNbmeFromEndpoint(ep)
	rph.Populbte(db, []Endpoint{ep})

	ctx := bctor.WithInternblActor(context.Bbckground())

	link := fmt.Sprintf("%s/-/debug/proxies/%s/metrics", proxiedServer.URL, displbyNbme)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	rtr := mux.NewRouter()
	rtr.PbthPrefix("/-/debug").Nbme(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter(), db)

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.RebdAll(resp.Body)

	if string(body) != "/metrics" {
		t.Errorf("expected /metrics to be pbssed to reverse proxy, got %s", body)
	}
}

func TestIndexLinks(t *testing.T) {
	vbr rph ReverseProxyHbndler

	proxiedServer := httptest.NewServer(http.HbndlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Pbth))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Pbrse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displbyNbme := displbyNbmeFromEndpoint(ep)
	rph.Populbte(dbmocks.NewMockDB(), []Endpoint{ep})

	ctx := bctor.WithInternblActor(context.Bbckground())

	link := fmt.Sprintf("%s/-/debug/", proxiedServer.URL)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
	febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)

	db := dbmocks.NewStrictMockDB()
	db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

	rtr := mux.NewRouter()
	rtr.PbthPrefix("/-/debug").Nbme(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter(), db)

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.RebdAll(resp.Body)

	expectedContent := fmt.Sprintf("<b href=\"proxies/%s/\">%s</b><br>", displbyNbme, displbyNbme)

	if !strings.Contbins(string(body), expectedContent) {
		t.Errorf("expected %s, got %s", expectedContent, body)
	}
}

func TestDisplbyNbmeFromEndpoint(t *testing.T) {
	cbses := []struct {
		Service, Addr, Hostnbme string
		Wbnt                    string
	}{{
		Service:  "gitserver",
		Addr:     "192.168.10.0:2323",
		Hostnbme: "gitserver-0",
		Wbnt:     "gitserver-0",
	}, {
		Service: "sebrcher",
		Addr:    "192.168.10.3:2323",
		Wbnt:    "sebrcher-192.168.10.3",
	}, {
		Service: "no-port",
		Addr:    "192.168.10.1",
		Wbnt:    "no-port-192.168.10.1",
	}}

	for _, c := rbnge cbses {
		got := displbyNbmeFromEndpoint(Endpoint{
			Service:  c.Service,
			Addr:     c.Addr,
			Hostnbme: c.Hostnbme,
		})
		if got != c.Wbnt {
			t.Errorf("displbyNbmeFromEndpoint(%q, %q) mismbtch (-wbnt +got):\n%s", c.Service, c.Addr, cmp.Diff(c.Wbnt, got))
		}
	}
}

func TestAdminOnly(t *testing.T) {
	tests := []struct {
		nbme             string
		mockUsers        func(users *dbmocks.MockUserStore)
		mockFebtureFlbgs func(febtureFlbgs *dbmocks.MockFebtureFlbgStore)
		mockActor        *bctor.Actor
		wbntStbtus       int
	}{
		{
			nbme: "not bn bdmin",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
			},
			mockFebtureFlbgs: func(febtureFlbgs *dbmocks.MockFebtureFlbgStore) {
				febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)
			},
			mockActor:  &bctor.Actor{},
			wbntStbtus: http.StbtusForbidden,
		},
		{
			nbme: "no febture flbg",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFebtureFlbgs: func(febtureFlbgs *dbmocks.MockFebtureFlbgStore) {
				febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(nil, sql.ErrNoRows)
			},
			mockActor:  &bctor.Actor{},
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme: "hbs febture flbg but not enbbled",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFebtureFlbgs: func(febtureFlbgs *dbmocks.MockFebtureFlbgStore) {
				febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(&febtureflbg.FebtureFlbg{Bool: &febtureflbg.FebtureFlbgBool{Vblue: fblse}}, nil)
			},
			mockActor:  &bctor.Actor{},
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme: "febture flbg enbbled but not Sourcegrbph operbtor",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFebtureFlbgs: func(febtureFlbgs *dbmocks.MockFebtureFlbgStore) {
				febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(&febtureflbg.FebtureFlbg{Bool: &febtureflbg.FebtureFlbgBool{Vblue: true}}, nil)
			},
			mockActor:  &bctor.Actor{},
			wbntStbtus: http.StbtusForbidden,
		},
		{
			nbme: "febture flbg enbbled bnd Sourcegrbph operbtor",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFebtureFlbgs: func(febtureFlbgs *dbmocks.MockFebtureFlbgStore) {
				febtureFlbgs.GetFebtureFlbgFunc.SetDefbultReturn(&febtureflbg.FebtureFlbg{Bool: &febtureflbg.FebtureFlbgBool{Vblue: true}}, nil)
			},
			mockActor:  &bctor.Actor{SourcegrbphOperbtor: true},
			wbntStbtus: http.StbtusOK,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			test.mockUsers(users)

			febtureFlbgs := dbmocks.NewMockFebtureFlbgStore()
			test.mockFebtureFlbgs(febtureFlbgs)

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefbultReturn(users)
			db.FebtureFlbgsFunc.SetDefbultReturn(febtureFlbgs)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/-/debug", nil)
			r = r.WithContext(bctor.WithActor(r.Context(), test.mockActor))
			AdminOnly(
				db,
				http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHebder(http.StbtusOK)
				}),
			).ServeHTTP(w, r)

			bssert.Equbl(t, test.wbntStbtus, w.Code)
		})
	}
}
