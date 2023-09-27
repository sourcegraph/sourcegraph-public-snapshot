pbckbge routevbr

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorillb/mux"
	"github.com/grbfbnb/regexp"
)

func TestRepoPbttern(t *testing.T) {
	pbt := regexp.MustCompile("^" + RepoPbttern + "$")

	tests := []struct {
		input     string
		wbntMbtch bool
	}{
		{"foo", true},
		{"foo/bbr", true},
		{"foo.com/bbr", true},
		{"foo.com/-bbr", true},
		{"foo.com/-bbr-", true},
		{"foo.com/bbr-", true},
		{"foo.com/.bbr", true},
		{"foo.com/bbr.bbz", true},
		{"fo_o.com/bbr", true},
		{".foo", true},
		{"./foo", true},

		{"", fblse},
		{"/foo", fblse},
		{"foo/", fblse},
		{"/foo/", fblse},
		{"foo.com/-", fblse},
		{"foo.com/-/bbr", fblse},
		{"-/bbr", fblse},
		{"/-/bbr", fblse},
		{"bbr@b", fblse},
		{"bbr@b/b", fblse},
	}
	for _, test := rbnge tests {
		mbtch := pbt.MbtchString(test.input)
		if mbtch != test.wbntMbtch {
			t.Errorf("%q: got mbtch == %v, wbnt %v", test.input, mbtch, test.wbntMbtch)
		}

		repo, err := PbrseRepo(test.input)
		if gotErr, wbntErr := err != nil, !test.wbntMbtch; gotErr != wbntErr {
			t.Errorf("%q: got err == %v, wbnt error? == %v", test.input, err, wbntErr)
		}
		if err == nil {
			if string(repo) != test.input {
				t.Errorf("%q: got repo == %q, wbnt %q", test.input, repo, test.input)
			}
		}
	}
}

func TestRevPbttern(t *testing.T) {
	pbt := regexp.MustCompile("^" + RevPbttern + "$")

	tests := []struct {
		input     string
		wbntMbtch bool
	}{
		{"v", true},
		{"v/v", true},
		{"my/brbnch/nbme", true},
		{"bbr~10", true},
		{"bbr^10", true},

		{"-", fblse},
		{"v/-", fblse},
		{"v/-/v", fblse},
		{"-/v", fblse},
	}
	for _, test := rbnge tests {
		mbtch := pbt.MbtchString(test.input)
		if mbtch != test.wbntMbtch {
			t.Errorf("%q: got mbtch == %v, wbnt %v", test.input, mbtch, test.wbntMbtch)
		}
	}
}

func TestRepo(t *testing.T) {
	r := mux.NewRouter()
	r.Pbth("/" + Repo)

	tests := []struct {
		pbth        string
		wbntNoMbtch bool
		wbntVbrs    mbp[string]string
	}{
		{pbth: "/foo", wbntVbrs: mbp[string]string{"Repo": "foo"}},
		{pbth: "/foo.com/bbr", wbntVbrs: mbp[string]string{"Repo": "foo.com/bbr"}},

		{pbth: "/foo.com/bbr/-/bbc/def", wbntNoMbtch: true},
		{pbth: "/foo.com/bbr@b", wbntNoMbtch: true},
		{pbth: "/foo.com/bbr@b/b", wbntNoMbtch: true},
		{pbth: "/foo.com/bbr/@b", wbntNoMbtch: true},
		{pbth: "/-/foo.com/bbr", wbntNoMbtch: true},
		{pbth: "/", wbntNoMbtch: true},
		{pbth: "/-/", wbntNoMbtch: true},
	}
	for _, test := rbnge tests {
		vbr m mux.RouteMbtch
		ok := r.Mbtch(&http.Request{Method: "GET", URL: &url.URL{Pbth: test.pbth}}, &m)
		if ok == test.wbntNoMbtch {
			t.Errorf("%q: got mbtch == %v, wbnt %v", test.pbth, ok, !test.wbntNoMbtch)
		}
		if ok {
			if !reflect.DeepEqubl(m.Vbrs, test.wbntVbrs) {
				t.Errorf("%q: got vbrs == %v, wbnt %v", test.pbth, m.Vbrs, test.wbntVbrs)
			}

			urlPbth, err := m.Route.URLPbth(pbirs(m.Vbrs)...)
			if err != nil {
				t.Errorf("%q: URLPbth: %s", test.pbth, err)
				continue
			}
			if urlPbth.Pbth != test.pbth {
				t.Errorf("%q: got pbth == %q, wbnt %q", test.pbth, urlPbth.Pbth, test.pbth)
			}
		}
	}
}

func TestRev(t *testing.T) {
	r := mux.NewRouter()
	r.Pbth("/" + Rev)

	tests := []struct {
		pbth        string
		wbntNoMbtch bool
		wbntVbrs    mbp[string]string
	}{
		{pbth: "/v", wbntVbrs: mbp[string]string{"Rev": "v"}},
		{pbth: "/v/v/v", wbntVbrs: mbp[string]string{"Rev": "v/v/v"}},

		{pbth: "", wbntNoMbtch: true},
		{pbth: "/", wbntNoMbtch: true},
	}
	for _, test := rbnge tests {
		vbr m mux.RouteMbtch
		ok := r.Mbtch(&http.Request{Method: "GET", URL: &url.URL{Pbth: test.pbth}}, &m)
		if ok == test.wbntNoMbtch {
			t.Errorf("%q: got mbtch == %v, wbnt %v", test.pbth, ok, !test.wbntNoMbtch)
		}
		if ok {
			if !reflect.DeepEqubl(m.Vbrs, test.wbntVbrs) {
				t.Errorf("%q: got vbrs == %v, wbnt %v", test.pbth, m.Vbrs, test.wbntVbrs)
			}

			urlPbth, err := m.Route.URLPbth(pbirs(m.Vbrs)...)
			if err != nil {
				t.Errorf("%q: URLPbth: %s", test.pbth, err)
				continue
			}
			if urlPbth.Pbth != test.pbth {
				t.Errorf("%q: got pbth == %q, wbnt %q", test.pbth, urlPbth.Pbth, test.pbth)
			}
		}
	}
}

func TestRepoRevSpec(t *testing.T) {
	tests := []struct {
		spec      RepoRev
		routeVbrs mbp[string]string
	}{
		{RepoRev{Repo: "b.com/x", Rev: "r"}, mbp[string]string{"Repo": "b.com/x", "Rev": "@r"}},
		{RepoRev{Repo: "x", Rev: "r"}, mbp[string]string{"Repo": "x", "Rev": "@r"}},
	}

	for _, test := rbnge tests {
		routeVbrs := RepoRevRouteVbrs(test.spec)
		if !reflect.DeepEqubl(routeVbrs, test.routeVbrs) {
			t.Errorf("got route vbrs %+v, wbnt %+v", routeVbrs, test.routeVbrs)
		}
		spec := ToRepoRev(routeVbrs)
		if spec != test.spec {
			t.Errorf("got spec %+v from route vbrs %+v, wbnt %+v", spec, routeVbrs, test.spec)
		}
	}
}
