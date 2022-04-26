package pypi

import (
	"context"
	"flag"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
)

var updateRegex = flag.String("update", "", "Update testdata of tests matching the given regex")

func update(name string) bool {
	if updateRegex == nil || *updateRegex == "" {
		return false
	}
	return regexp.MustCompile(*updateRegex).MatchString(name)
}

func TestRelease(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "release", update(t.Name()))

	result, err := cli.Release(ctx, "hsf", "1.1.0")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update=", update(t.Name()))
	testutil.AssertGolden(t, "testdata/golden/release.json", update(t.Name()), result)
}

func TestProject(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "project", update(t.Name()))

	result, err := cli.project(ctx, "tiny")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("update=", update(t.Name()))
	testutil.AssertGolden(t, "testdata/golden/project.json", update(t.Name()), result)
}

func TestGetArchive(t *testing.T) {
	ctx := context.Background()
	cli := newTestClient(t, "GetArchive", update(t.Name()))

	u, err := cli.GetArchive(ctx, "hsf", "1.1.0")
	if err != nil {
		t.Fatal(err)
	}

	err = u.Unpack(t.TempDir())
	if err != nil {
		t.Fatal()
	}
}

func mkProjectInfo(t *testing.T, urls []string) *ProjectInfo {
	t.Helper()
	releaseURLs := make([]ReleaseURL, 0, len(urls))
	for _, u := range urls {
		var t string
		if filepath.Ext(u) == ".whl" {
			t = "bdist_wheel"
		} else {
			t = "sdist"
		}
		releaseURLs = append(releaseURLs, ReleaseURL{
			URL:         u,
			PackageType: t,
		})
	}

	return &ProjectInfo{
		Info: Info{
			Name:    "deadbeef",
			Version: "1.0",
		},
		URLS: releaseURLs,
	}
}

func TestSelectURL(t *testing.T) {
	tc := []struct {
		name string
		have *ProjectInfo
		want string
	}{
		{
			name: "if no tarball prefer any",
			have: mkProjectInfo(t, []string{
				"https://files.pythonhosted.org/packages/4b/8d/5da53dbf3530bfc19bcf81cca3109bdfce22f94e6599ee3db223b0c338e7/grpcio-1.44.0-cp39-cp39-win32.whl",
				"https://files.pythonhosted.org/packages/e2/29/059a2555a38dacd066a7a83fa7f735d1b6ff6b1cde3e62cdd09597dbe7be/grpcio-1.44.0-cp39-cp39-win_amd64.whl",
				"https://files.pythonhosted.org/packages/2d/61/08076519c80041bc0ffa1a8af0cbd3bf3e2b62af10435d269a9d0f40564d/requests-2.27.1-py2.py3-none-any.whl",
			}),
			want: "https://files.pythonhosted.org/packages/2d/61/08076519c80041bc0ffa1a8af0cbd3bf3e2b62af10435d269a9d0f40564d/requests-2.27.1-py2.py3-none-any.whl",
		},
		{
			name: "pick tarball",
			have: mkProjectInfo(t, []string{
				"https://files.pythonhosted.org/packages/2d/61/08076519c80041bc0ffa1a8af0cbd3bf3e2b62af10435d269a9d0f40564d/requests-2.27.1-py2.py3-none-any.whl",
				"https://files.pythonhosted.org/packages/60/f3/26ff3767f099b73e0efa138a9998da67890793bfa475d8278f84a30fec77/requests-2.27.1.tar.gz",
			}),
			want: "https://files.pythonhosted.org/packages/60/f3/26ff3767f099b73e0efa138a9998da67890793bfa475d8278f84a30fec77/requests-2.27.1.tar.gz",
		},
		{
			name: "pick first",
			have: mkProjectInfo(t, []string{
				"https://files.pythonhosted.org/packages/e2/29/059a2555a38dacd066a7a83fa7f735d1b6ff6b1cde3e62cdd09597dbe7be/grpcio-1.44.0-cp39-cp39-win_amd64.whl",
				"https://files.pythonhosted.org/packages/4b/8d/5da53dbf3530bfc19bcf81cca3109bdfce22f94e6599ee3db223b0c338e7/grpcio-1.44.0-cp39-cp39-win32.whl",
			}),
			want: "https://files.pythonhosted.org/packages/e2/29/059a2555a38dacd066a7a83fa7f735d1b6ff6b1cde3e62cdd09597dbe7be/grpcio-1.44.0-cp39-cp39-win_amd64.whl",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			got, err := selectURL(c.have)
			if err != nil {
				t.Fatal(err)
			}
			if got.URL != c.want {
				t.Fatalf("\nwant: %s\ngot : %s\n", c.want, got.URL)
			}
		})
	}
}

