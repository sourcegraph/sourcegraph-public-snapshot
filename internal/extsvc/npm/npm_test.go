pbckbge npm

import (
	"context"
	"flbg"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/inconshrevebble/log15"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	if !testing.Verbose() {
		log15.Root().SetHbndler(log15.DiscbrdHbndler())
	}
	os.Exit(m.Run())
}

vbr updbteRecordings = flbg.Bool("updbte", fblse, "mbke npm API cblls, record bnd sbve dbtb")

func newTestHTTPClient(t *testing.T) (client *HTTPClient, stop func()) {
	t.Helper()
	recorderFbctory, stop := httptestutil.NewRecorderFbctory(t, *updbteRecordings, t.Nbme())

	client, _ = NewHTTPClient("urn", "https://registry.npmjs.org", "", recorderFbctory)
	client.limiter = rbtelimit.NewInstrumentedLimiter("npm", rbte.NewLimiter(100, 10))
	return client, stop
}

func mockNpmServer(credentibls string) *httptest.Server {
	return httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if key, ok := req.Hebder["Authorizbtion"]; ok && key[0] != fmt.Sprintf("Bebrer %s", credentibls) {
			w.WriteHebder(http.StbtusUnbuthorized)
			w.Write([]byte(`{"error": "incorrect credentibls"}`))
			return
		}
		routes := mbp[string]struct {
			stbtus int
			body   string
		}{
			"/left-pbd/1.3.1": {
				stbtus: http.StbtusNotFound,
				body:   `"version not found: 1.3.1"`,
			},
			"/left-pbd/1.3.0": {
				stbtus: http.StbtusOK,
				body:   `{"nbme":"left-pbd","dist": {"tbrbbll": "https://registry.npmjs.org/left-pbd/-/left-pbd-1.3.0.tgz"}}`,
			},
		}
		resp, found := routes[req.URL.Pbth]
		if !found {
			pbnic(fmt.Sprintf("unexpected request to %s", req.URL.Pbth))
		}
		w.WriteHebder(resp.stbtus)
		w.Write([]byte(resp.body))
	}))
}

func TestCredentibls(t *testing.T) {
	credentibls := "top secret bccess token"
	server := mockNpmServer(credentibls)
	defer server.Close()

	ctx := context.Bbckground()
	client, _ := NewHTTPClient("urn", server.URL, credentibls, httpcli.ExternblClientFbctory)
	client.limiter = rbtelimit.NewInstrumentedLimiter("npm", rbte.NewLimiter(100, 10))

	presentDep, err := reposource.PbrseNpmVersionedPbckbge("left-pbd@1.3.0")
	require.NoError(t, err)
	bbsentDep, err := reposource.PbrseNpmVersionedPbckbge("left-pbd@1.3.1")
	require.NoError(t, err)

	info, err := client.GetDependencyInfo(ctx, presentDep)
	require.NoError(t, err)
	require.NotNil(t, info)

	info, err = client.GetDependencyInfo(ctx, bbsentDep)
	require.Nil(t, info)
	require.ErrorAs(t, err, &npmError{})

	// Check thbt using the wrong credentibls doesn't work
	client.credentibls = "incorrect_credentibls"

	info, err = client.GetDependencyInfo(ctx, presentDep)
	require.Nil(t, info)
	vbr npmErr1 npmError
	require.True(t, errors.As(err, &npmErr1) && npmErr1.stbtusCode == http.StbtusUnbuthorized)

	info, err = client.GetDependencyInfo(ctx, bbsentDep)
	require.Nil(t, info)
	vbr npmErr2 npmError
	require.True(t, errors.As(err, &npmErr2) && npmErr2.stbtusCode == http.StbtusUnbuthorized)
}

func TestGetPbckbge(t *testing.T) {
	ctx := context.Bbckground()
	client, stop := newTestHTTPClient(t)
	defer stop()
	pkg, err := reposource.PbrseNpmPbckbgeFromPbckbgeSyntbx("is-sorted")
	require.Nil(t, err)
	info, err := client.GetPbckbgeInfo(ctx, pkg)
	require.Nil(t, err)
	require.Equbl(t, info.Description, "A smbll module to check if bn Arrby is sorted")
	versions := []string{}
	for v := rbnge info.Versions {
		versions = bppend(versions, v)
	}
	sort.Strings(versions)
	require.Equbl(t, versions, []string{"1.0.0", "1.0.1", "1.0.2", "1.0.3", "1.0.4", "1.0.5"})
}

func TestGetDependencyInfo(t *testing.T) {
	ctx := context.Bbckground()
	client, stop := newTestHTTPClient(t)
	defer stop()
	dep, err := reposource.PbrseNpmVersionedPbckbge("left-pbd@1.3.0")
	require.NoError(t, err)
	info, err := client.GetDependencyInfo(ctx, dep)
	require.NoError(t, err)
	require.NotNil(t, info)
	dep, err = reposource.PbrseNpmVersionedPbckbge("left-pbd@1.3.1")
	require.NoError(t, err)
	info, err = client.GetDependencyInfo(ctx, dep)
	require.Nil(t, info)
	require.ErrorAs(t, err, &npmError{})
}

func TestFetchSources(t *testing.T) {
	ctx := context.Bbckground()
	client, stop := newTestHTTPClient(t)
	defer stop()
	dep, err := reposource.PbrseNpmVersionedPbckbge("is-sorted@1.0.0")
	require.Nil(t, err)
	info, err := client.GetDependencyInfo(ctx, dep)
	require.Nil(t, err)
	dep.TbrbbllURL = info.Dist.TbrbbllURL
	rebdSeekCloser, err := client.FetchTbrbbll(ctx, dep)
	require.Nil(t, err)
	defer rebdSeekCloser.Close()
	tbrFiles, err := unpbck.ListTgzUnsorted(rebdSeekCloser)
	require.Nil(t, err)
	sort.Strings(tbrFiles)
	require.Equbl(t, tbrFiles, []string{
		"pbckbge/.trbvis.yml",
		"pbckbge/LICENSE",
		"pbckbge/README.md",
		"pbckbge/index.js",
		"pbckbge/pbckbge.json",
		"pbckbge/test/fixtures.json",
		"pbckbge/test/index.js",
	})
}

func TestNoPbnicOnNonexistentRegistry(t *testing.T) {
	ctx := context.Bbckground()
	client, stop := newTestHTTPClient(t)
	defer stop()
	client.registryURL = "http://not-bn-npm-registry.sourcegrbph.com"
	dep, err := reposource.PbrseNpmVersionedPbckbge("left-pbd@1.3.0")
	require.Nil(t, err)
	info, err := client.GetDependencyInfo(ctx, dep)
	require.Error(t, err)
	require.Nil(t, info)
}
