package perforce

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	srp "github.com/sourcegraph/sourcegraph/internal/authz/subrepoperms"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	p4types "github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/schema"
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
		},
	}, {
		name:  "//gra*/dep*",
		match: "//gra*/dep*",
		want:  `//gra*/dep*{/,}`,
		shouldMatch: []string{
			"//graph/depot/", "//graphs/depots",
		},
		shouldNotMatch: []string{"//graph/depot/release1/"},
	}, {
		name:        "//depot/main/rel...",
		match:       "//depot/main/rel...",
		want:        "//depot/main/rel**",
		shouldMatch: []string{"//depot/main/rel/", "//depot/main/releases/", "//depot/main/release-note.txt", "//depot/main/rel1/product1"},
	}, {
		name:        "//depot/*",
		match:       "//depot/*",
		want:        "//depot/*{/,}",
		shouldMatch: []string{"//depot/main", "//depot/main/"},
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

func mustGlob(t *testing.T, match string) globMatch {
	m, err := convertToGlobMatch(match)
	if err != nil {
		t.Error(err)
	}
	return m
}

// mustGlobPattern gets the glob pattern for a given p4 match for use in testing
func mustGlobPattern(t *testing.T, match string) string {
	return mustGlob(t, match).pattern
}

func TestMatchesAgainstDepot(t *testing.T) {
	type args struct {
		match globMatch
		depot string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{{
		name: "simple match",
		args: args{
			match: mustGlob(t, "//depot/main/..."),
			depot: "//depot/main/",
		},
		want: true,
	}, {
		name: "no wildcard in match",
		args: args{
			match: mustGlob(t, "//depot/"),
			depot: "//depot/main/",
		},
		want: false,
	}, {
		name: "match parent path",
		args: args{
			match: mustGlob(t, "//depot/..."),
			depot: "//depot/main/",
		},
		want: true,
	}, {
		name: "match sub path with all wildcard",
		args: args{
			match: mustGlob(t, "//depot/.../file"),
			depot: "//depot/main/",
		},
		want: true,
	}, {
		name: "match sub path with dir wildcard",
		args: args{
			match: mustGlob(t, "//depot/*/file"),
			depot: "//depot/main/",
		},
		want: true,
	}, {
		name: "match sub path with dir and all wildcards",
		args: args{
			match: mustGlob(t, "//depot/*/file/.../path"),
			depot: "//depot/main/",
		},
		want: true,
	}, {
		name: "match sub path with dir wildcard that's deeply nested",
		args: args{
			match: mustGlob(t, "//depot/*/file/*/another-file/path/"),
			depot: "//depot/main/",
		},
		want: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesAgainstDepot(tt.args.match, tt.args.depot); got != tt.want {
				t.Errorf("matchesAgainstDepot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanFullRepoPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
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

	db := dbmocks.NewMockDB()

	p := NewProvider(logger, db, gitserver.NewStrictMockClient(), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
	p.depots = []extsvc.RepoID{
		"//depot/main/",
		"//depot/training/",
		"//depot/test/",
		"//depot/rickroll/",
		"//not-depot/not-main/", // no rules exist
	}
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
	}
	if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), fullRepoPermsScanner(logger, perms, p.depots), false); err != nil {
		t.Fatal(err)
	}

	// See sample-protects-u.txt for notes
	want := &authz.ExternalUserPermissions{
		Exacts: []extsvc.RepoID{
			"//depot/main/",
			"//depot/training/",
			"//depot/test/",
		},
		SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs{
			"//depot/main/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "-/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/base/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/*/stuff/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/frontend/.../stuff/*"), IP: "*"},
					{Path: mustGlobPattern(t, "/config.yaml"), IP: "*"},
					{Path: mustGlobPattern(t, "/subdir/**"), IP: "*"},
					{Path: mustGlobPattern(t, "-/subdir/remove/"), IP: "*"},
					{Path: mustGlobPattern(t, "/subdir/some-dir/also-remove/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/subdir/another-dir/also-remove/..."), IP: "*"},
					{Path: mustGlobPattern(t, "-/subdir/*/also-remove/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/.../README.md"), IP: "*"},
					{Path: mustGlobPattern(t, "/dir.yaml"), IP: "*"},
					{Path: mustGlobPattern(t, "-/.../.secrets.env"), IP: "*"},
				},
			},
			"//depot/test/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "*"},
					{Path: mustGlobPattern(t, "/.../README.md"), IP: "*"},
					{Path: mustGlobPattern(t, "/dir.yaml"), IP: "*"},
					{Path: mustGlobPattern(t, "-/.../.secrets.env"), IP: "*"},
				},
			},
			"//depot/training/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "*"},
					{Path: mustGlobPattern(t, "-/secrets/..."), IP: "*"},
					{Path: mustGlobPattern(t, "-/.env"), IP: "*"},
					{Path: mustGlobPattern(t, "/.../README.md"), IP: "*"},
					{Path: mustGlobPattern(t, "/dir.yaml"), IP: "*"},
					{Path: mustGlobPattern(t, "-/.../.secrets.env"), IP: "*"},
				},
			},
		},
	}
	if diff := cmp.Diff(want, perms); diff != "" {
		t.Fatal(diff)
	}
}

func TestScanIPPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdata/sample-protects-ip.txt")
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

	db := dbmocks.NewMockDB()

	p := NewProvider(logger, db, gitserver.NewStrictMockClient(), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
	p.depots = []extsvc.RepoID{
		"//depot/src/",
		"//depot/project1/",
		"//depot/project2/",
		"//depot/local/",
		"//depot/test/",
		"//depot/rickroll/",
		"//not-depot/not-main/", // no rules exist
	}

	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
	}
	if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), fullRepoPermsScanner(logger, perms, p.depots), false); err != nil {
		t.Fatal(err)
	}
	want := &authz.ExternalUserPermissions{
		Exacts: []extsvc.RepoID{
			"//depot/src/",
			"//depot/project1/",
			"//depot/project2/",
			"//depot/local/",
			"//depot/test/",
			"//depot/rickroll/",
			"//not-depot/not-main/",
		},
		SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs{
			"//depot/src/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "*"},
					{Path: mustGlobPattern(t, "-/main/..."), IP: "192.168.10.0/24"},
					{Path: mustGlobPattern(t, "-/main/..."), IP: "[2001:db8:16:81::]/48"},
					{Path: mustGlobPattern(t, "/main/..."), IP: "proxy-192.168.10.0/24"},
					{Path: mustGlobPattern(t, "/main/..."), IP: "proxy-[2001:db8:16:81::]/48"},
					{Path: mustGlobPattern(t, "-/dev/..."), IP: "proxy-10.0.0.0/8"},
					{Path: mustGlobPattern(t, "-/dev/..."), IP: "proxy-[2001:db8:1008::]/32"},
					{Path: mustGlobPattern(t, "/dev/..."), IP: "10.0.0.0/8"},
					{Path: mustGlobPattern(t, "/dev/..."), IP: "[2001:db8:1008::]/32"},
				},
			},
			"//depot/rickroll/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "*"},
				},
			},
			"//depot/project1/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "192.168.41.2"},
				},
			},

			"//depot/project2/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "[2001:db8:195:1:2::1234]"},
				},
			},
			"//depot/local/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "192.168.41.*"},
				},
			},
			"//depot/test/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "/..."), IP: "[2001:db8:1:2:*]"},
				},
			},
			"//not-depot/not-main/": {Paths: []authz.PathWithIP{{Path: "/**", IP: "*"}}},
		},
	}

	if diff := cmp.Diff(want, perms); diff != "" {
		t.Fatalf("unexpected permissions (-want +got):\n%s", diff)
	}
}

