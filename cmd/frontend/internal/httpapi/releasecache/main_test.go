package releasecache

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func newClientFactory(t testing.TB, name string) (*httpcli.Factory, func(testing.TB)) {
	cassetteName := filepath.Join("testdata", strings.ReplaceAll(name, " ", "-"))
	rec := newRecorder(t, cassetteName, update(name))
	return httpcli.NewFactory(httpcli.NewMiddleware(), httptestutil.NewRecorderOpt(rec)),
		func(t testing.TB) { save(t, rec) }
}

func save(t testing.TB, rec *recorder.Recorder) {
	if err := rec.Stop(); err != nil {
		t.Errorf("failed to update test data: %s", err)
	}
}

func newRecorder(t testing.TB, file string, record bool) *recorder.Recorder {
	rec, err := httptestutil.NewRecorder(file, record, func(i *cassette.Interaction) error {
		// The ratelimit.Monitor type resets its internal timestamp if it's
		// updated with a timestamp in the past. This makes tests ran with
		// recorded interations just wait for a very long time. Removing
		// these headers from the cassette effectively disables rate-limiting
		// in tests which replay HTTP interactions, which is desired behaviour.
		for _, name := range [...]string{
			"RateLimit-Limit",
			"RateLimit-Observed",
			"RateLimit-Remaining",
			"RateLimit-Reset",
			"RateLimit-Resettime",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		} {
			i.Response.Headers.Del(name)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	return rec
}
