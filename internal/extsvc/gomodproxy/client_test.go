package gomodproxy

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log
	"golang.org/x/mod/module"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestClient_GetVersion(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "GetVersion", update(t.Name()))

	type result struct {
		Version *module.Version
		Error   string
	}

	var results []result
	for _, tc := range []string{
		"github.com/gorilla/mux", // no version => latest version
		"github.com/tsenart/vegeta/v12@v12.8.4",
		"github.com/Nike-Inc/cerberus-go-client/v3@v3.0.1-ALPHA", // test error + escaping
	} {
		var mod, version string
		if ps := strings.SplitN(tc, "@", 2); len(ps) == 2 {
			mod, version = ps[0], ps[1]
		} else {
			mod = ps[0]
		}
		v, err := cli.GetVersion(ctx, reposource.PackageName(mod), version)
		results = append(results, result{v, fmt.Sprint(err)})
	}

	testutil.AssertGolden(t, "testdata/golden/GetVersions.json", update(t.Name()), results)
}

func TestClient_GetZip(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "GetZip", update(t.Name()))

	type result struct {
		ZipHash  string
		ZipFiles []string
		Error    string
	}

	var results []result
	for _, tc := range []string{
		"github.com/dgryski/go-bloomf@v0.0.0-20220209175004-758619da47c2",
		"github.com/Nike-Inc/cerberus-go-client/v3@v3.0.1-ALPHA", // test error + escaping
	} {
		var mod, version string
		if ps := strings.SplitN(tc, "@", 2); len(ps) == 2 {
			mod, version = ps[0], ps[1]
		} else {
			mod = ps[0]
		}

		zipStream, err := cli.GetZip(ctx, reposource.PackageName(mod), version)
		t.Cleanup(func() {
			if zipStream != nil {
				_ = zipStream.Close()
			}
		})

		var zipBytes []byte
		if zipStream != nil {
			zipBytes, err = io.ReadAll(zipStream)
			if err != nil {
				t.Fatal(err)
			}
		}

		r := result{Error: fmt.Sprint(err)}

		if len(zipBytes) > 0 {
			zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range zr.File {
				r.ZipFiles = append(r.ZipFiles, f.Name)
			}

			h := sha256.Sum256(zipBytes)
			r.ZipHash = hex.EncodeToString(h[:])
		}

		results = append(results, r)
	}

	testutil.AssertGolden(t, "testdata/golden/GetZip.json", update(t.Name()), results)
}

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))
	}
	os.Exit(m.Run())
}

// newTestClient returns a gomodproxy.Client that records its interactions
// to testdata/vcr/.
func newTestClient(t testing.TB, name string, update bool) *Client {
	cassete := filepath.Join("testdata/vcr/", normalize(name))
	rec, err := httptestutil.NewRecorder(cassete, update)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	})

	hc := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec))

	c := &schema.GoModulesConnection{
		Urls: []string{"https://proxy.golang.org"},
	}

	cli := NewClient("urn", c.Urls, hc)
	cli.limiter = ratelimit.NewInstrumentedLimiter("gomod", rate.NewLimiter(100, 10))
	return cli
}

var normalizer = lazyregexp.New("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}