func TestScanFullRepoPermissionsWithWildcardMatchingDepot(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdata/sample-protects-m.txt")
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

	db := dbmocks.NewMockDB()

	p := NewProvider(logger, db, gitserver.NewStrictMockClient(), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
	p.depots = []extsvc.RepoID{
		"//depot/main/base/",
	}
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
	}
	if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), fullRepoPermsScanner(logger, perms, p.depots), false); err != nil {
		t.Fatal(err)
	}

	want := &authz.ExternalUserPermissions{
		Exacts: []extsvc.RepoID{
			"//depot/main/base/",
		},
		SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs{
			"//depot/main/base/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "-/**"), IP: "*"},
					{Path: mustGlobPattern(t, "/**"), IP: "*"},
					{Path: mustGlobPattern(t, "-/**"), IP: "*"},
					{Path: mustGlobPattern(t, "-/**/base/build/deleteorgs.txt"), IP: "*"},
					{Path: mustGlobPattern(t, "-/build/deleteorgs.txt"), IP: "*"},
					{Path: mustGlobPattern(t, "-/**/base/build/**/asdf.txt"), IP: "*"},
					{Path: mustGlobPattern(t, "-/build/**/asdf.txt"), IP: "*"},
				},
			},
		},
	}

	if diff := cmp.Diff(want, perms); diff != "" {
		t.Fatal(diff)
	}
}

