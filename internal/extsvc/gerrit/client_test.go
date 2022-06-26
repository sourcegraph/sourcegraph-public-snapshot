package gerrit

import (
	"context"
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update testdata")

func TestClient_ListProjects(t *testing.T) {
	cli, save := NewTestClient(t, "ListProjects", *update)
	defer save()

	ctx := context.Background()

	args := ListProjectsArgs{
		Cursor: &Pagination{PerPage: 5, Page: 1},
	}

	resp, _, err := cli.ListProjects(ctx, args)
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/ListProjects.json", true, resp)
}

func TestClient_GetProjectAccessPermissions(t *testing.T) {
	cli, save := NewTestClient(t, "GetProjectAccessPermissions", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.GetProjectAccessPermissions(ctx, "sourcegraph/create")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetProjectAccessPermissions.json", *update, resp)
}

func TestClient_GetGroupByName(t *testing.T) {
	cli, save := NewTestClient(t, "GetGroupByName", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.GetGroupByName(ctx, "Sourcegraph")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/GetGroupByName.json", *update, resp)
}

func TestClient_ListGroupMembers(t *testing.T) {
	cli, save := NewTestClient(t, "ListGroupMembers", *update)
	defer save()

	ctx := context.Background()

	resp, err := cli.ListGroupMembers(ctx, "40bb3b32f9da57a0697285daabb777965720b96e")
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/ListGroupMembers.json", *update, resp)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
	os.Exit(m.Run())
}

// NewTestClient returns a gerrit.Client that records its interactions
// to testdata/vcr/.
func NewTestClient(t testing.TB, name string, update bool) (*Client, func()) {
	t.Helper()
	cassete := filepath.Join("testdata/vcr/", normalize(name))
	rec, err := httptestutil.NewRecorder(cassete, update)
	if err != nil {
		t.Fatal(err)
	}
	rec.SetMatcher(ignoreHostMatcher)

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	instanceURL := os.Getenv("GERRIT_URL")
	if instanceURL == "" {
		instanceURL = "https://gerrit.sgdev.org"
	}

	c := &schema.GerritConnection{
		Url:      instanceURL,
		Username: os.Getenv("GERRIT_USERNAME"),
		Password: os.Getenv("GERRIT_PASSWORD"),
	}
	cli, err := NewClient("urn", c, hc)
	if err != nil {
		t.Fatal(err)
	}

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

var normalizer = lazyregexp.New("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}

func ignoreHostMatcher(r *http.Request, i cassette.Request) bool {
	if r.Method != i.Method {
		return false
	}
	u, err := url.Parse(i.URL)
	if err != nil {
		return false
	}
	u.Host = r.URL.Host
	u.Scheme = r.URL.Scheme
	return r.URL.String() == u.String()
}
