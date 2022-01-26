package npm

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"flag"
	"io"
	"os"
	"sort"
	"testing"

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
	client = NewHTTPClient("https://registry.npmjs.org", &rateLimit)
	doer, err := recorderFactory.Doer()
	require.Nil(t, err)
	client.doer = doer
	return client, stop
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
