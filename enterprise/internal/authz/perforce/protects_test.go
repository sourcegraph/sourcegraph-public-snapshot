package perforce

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

func TestConvertToPostgresMatch(t *testing.T) {
	// Only needs to implement directory-level perforce protects
	tests := []struct {
		name  string
		match string
		want  string
	}{{
		name:  "*",
		match: "//Sourcegraph/Engineering/*/Frontend/",
		want:  "//Sourcegraph/Engineering/[^/]+/Frontend/",
	}, {
		name:  "...",
		match: "//Sourcegraph/Engineering/.../Frontend/",
		want:  "//Sourcegraph/Engineering/%/Frontend/",
	}, {
		name:  "* and ...",
		match: "//Sourcegraph/*/Src/.../Frontend/",
		want:  "//Sourcegraph/[^/]+/Src/%/Frontend/",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToPostgresMatch(tt.match)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestConvertToGlobMatch(t *testing.T) {
	// Should fully implement perforce protects
	// Some cases taken directly from https://www.perforce.com/manuals/cmdref/Content/CmdRef/filespecs.html
	// Useful for debugging:
	//
	//   go run github.com/gobwas/glob/cmd/globdraw -p '{//gra*/dep*/,//gra*/dep*}' -s '/' | dot -Tpng -o pattern.png
	//
	tests := []struct {
		name  string
		match string
		want  string

		shouldMatch    []string
		shouldNotMatch []string
	}{{
		name:  "*",
		match: "//Sourcegraph/Engineering/*/Frontend/",
		want:  "//Sourcegraph/Engineering/*/Frontend/",
	}, {
		name:  "...",
		match: "//Sourcegraph/Engineering/.../Frontend/",
		want:  "//Sourcegraph/Engineering/**/Frontend/",
	}, {
		name:           "* and ...",
		match:          "//Sourcegraph/*/Src/.../Frontend/",
		want:           "//Sourcegraph/*/Src/**/Frontend/",
		shouldMatch:    []string{"//Sourcegraph/Path/Src/One/Two/Frontend/"},
		shouldNotMatch: []string{"//Sourcegraph/One/Two/Src/Path/Frontend/"},
	}, {
		name:  "./....c",
		match: "./....c",
		want:  "./**.c",
		shouldMatch: []string{
			"./file.c", "./dir/file.c",
			// TODO: unsure if this needs to be matched
			// "file.c", "dir/file.c"
		},
	}, {
		name:  "//gra*/dep*",
		match: "//gra*/dep*",
		want:  `//gra*/dep*{/,}`,
		shouldMatch: []string{
			"//graph/depot/", "//graphs/depots",
			// TODO: unsure if this needs to be matched
			// "gravity/deposits",
		},
		shouldNotMatch: []string{"//graph/depot/release1/"},
	}, {
		name:        "//depot/main/rel...",
		match:       "//depot/main/rel...",
		want:        "//depot/main/rel**",
		shouldMatch: []string{"//depot/main/rel/", "//depot/main/releases/", "//depot/main/release-note.txt", "//depot/main/rel1/product1"},
	}, {
		name:        "//app/*",
		match:       "//app/*",
		want:        "//app/*{/,}",
		shouldMatch: []string{"//app/main", "//app/main/"},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToGlobMatch(tt.match)
			if err != nil {
				t.Fatal(fmt.Sprintf("unexpected error: %+v", err))
			}
			if diff := cmp.Diff(tt.want, got.pattern); diff != "" {
				t.Fatal(diff)
			}
			if len(tt.shouldMatch) > 0 {
				for _, m := range tt.shouldMatch {
					if !got.Match(m) {
						t.Errorf("%q should have matched %q", got.pattern, m)
					}
				}
			}
			if len(tt.shouldNotMatch) > 0 {
				for _, m := range tt.shouldNotMatch {
					if got.Match(m) {
						t.Errorf("%q should not have matched %q", got.pattern, m)
					}
				}
			}
		})
	}
}

// mustGlobPattern gets the glob pattern for a given p4 match for use in testing
func mustGlobPattern(t *testing.T, match string) string {
	m, err := convertToGlobMatch(match)
	if err != nil {
		t.Error(err)
	}
	return m.pattern
}

func TestScanFullRepoPermissions(t *testing.T) {
	f, err := os.Open("testdata/sample-protects-u.txt")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	rc := io.NopCloser(bytes.NewReader(data))

	execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
		return rc, nil, nil
	})

	p := NewTestProvider("", "ssl:111.222.333.444:1666", "admin", "password", execer)
	p.depots = []extsvc.RepoID{
		"//app/main/",
		"//app/training/",
		"//app/test/",
		"//app/rickroll/",
		"//not-app/not-main/", // no rules exist
	}
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissions),
	}
	if err := scanProtects(rc, fullRepoPermsScanner(perms, p.depots)); err != nil {
		t.Fatal(err)
	}

	// See sample-protects-u.txt for notes
	want := &authz.ExternalUserPermissions{
		Exacts: []extsvc.RepoID{
			"//app/main/",
			"//app/training/",
			"//app/test/",
		},
		SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissions{
			"//app/main/": {
				PathIncludes: []string{
					mustGlobPattern(t, "//app/main/core/..."),
					mustGlobPattern(t, "//app/main/*/stuff/..."),
					mustGlobPattern(t, "//app/main/frontend/.../stuff/*"),
					mustGlobPattern(t, "//*/main/config.yaml"),
					mustGlobPattern(t, "//app/main/subdir/**"),
					mustGlobPattern(t, "//app/.../README.md"),
					mustGlobPattern(t, "//app/*/dir.yaml"),
				},
				PathExcludes: []string{
					mustGlobPattern(t, "//app/main/subdir/remove/"),
					mustGlobPattern(t, "//app/main/subdir/*/also-remove/..."),
					mustGlobPattern(t, "//.../.secrets.env"),
				},
			},
			"//app/test/": {
				PathIncludes: []string{
					mustGlobPattern(t, "//app/test/..."),
					mustGlobPattern(t, "//app/.../README.md"),
					mustGlobPattern(t, "//app/*/dir.yaml"),
				},
				PathExcludes: []string{
					mustGlobPattern(t, "//.../.secrets.env"),
				},
			},
			"//app/training/": {
				PathIncludes: []string{
					mustGlobPattern(t, "//app/training/..."),
					mustGlobPattern(t, "//app/.../README.md"),
					mustGlobPattern(t, "//app/*/dir.yaml"),
				},
				PathExcludes: []string{
					mustGlobPattern(t, "//app/training/secrets/..."),
					mustGlobPattern(t, "//app/training/.env"),
					mustGlobPattern(t, "//.../.secrets.env"),
				},
			},
		},
	}
	if diff := cmp.Diff(want, perms); diff != "" {
		t.Fatal(diff)
	}
}

func TestScanAllUsers(t *testing.T) {
	ctx := context.Background()
	f, err := os.Open("testdata/sample-protects-a.txt")
	if err != nil {
		t.Fatal(err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	rc := io.NopCloser(bytes.NewReader(data))

	execer := p4ExecFunc(func(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error) {
		return rc, nil, nil
	})

	p := NewTestProvider("", "ssl:111.222.333.444:1666", "admin", "password", execer)
	p.cachedGroupMembers = map[string][]string{
		"dev": {"user1", "user2"},
	}
	p.cachedAllUserEmails = map[string]string{
		"user1": "user1@example.com",
		"user2": "user2@example.com",
	}

	users := make(map[string]struct{})
	if err := scanProtects(rc, allUsersScanner(ctx, p, users)); err != nil {
		t.Fatal(err)
	}
	want := map[string]struct{}{
		"user1": {},
		"user2": {},
	}
	if diff := cmp.Diff(want, users); diff != "" {
		t.Fatal(diff)
	}
}