func TestGetPlatform(t *testing.T) {
	have := []string{
		"https://files.pythonhosted.org/packages/48/ff/f46afbaf2f7400476443de2a916a63c7959bda7b701b06cd0c5305b1cdcf/grpcio-1.44.0-cp37-cp37m-win32.whl",
		"https://files.pythonhosted.org/packages/3f/89/4920de0ec2ccbcb26e3f8941c20ba797807d49b11d4424d71cf001c646f7/grpcio-1.44.0-cp37-cp37m-win_amd64.whl",
		"https://files.pythonhosted.org/packages/c5/7a/35c847b757de67bef35b058049481ad970a4110094cc1ca523e6e9db46cb/grpcio-1.44.0-cp38-cp38-linux_armv7l.whl",
		"https://files.pythonhosted.org/packages/4e/05/9975d04ecb33d94e2aa9fd60d6201d2faa294be74f14ec4cf048579fe315/grpcio-1.44.0-cp38-cp38-macosx_10_10_x86_64.whl",
		"https://files.pythonhosted.org/packages/c7/bc/9513390807ce0265c5a072345bd57da90173ea9f82d5744a25e433a5386a/grpcio-1.44.0-cp38-cp38-manylinux2010_i686.whl",
		"https://files.pythonhosted.org/packages/b4/6f/40ae3f63c944705d5ce31cec02dd8a5bf1898ebdf828269b8da97c47adfa/grpcio-1.44.0-cp38-cp38-manylinux2010_x86_64.whl",
		"https://files.pythonhosted.org/packages/ab/4a/d3b3bcdd3eba0b9a9b0bb9752a813fb856df6aa5061da3a2900f33c50762/grpcio-1.44.0-cp38-cp38-manylinux_2_17_aarch64.whl",
		"https://files.pythonhosted.org/packages/f7/10/47060cd72e190f4cbe1a579cc5c15667b3a3fc5b479619198b570e4b5838/grpcio-1.44.0-cp38-cp38-manylinux_2_17_i686.manylinux2014_i686.whl",
		"https://files.pythonhosted.org/packages/5b/92/a00eed89bae16e48644f514c842b1cc6deaf0f79cb7dcfeda2dc514e11af/grpcio-1.44.0-cp38-cp38-manylinux_2_17_x86_64.manylinux2014_x86_64.whl",
		"https://files.pythonhosted.org/packages/61/0c/dcd6f8a86de1aa29c5f597be8a21311eeaa24b0fd6f040efe60000c58a38/grpcio-1.44.0-cp38-cp38-win32.whl",
		"https://files.pythonhosted.org/packages/d5/a1/0f893b9e639c89181d519e20f10373a176e666936feb634980c19f0132fb/grpcio-1.44.0-cp38-cp38-win_amd64.whl",
		"https://files.pythonhosted.org/packages/90/f6/900f63bd7d78db62713f8599d98030d1af1a43accd173a3a0e2b5954ab4b/grpcio-1.44.0-cp39-cp39-linux_armv7l.whl",
		"https://files.pythonhosted.org/packages/23/02/324dcaacfc7c11c00e86f03f318092407663eb52e9684ef0c44af72a3857/grpcio-1.44.0-cp39-cp39-macosx_10_10_x86_64.whl",
		"https://files.pythonhosted.org/packages/d0/cc/ec20d7d492b0b2d7bebebc19e3ec13ea3b70ec5ad943b1fd3f464e9f262c/grpcio-1.44.0-cp39-cp39-manylinux2010_i686.whl",
		"https://files.pythonhosted.org/packages/f2/c9/30632461724b7a3ca859b6726e7ab70e8d792d4e08da06045735f7a8d283/grpcio-1.44.0-cp39-cp39-manylinux2010_x86_64.whl",
		"https://files.pythonhosted.org/packages/10/3f/0533846f6448c73c0c9eaa57bbd7bf076e9b13dedb1fc7ad22e2acdfdd0c/grpcio-1.44.0-cp39-cp39-manylinux_2_17_aarch64.whl",
		"https://files.pythonhosted.org/packages/0b/6a/9c41d50f0c39c6da19bc9c2d27743dca9447d0f73f80f78989f45c3c0b6b/grpcio-1.44.0-cp39-cp39-manylinux_2_17_i686.manylinux2014_i686.whl",
		"https://files.pythonhosted.org/packages/55/e8/7de0fdd337632f7ed5bd267472f067702d4089b58bc06e5739511380b9af/grpcio-1.44.0-cp39-cp39-manylinux_2_17_x86_64.manylinux2014_x86_64.whl",
		"https://files.pythonhosted.org/packages/4b/8d/5da53dbf3530bfc19bcf81cca3109bdfce22f94e6599ee3db223b0c338e7/grpcio-1.44.0-cp39-cp39-win32.whl",
		"https://files.pythonhosted.org/packages/e2/29/059a2555a38dacd066a7a83fa7f735d1b6ff6b1cde3e62cdd09597dbe7be/grpcio-1.44.0-cp39-cp39-win_amd64.whl",
		"https://files.pythonhosted.org/packages/2d/61/08076519c80041bc0ffa1a8af0cbd3bf3e2b62af10435d269a9d0f40564d/requests-2.27.1-py2.py3-none-any.whl",
	}
	want := []string{
		"win32",
		"win_amd64",
		"linux_armv7l",
		"macosx_10_10_x86_64",
		"manylinux2010_i686",
		"manylinux2010_x86_64",
		"manylinux_2_17_aarch64",
		"manylinux_2_17_i686.manylinux2014_i686",
		"manylinux_2_17_x86_64.manylinux2014_x86_64",
		"win32",
		"win_amd64",
		"linux_armv7l",
		"macosx_10_10_x86_64",
		"manylinux2010_i686",
		"manylinux2010_x86_64",
		"manylinux_2_17_aarch64",
		"manylinux_2_17_i686.manylinux2014_i686",
		"manylinux_2_17_x86_64.manylinux2014_x86_64",
		"win32",
		"win_amd64",
		"any",
	}

	got := make([]string, 0, len(have))
	for _, h := range have {
		got = append(got, getPlatform(h))
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("-want, +got:\n%s", d)
	}
}

