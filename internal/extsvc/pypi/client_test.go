package pypi

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/internal/unpack"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestDownload(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "Download", update(t.Name()))

	b, err := cli.Project(ctx, "requests")
	if err != nil {
		t.Fatal(err)
	}

	files, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	// Pick the oldest tarball.
	j := -1
	for i, f := range files {
		if path.Ext(f.Name) == ".gz" {
			j = i
			break
		}
	}

	p, err := cli.Download(ctx, files[j].URL)
	if err != nil {
		t.Fatal(err)
	}

	tmp := t.TempDir()
	err = unpack.Tgz(bytes.NewReader(p), tmp, unpack.Opts{})
	if err != nil {
		t.Fatal(err)
	}

	hasher := sha1.New()
	var tarFiles []string

	err = filepath.WalkDir(tmp, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, lErr := os.Open(path)
		if lErr != nil {
			return lErr
		}
		defer f.Close()
		b, lErr := io.ReadAll(f)
		if lErr != nil {
			return lErr
		}
		hasher.Write(b)
		tarFiles = append(tarFiles, strings.TrimPrefix(path, tmp))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	testutil.AssertGolden(t, "testdata/golden/requests", update(t.Name()), struct {
		TarHash string
		Files   []string
	}{
		TarHash: hex.EncodeToString(hasher.Sum(nil)),
		Files:   tarFiles,
	})
}

func TestParse(t *testing.T) {
	cli := newTestClient(t, "Parse", update(t.Name()))
	b, err := cli.Project(context.Background(), "gpg-vault")
	if err != nil {
		t.Fatal(err)
	}
	files, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
	testutil.AssertGolden(t, "testdata/golden/gpg-vault", update(t.Name()), files)
}

func TestParse_empty(t *testing.T) {
	b := []byte(`
<!DOCTYPE html>
<html>
  <body>
  </body>
</html>
`)

	_, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParse_broken(t *testing.T) {
	tmpl, err := template.New("project").Parse(`<!DOCTYPE html>
<html>
  <body>
	{{.Body}}
  </body>
</html>
`)
	if err != nil {
		t.Fatal(err)
	}

	tc := []struct {
		name string
		Body string
	}{
		{
			name: "no text",
			Body: "<a href=\"/frob-1.0.0.tar.gz/\"></a>",
		},
		{
			name: "text does not match base",
			Body: "<a href=\"/frob-1.0.0.tar.gz/\">foo</a>",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			err = tmpl.Execute(&buf, c)
			if err != nil {
				t.Fatal(err)
			}
			_, err := Parse(buf.Bytes())
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestParse_PEP503(t *testing.T) {
	// There may be any other HTML elements on the API pages as long as the required
	// anchor elements exist.
	b := []byte(`
<!DOCTYPE html>
<html>
  <head>
    <meta name="pypi:repository-version" content="1.0">
    <title>Links for frob</title>
  </head>
  <body>
	<h1>Links for frob</h1>
    <a href="/frob-1.0.0.tar.gz/" data-requires-python="&gt;=3">frob-1.0.0.tar.gz</a>
	<h2>More links for frob</h1>
	<div>
	    <a href="/frob-2.0.0.tar.gz/" data-gpg-sig="true">frob-2.0.0.tar.gz</a>
	    <a>frob-3.0.0.tar.gz</a>
	</div>
  </body>
</html>
`)

	got, err := Parse(b)
	if err != nil {
		t.Fatal(err)
	}

	tr := true
	want := []File{
		{
			Name:               "frob-1.0.0.tar.gz",
			URL:                "/frob-1.0.0.tar.gz/",
			DataRequiresPython: ">=3",
		},
		{
			Name:       "frob-2.0.0.tar.gz",
			URL:        "/frob-2.0.0.tar.gz/",
			DataGPGSig: &tr,
		},
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want, +got\n%s", d)
	}
}

// goos: darwin
// goarch: arm64
// pkg: github.com/sourcegraph/sourcegraph/internal/extsvc/pypi
// BenchmarkParse-10           5180            229265 ns/op
func BenchmarkParse(b *testing.B) {
	cli := newTestClient(b, "Download", update("TestDownload"))
	data, err := cli.Project(context.Background(), "requests")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestToWheel(t *testing.T) {
	have := []string{
		"requests-2.16.2-py2.py3-none-any.whl",
		"grpcio-1.46.0rc2-cp39-cp39-win_amd64.whl",
	}
	want := []Wheel{
		{
			Distribution: "requests",
			Version:      "2.16.2",
			BuildTag:     "",
			PythonTag:    "py2.py3",
			ABITag:       "none",
			PlatformTag:  "any",
		},
		{
			Distribution: "grpcio",
			Version:      "1.46.0rc2",
			BuildTag:     "",
			PythonTag:    "cp39",
			ABITag:       "cp39",
			PlatformTag:  "win_amd64",
		},
	}

	var got []Wheel
	for _, h := range have {
		g, err := ToWheel(h)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, *g)
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want, +got\n%s", d)
	}
}

// newTestClient returns a pypi Client that records its interactions
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

	doer, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	c := NewClient("urn", []string{"https://pypi.org/simple"})
	c.cli = doer
	return c
}
