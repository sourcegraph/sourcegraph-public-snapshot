package npm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

var updateRecordings = flag.Bool("update", false, "make NPM API calls, record and save data")

func newTestHTTPClient(t *testing.T) (client *HTTPClient, stop func()) {
	t.Helper()
	recorderFactory, stop := httptestutil.NewRecorderFactory(t, *updateRecordings, t.Name())
	rateLimit := schema.NPMRateLimit{true, 1000}
	client = NewHTTPClient("https://registry.npmjs.org", &rateLimit, "")
	doer, err := recorderFactory.Doer()
	require.Nil(t, err)
	client.doer = doer
	return client, stop
}

func mockNPMServer(credentials string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if key, ok := req.Header["Authorization"]; ok && key[0] != fmt.Sprintf("Bearer %s", credentials) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "incorrect credentials"}`))
			return
		}
		routes := map[string]struct {
			status int
			body   string
		}{
			"/left-pad/1.3.1": {
				status: http.StatusNotFound,
				body:   `"version not found: 1.3.1"`,
			},
			"/left-pad/1.3.0": {
				status: http.StatusOK,
				body:   `{"name":"left-pad","dist": {"tarball": "https://registry.npmjs.org/left-pad/-/left-pad-1.3.0.tgz"}}`,
			},
		}
		resp, found := routes[req.URL.Path]
		if !found {
			panic(fmt.Sprintf("unexpected request to %s", req.URL.Path))
		}
		w.WriteHeader(resp.status)
		w.Write([]byte(resp.body))
	}))
}

func TestCredentials(t *testing.T) {
	credentials := "top secret access token"
	server := mockNPMServer(credentials)
	defer server.Close()

	ctx := context.Background()
	rateLimit := schema.NPMRateLimit{true, 1000}
	client := NewHTTPClient(server.URL, &rateLimit, credentials)

	presentDep, err := reposource.ParseNPMDependency("left-pad@1.3.0")
	require.Nil(t, err)
	absentDep, err := reposource.ParseNPMDependency("left-pad@1.3.1")
	require.Nil(t, err)

	exists, err := client.DoesDependencyExist(ctx, *presentDep)
	require.Nil(t, err)
	require.True(t, exists)

	exists, _ = client.DoesDependencyExist(ctx, *absentDep)
	require.False(t, exists)

	// Check that using the wrong credentials doesn't work
	client.credentials = "incorrect_credentials"

	_, err = client.DoesDependencyExist(ctx, *presentDep)
	var npmErr1 npmError
	require.True(t, errors.As(err, &npmErr1) && npmErr1.statusCode == http.StatusUnauthorized)

	_, err = client.DoesDependencyExist(ctx, *absentDep)
	var npmErr2 npmError
	require.True(t, errors.As(err, &npmErr2) && npmErr2.statusCode == http.StatusUnauthorized)
}

func TestAvailablePackageVersions(t *testing.T) {
	ctx := context.Background()
	client, stop := newTestHTTPClient(t)
	defer stop()
	pkg, err := reposource.ParseNPMPackageFromPackageSyntax("is-sorted")
	require.Nil(t, err)
	versionMap, err := client.AvailablePackageVersions(ctx, *pkg)
	require.Nil(t, err)
	versions := []string{}
	for v := range versionMap {
		versions = append(versions, v)
	}
	sort.Strings(versions)
	require.Equal(t, versions, []string{"1.0.0", "1.0.1", "1.0.2", "1.0.3", "1.0.4", "1.0.5"})
}

func TestDoesDependencyExist(t *testing.T) {
	ctx := context.Background()
	client, stop := newTestHTTPClient(t)
	defer stop()
	dep, err := reposource.ParseNPMDependency("left-pad@1.3.0")
	require.Nil(t, err)
	exists, err := client.DoesDependencyExist(ctx, *dep)
	require.Nil(t, err)
	require.True(t, exists)
	dep, err = reposource.ParseNPMDependency("left-pad@1.3.1")
	require.Nil(t, err)
	exists, _ = client.DoesDependencyExist(ctx, *dep)
	require.False(t, exists)
}

func TestFetchSources(t *testing.T) {
	ctx := context.Background()
	client, stop := newTestHTTPClient(t)
	defer stop()
	dep, err := reposource.ParseNPMDependency("is-sorted@1.0.0")
	require.Nil(t, err)
	readSeekCloser, err := client.FetchTarball(ctx, *dep)
	require.Nil(t, err)
	defer readSeekCloser.Close()
	gzipReader, err := gzip.NewReader(readSeekCloser)
	require.Nil(t, err)
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	tarFiles := []string{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		require.Nil(t, err)
		tarFiles = append(tarFiles, header.Name)
	}
	sort.Strings(tarFiles)
	require.Equal(t, tarFiles, []string{
		"package/.travis.yml",
		"package/LICENSE",
		"package/README.md",
		"package/index.js",
		"package/package.json",
		"package/test/fixtures.json",
		"package/test/index.js",
	})
}

func TestNoPanicOnNonexistentRegistry(t *testing.T) {
	ctx := context.Background()
	client, stop := newTestHTTPClient(t)
	defer stop()
	client.registryURL = "http://not-an-npm-registry.sourcegraph.com"
	dep, err := reposource.ParseNPMDependency("left-pad@1.3.0")
	require.Nil(t, err)
	_, err = client.DoesDependencyExist(ctx, *dep)
	require.NotNil(t, err)
}