// TODO: add fuzz test with go 1.18
//func FuzzGetPlatform(f *testing.F) {
//	testcases := []string{
//		"foo-bar.whl",
//		"none-any.whl",
//		"cp39-cp39-manylinux_2_17_aarch64.whl",
//		"bas",
//		"",
//		".whl-",
//		"-",
//		".-.-.-",
//	}
//	for _, tc := range testcases {
//		f.Add(tc)
//	}
//	f.Fuzz(func(t *testing.T, url string) {
//		_ = getPlatform(url)
//	})
//}

// TODO: Use real world examples for each case.
func TestToArchiveType(t *testing.T) {
	tc := []struct {
		have string
		want archiveType
	}{
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.zip",
			want: zip,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.tar.gz",
			want: gztar,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.tar.bz2",
			want: bztar,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.tar.xz",
			want: xztar,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.tar.Z",
			want: ztar,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.tar",
			want: tar,
		},
		{
			have: "https://files.pythonhosted.org/packages/60/f3/26/requests-2.27.1.exe",
			want: "",
		},
	}

	for _, c := range tc {
		t.Run(string(c.want), func(t *testing.T) {
			if got := toArchiveType(c.have); got != c.want {
				t.Fatalf("want %s, got \"%s\"", c.want, got)
			}
		})
	}
}

// newTestClient returns a pypi Client that records its interactions
// to testdata/vcr/.
func newTestClient(t testing.TB, name string, update bool) *Client {
	// TODO: do I need normalize?
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

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	return NewClient("urn", []string{"https://pypi.org"}, hc)
}

var normalizer = lazyregexp.New("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}