func TestFullScanMatchRules(t *testing.T) {
	for _, tc := range []struct {
		name                string
		depot               string
		protects            string
		protectsFile        string
		canReadAll          []string
		cannotReadAny       []string
		noRules             bool
		ignoreRulesWithHost bool
	}{
		// Confirm the rules as defined in
		// https://www.perforce.com/manuals/p4sag/Content/P4SAG/protections-implementation.html
		//
		// Modified slightly by removing the //depot prefix and only including rules
		// applicable to the actual user since that's what we'll get back from protect -u
		{
			name:  "Without an exclusionary mapping, the most permissive line rules",
			depot: "//depot/",
			protects: `
write       group       Dev2    *    //depot/dev/...
read        group       Dev1    *    //depot/dev/productA/...
write       group       Dev1    *    //depot/elm_proj/...
`,
			canReadAll:          []string{"dev/productA/readme.txt"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Rules that include a host are ignored",
			depot: "//depot/",
			protects: `
write       group       Dev2    *    //depot/dev/...
read        group       Dev1    *    //depot/dev/productA/...
write       group       Dev1    *    //depot/elm_proj/...
read		group		Dev1    192.168.10.1/24    -//depot/dev/productA/...
`,
			canReadAll:          []string{"dev/productA/readme.txt"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Exclusion overrides prior inclusion",
			depot: "//depot/",
			protects: `
write   group   Dev1   *   //depot/dev/...            ## Maria is a member of Dev1
list    group   Dev1   *   -//depot/dev/productA/...  ## exclusionary mapping overrides the line above
`,
			cannotReadAny:       []string{"dev/productA/readme.txt"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Exclusionary mapping and =, - before file path",
			depot: "//depot/",
			protects: `
list   group   Rome    *   -//depot/dev/prodA/...   ## exclusion of list implies no read, no write, etc.
read   group   Rome    *   //depot/dev/prodA/...   ## Rome can only read this one path
`,
			cannotReadAny: []string{"dev/prodB/things.txt"},
			// The include appears after the exclude so it should take preference
			canReadAll:          []string{"dev/prodA/things.txt"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Exclusionary mapping and =, - before the file path and = before the access level",
			depot: "//depot/",
			protects: `
read  group     Rome    *  //depot/dev/...
=read  group    Rome    *  -//depot/dev/prodA/...   ## Rome cannot read this one path
`,
			cannotReadAny:       []string{"dev/prodA/things.txt"},
			ignoreRulesWithHost: true,
		},
		// Extra test cases not from the above perforce page. These are obfuscated tests
		// generated from production use cases
		{
			name:                "File visibility on a group allowing repository level access",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-ed.txt",
			canReadAll:          []string{"depot/foo/bar", "depot/foo/bar/activities", "depot/foo/bar/activity-platform-api/BUILD", "depot/foo/bar/aa/build/README"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Restricted access tests with edm",
			depot: "//depot/foo/bar/",
			// NOTE that this file has many =write exclude rules which are ignored since
			// revoking write access with = does not remove read access.
			protectsFile: "testdata/sample-protects-edm.txt",
			// This file only includes exclude rules so our logic strips them out so that we
			// end up with zero rules.
			noRules:             true,
			ignoreRulesWithHost: true,
		},
		{
			name:         "Restricted access tests",
			depot:        "//depot/foo/bar/",
			protectsFile: "testdata/sample-protects-e.txt",
			// This file only includes exclude rules so our logic strips them out so that we
			// end up with zero rules.
			noRules:             true,
			ignoreRulesWithHost: true,
		},
		{
			name:                "Allow read access to a path using a rule containing a wildcard",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-edb.txt",
			canReadAll:          []string{"db/plpgsql/seed.psql"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Singular group allowing read access to a particular path",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-readonly.txt",
			canReadAll:          []string{"pom.xml"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Allow high, allow low",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-dcro.txt",
			canReadAll:          []string{"depot/foo/bar"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Allow high, deny low",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-everyone-revoke-read.txt",
			cannotReadAny:       []string{"depot/foo/bar"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Deny high, deny low",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-everyone-revoke-read.txt",
			cannotReadAny:       []string{"depot/foo/bar"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Allow path, allow path",
			depot:               "//depot/foo/bar/",
			protectsFile:        "testdata/sample-protects-ro-aw.txt",
			canReadAll:          []string{"depot/foo/bar"},
			ignoreRulesWithHost: true,
		},
		{
			name:                "Allow read access to a path using a rule containing a wildcard",
			depot:               "//depot/236/freeze/cc/",
			protectsFile:        "testdata/sample-protects-edb.txt",
			canReadAll:          []string{"db/plpgsql/seed.psql"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Leading slash edge cases",
			depot: "//depot/",
			protects: `
read   group   Rome    *   //depot/.../something.java   ## Can read all files named 'something.java'
read   group   Rome    *   -//depot/dev/prodA/...   ## Except files in this directory
`,
			cannotReadAny:       []string{"dev/prodA/something.java", "dev/prodA/another_dir/something.java", "/dev/prodA/something.java", "/dev/prodA/another_dir/something.java"},
			canReadAll:          []string{"something.java", "/something.java", "dev/prodB/something.java", "/dev/prodC/something.java"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Deny all, grant some",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   -//depot/main/...
read    group   Dev1    *   -//depot/main/.../*.java
read    group   Dev1    *   //depot/main/.../dev/foo.java
`,
			canReadAll:          []string{"dev/foo.java"},
			cannotReadAny:       []string{"dev/bar.java"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Grant all, deny some",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/...
read    group   Dev1    *   //depot/main/.../*.java
read    group   Dev1    *   -//depot/main/.../dev/foo.java
`,
			canReadAll:          []string{"dev/bar.java"},
			cannotReadAny:       []string{"dev/foo.java"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Tricky minus names",
			depot: "//-depot/-main/",
			protects: `
read    group   Dev1    *   //-depot/-main/...
read    group   Dev1    *   //-depot/-main/.../*.java
read    group   Dev1    *   -//-depot/-main/.../dev/foo.java
`,
			canReadAll:          []string{"dev/bar.java", "/-minus/dev/bar.java"},
			cannotReadAny:       []string{"dev/foo.java"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Root matching",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/.../*.java
`,
			canReadAll:          []string{"dev/bar.java", "foo.java", "/foo.java"},
			cannotReadAny:       []string{"dev/foo.go"},
			ignoreRulesWithHost: true,
		},
		{
			name:  "Root matching, multiple levels",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/.../.../*.java
`,
			canReadAll:          []string{"/foo/dev/bar.java", "foo.java", "/foo.java"},
			cannotReadAny:       []string{"dev/foo.go"},
			ignoreRulesWithHost: true,
		},
		{
			// In this case, Perforce still shows the parent directory
			name:  "Files in side directory hidden",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/...
read    group   Dev1    *   -//depot/main/dir/*.java
`,
			canReadAll:          []string{"dir/"},
			cannotReadAny:       []string{"dir/foo.java", "dir/bar.java"},
			ignoreRulesWithHost: true,
		},
		{
			// Directory excluded, but file inside included: Directory visible
			name:  "Directory excluded",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/...
read    group   Dev1    *   -//depot/main/dir/...
read    group   Dev1    *   //depot/main/dir/file.java
`,
			canReadAll:          []string{"dir/file.java", "dir/"},
			ignoreRulesWithHost: true,
		},
		{
			// Should still be able to browse directories
			name:  "Rules start with wildcard",
			depot: "//depot/main/",
			protects: `
read    group   Dev1    *   //depot/main/.../*.go
`,
			canReadAll:          []string{"dir/file.go", "dir/"},
			ignoreRulesWithHost: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			logger := logtest.Scoped(t)
			if !strings.HasPrefix(tc.depot, "/") {
				t.Fatal("depot must end in '/'")
			}
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					ExperimentalFeatures: &schema.ExperimentalFeatures{
						SubRepoPermissions: &schema.SubRepoPermissions{
							Enabled: true,
						},
					},
				},
			})
			t.Cleanup(func() { conf.Mock(nil) })

			ctx := context.Background()
			ctx = actor.WithActor(ctx, &actor.Actor{UID: 1})

			var rc io.ReadCloser
			var err error
			if len(tc.protects) > 0 {
				rc = io.NopCloser(strings.NewReader(tc.protects))
			}
			if len(tc.protectsFile) > 0 {
				rc, err = os.Open(tc.protectsFile)
				if err != nil {
					t.Fatal(err)
				}
			}
			t.Cleanup(func() {
				if err := rc.Close(); err != nil {
					t.Fatal(err)
				}
			})

			db := dbmocks.NewMockDB()

			p := NewProvider(logger, db, gitserver.NewStrictMockClient(), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
			p.depots = []extsvc.RepoID{
				extsvc.RepoID(tc.depot),
			}
			perms := &authz.ExternalUserPermissions{
				SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
			}
			if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), fullRepoPermsScanner(logger, perms, p.depots), tc.ignoreRulesWithHost); err != nil {
				t.Fatal(err)
			}
			rules, ok := perms.SubRepoPermissions[extsvc.RepoID(tc.depot)]
			if !ok && tc.noRules {
				return
			}
			if !ok && !tc.noRules {
				t.Fatal("no rules found")
			} else if ok && tc.noRules {
				t.Fatal("expected no rules")
			}
			checker := srp.NewSimpleChecker(api.RepoName(tc.depot), rules.Paths)

			if len(tc.canReadAll) > 0 {
				ok, err = authz.CanReadAllPaths(ctx, checker, api.RepoName(tc.depot), tc.canReadAll)
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					t.Fatal("should be able to read path")
				}
			}
			if len(tc.cannotReadAny) > 0 {
				for _, path := range tc.cannotReadAny {
					ok, err = authz.CanReadAllPaths(ctx, checker, api.RepoName(tc.depot), []string{path})
					if err != nil {
						t.Fatal(err)
					}
					if ok {
						t.Errorf("should not be able to read %q, but can", path)
					}
				}
			}
		})
	}
}

func TestFullScanWildcardDepotMatching(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdata/sample-protects-x.txt")
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

	db := dbmocks.NewMockDB()

	p := NewProvider(logger, db, gitserver.NewStrictMockClient(), "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
	p.depots = []extsvc.RepoID{
		"//depot/654/deploy/base/",
	}
	perms := &authz.ExternalUserPermissions{
		SubRepoPermissions: make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs),
	}
	if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), fullRepoPermsScanner(logger, perms, p.depots), false); err != nil {
		t.Fatal(err)
	}

	want := &authz.ExternalUserPermissions{
		Exacts: []extsvc.RepoID{
			"//depot/654/deploy/base/",
		},
		SubRepoPermissions: map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs{
			"//depot/654/deploy/base/": {
				Paths: []authz.PathWithIP{
					{Path: mustGlobPattern(t, "-/**"), IP: "*"},
					{Path: mustGlobPattern(t, "-/**/base/build/deleteorgs.txt"), IP: "*"},
					{Path: mustGlobPattern(t, "-/build/deleteorgs.txt"), IP: "*"},
					{Path: mustGlobPattern(t, "-/asdf/plsql/base/cCustomSchema*.sql"), IP: "*"},
					{Path: mustGlobPattern(t, "/db/upgrade-scripts/**"), IP: "*"},
					{Path: mustGlobPattern(t, "/db/my_db/upgrade-scripts/**"), IP: "*"},
					{Path: mustGlobPattern(t, "/asdf/config/my_schema.xml"), IP: "*"},
					{Path: mustGlobPattern(t, "/db/plpgsql/**"), IP: "*"},
				},
			},
		},
	}

	if diff := cmp.Diff(want, perms); diff != "" {
		t.Fatal(diff)
	}
}

func TestCheckWildcardDepotMatch(t *testing.T) {
	testDepot := extsvc.RepoID("//depot/main/base/")
	testCases := []struct {
		name               string
		pattern            string
		original           string
		expectedNewRules   []string
		expectedFoundMatch bool
		depot              extsvc.RepoID
	}{
		{
			name:             "depot match ends with double wildcard",
			pattern:          "//depot/**/README.md",
			original:         "//depot/.../README.md",
			expectedNewRules: []string{"**/README.md"},
			depot:            "//depot/test/",
		},
		{
			name:             "single wildcard",
			pattern:          "//depot/*/dir.yaml",
			original:         "//depot/*/dir.yaml",
			expectedNewRules: []string{"dir.yaml"},
			depot:            "//depot/test/",
		},
		{
			name:             "single wildcard in depot match",
			pattern:          "//depot/**/base/build/deleteorgs.txt",
			original:         "//depot/.../base/build/deleteorgs.txt",
			expectedNewRules: []string{"**/base/build/deleteorgs.txt", "build/deleteorgs.txt"},
			depot:            testDepot,
		},
		{
			name:             "ends with wildcard",
			pattern:          "//depot/**",
			original:         "//depot/...",
			expectedNewRules: []string{"**"},
			depot:            testDepot,
		},
		{
			name:             "two wildcards",
			pattern:          "//depot/**/tests/**/my_test",
			original:         "//depot/.../test/.../my_test",
			expectedNewRules: []string{"**/tests/**/my_test"},
			depot:            testDepot,
		},
		{
			name:             "no match no effect",
			pattern:          "//foo/**/base/build/asdf.txt",
			original:         "//foo/.../base/build/asdf.txt",
			expectedNewRules: []string{"//foo/**/base/build/asdf.txt"},
			depot:            testDepot,
		},
		{
			name:             "original rule is fine, no changes needed",
			pattern:          "//**/.secrets.env",
			original:         "//.../.secrets.env",
			expectedNewRules: []string{"//**/.secrets.env"},
			depot:            testDepot,
		},
		{
			name:             "single wildcard match",
			pattern:          "//depot/6*/*/base/schema/submodules**",
			original:         "//depot/6*/*/base/schema/submodules**",
			expectedNewRules: []string{"schema/submodules**"},
			depot:            "//depot/654/deploy/base/",
		},
		{
			name:             "single wildcard match no double wildcard",
			pattern:          "//depot/6*/*/base/asdf/java/resources/foo.xml",
			original:         "//depot/6*/*/base/asdf/java/resources/foo.xml",
			expectedNewRules: []string{"asdf/java/resources/foo.xml"},
			depot:            "//depot/654/deploy/base/",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pattern := tc.pattern
			glob := mustGlob(t, pattern)
			rule := globMatch{
				glob,
				pattern,
				tc.original,
			}
			newRules := convertRulesForWildcardDepotMatch(rule, tc.depot, map[string]globMatch{})
			if diff := cmp.Diff(newRules, tc.expectedNewRules); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestScanAllUsers(t *testing.T) {
	logger := logtest.Scoped(t)
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
	gc := gitserver.NewStrictMockClient()
	gc.PerforceGroupMembersFunc.SetDefaultReturn(nil, nil)

	db := dbmocks.NewMockDB()

	p := NewProvider(logger, db, gc, "", "ssl:111.222.333.444:1666", "admin", "password", []extsvc.RepoID{}, false)
	p.cachedGroupMembers = map[string][]string{
		"dev": {"user1", "user2"},
	}
	p.cachedAllUserEmails = map[string]string{
		"user1": "user1@example.com",
		"user2": "user2@example.com",
	}

	users := make(map[string]struct{})
	if err := scanProtects(logger, testParseP4ProtectsRaw(t, rc), allUsersScanner(ctx, p, users), false); err != nil {
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

func testParseP4ProtectsRaw(t *testing.T, rc io.Reader) []*p4types.Protect {
	protects, err := parseP4ProtectsRaw(rc)
	require.NoError(t, err)
	return protects
}
