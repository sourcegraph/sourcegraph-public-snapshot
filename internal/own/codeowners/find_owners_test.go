package codeowners_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
)

type testCase struct {
	pattern string
	paths   []string
}

func TestFileOwnersMatch(t *testing.T) {
	cases := []testCase{
		{
			pattern: "filename",
			paths: []string{
				"/filename",
				"/prefix/filename",
			},
		},
		{
			pattern: "*.md",
			paths: []string{
				"/README.md",
				"/README.md.md",
				"/nested/index.md",
				"/weird/but/matching/.md",
			},
		},
		{
			// Regex components are interpreted literally.
			pattern: "[^a-z].md",
			paths: []string{
				"/[^a-z].md",
				"/nested/[^a-z].md",
			},
		},
		{
			pattern: "foo*bar*baz",
			paths: []string{
				"/foobarbaz",
				"/foo-bar-baz",
				"/foobarbazfoobarbazfoobarbaz",
			},
		},
		{
			pattern: "directory/path/",
			paths: []string{
				"/directory/path/file",
				"/directory/path/deeply/nested/file",
				"/prefix/directory/path/file",
				"/prefix/directory/path/deeply/nested/file",
			},
		},
		{
			pattern: "directory/path/**",
			paths: []string{
				"/directory/path/file",
				"/directory/path/deeply/nested/file",
				"/prefix/directory/path/file",
				"/prefix/directory/path/deeply/nested/file",
			},
		},
		{
			pattern: "directory/*",
			paths: []string{
				"/directory/file",
				"/prefix/directory/another_file",
			},
		},
		{
			pattern: "/toplevelfile",
			paths: []string{
				"/toplevelfile",
			},
		},
		{
			pattern: "/main/src/**/README.md",
			paths: []string{
				"/main/src/README.md",
				"/main/src/foo/bar/README.md",
			},
		},
	}
	for _, c := range cases {
		for _, path := range c.paths {
			pattern := c.pattern
			owner := []*codeownerspb.Owner{
				{Handle: "foo"},
			}
			file := codeowners.NewRuleset(&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{Pattern: pattern, Owner: owner},
				},
			})
			got := file.FindOwners(path)
			if !reflect.DeepEqual(got, owner) {
				t.Errorf("want %q to match %q", pattern, path)
			}
		}
	}
}

func TestFileOwnersNoMatch(t *testing.T) {
	cases := []testCase{
		{
			pattern: "filename",
			paths: []string{
				"/prefix_filename_suffix",
				"/src/prefix_filename",
				"/finemale/nested",
			},
		},
		{
			pattern: "*.md",
			paths: []string{
				"/README.mdf",
				"/not/matching/without/the/dot/md",
			},
		},
		{
			// Regex components are interpreted literally.
			pattern: "[^a-z].md",
			paths: []string{
				"/-.md",
				"/nested/%.md",
			},
		},
		{
			pattern: "foo*bar*baz",
			paths: []string{
				"/foo-ba-baz",
				"/foobarbaz.md",
			},
		},
		{
			pattern: "directory/leaf/",
			paths: []string{
				// These do not match as the right-most directory name `leaf`
				// is just a prefix to the corresponding directory on the given path.
				"/directory/leaf_and_more/file",
				"/prefix/directory/leaf_and_more/file",
				// These do not match as the pattern matches anything within
				// the sub-directory tree, but not the directory itself.
				"/directory/leaf",
				"/prefix/directory/leaf",
			},
		},
		{
			pattern: "directory/leaf/**",
			paths: []string{
				// These do not match as the right-most directory name `leaf`
				// is just a prefix to the corresponding directory on the given path.
				"/directory/leaf_and_more/file",
				"/prefix/directory/leaf_and_more/file",
				// These do not match as the pattern matches anything within
				// the sub-directory tree, but not the directory itself.
				"/directory/leaf",
				"/prefix/directory/leaf",
			},
		},
		{
			pattern: "directory/*",
			paths: []string{
				"/directory/nested/file",
				"/directory/deeply/nested/file",
			},
		},
		{
			pattern: "/toplevelfile",
			paths: []string{
				"/toplevelfile/nested",
				"/notreally/toplevelfile",
			},
		},
		{
			pattern: "/main/src/**/README.md",
			paths: []string{
				"/main/src/README.mdf",
				"/main/src/README.md/looks-like-a-file-but-was-dir",
				"/main/src/foo/bar/README.mdf",
				"/nested/main/src/README.md",
				"/nested/main/src/foo/bar/README.md",
			},
		},
	}
	for _, c := range cases {
		for _, path := range c.paths {
			pattern := c.pattern
			owner := []*codeownerspb.Owner{
				{Handle: "foo"},
			}
			file := codeowners.NewRuleset(&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{Pattern: pattern, Owner: owner},
				},
			})
			got := file.FindOwners(path)
			if got != nil {
				t.Errorf("want %q not to match %q", pattern, path)
			}
		}
	}
}

func TestFileOwnersOrder(t *testing.T) {
	wantOwner := []*codeownerspb.Owner{{Handle: "some-path-owner"}}
	file := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{
				Pattern: "/top-level-directory/",
				Owner:   []*codeownerspb.Owner{{Handle: "top-level-owner"}},
			},
			// The owner of the last matching pattern is being picked
			{
				Pattern: "some/path/*",
				Owner:   wantOwner,
			},
			{
				Pattern: "does/not/match",
				Owner:   []*codeownerspb.Owner{{Handle: "not-matching-owner"}},
			},
		},
	})
	got := file.FindOwners("/top-level-directory/some/path/main.go")
	assert.Equal(t, wantOwner, got)
}

func BenchmarkOwnersMatchLiteral(b *testing.B) {
	pattern := "/main/src/foo/bar/README.md"
	paths := []string{
		"/main/src/foo/bar/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMatchRelativeGlob(b *testing.B) {
	pattern := "**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMatchAbsoluteGlob(b *testing.B) {
	pattern := "/main/**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMismatchLiteral(b *testing.B) {
	pattern := "/main/src/foo/bar/README.md"
	paths := []string{
		"/main/src/foo/bar/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMismatchRelativeGlob(b *testing.B) {
	pattern := "**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMismatchAbsoluteGlob(b *testing.B) {
	pattern := "/main/**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMatchMultiHole(b *testing.B) {
	pattern := "/main/**/foo/**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMismatchMultiHole(b *testing.B) {
	pattern := "/main/**/foo/**/*.md"
	paths := []string{
		"/main/src/foo/bar/README.txt",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	rs := codeowners.NewRuleset(&codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	})
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func BenchmarkOwnersMatchLiteralLargeRuleset(b *testing.B) {
	pattern := "/main/src/foo/bar/README.md"
	paths := []string{
		"/main/src/foo/bar/README.md",
	}
	owner := []*codeownerspb.Owner{
		{Handle: "foo"},
	}
	f := &codeownerspb.File{
		Rule: []*codeownerspb.Rule{
			{Pattern: pattern, Owner: owner},
		},
	}
	for i := 0; i < 10000; i++ {
		f.Rule = append(f.Rule, &codeownerspb.Rule{Pattern: fmt.Sprintf("%s-%d", pattern, i), Owner: owner})
	}
	rs := codeowners.NewRuleset(f)
	// Warm cache.
	for _, path := range paths {
		rs.FindOwners(path)
	}

	for i := 0; i < b.N; i++ {
		rs.FindOwners(pattern)
	}
}

func init() {
	c := `/.github/workflows/codenotify.yml @unknwon
/.github/workflows/licenses-check.yml          @bobheadxi
/.github/workflows/licenses-update.yml         @bobheadxi
/.github/workflows/renovate-downstream.yml     @bobheadxi
/.github/workflows/renovate-downstream.json    @bobheadxi

/client/branded/src/search-ui/components/** @limitedmage @fkling

/client/jetbrains/** @vdavid @philipp-spiess

/client/shared/src/search/**/* @fkling

/client/shared/src/search/query/**/* @fkling

/client/web/src/enterprise/batches/**/* @eseliger @courier-new @BolajiOlajide

/client/web/src/enterprise/code-monitoring/**/* @limitedmage

/client/web/src/enterprise/codeintel/**/* @efritz

/client/web/src/enterprise/executors/**/* @efritz @eseliger

/client/web/src/integration/batches* @eseliger @courier-new @BolajiOlajide
/client/web/src/integration/search* @limitedmage
/client/web/src/integration/code-monitoring* @limitedmage

/client/web/src/search/**/* @limitedmage @fkling

/cmd/blobstore/**/* @slimsag

/cmd/frontend/graphqlbackend/observability.go @bobheadxi
/cmd/frontend/graphqlbackend/site_monitoring.go @bobheadxi
/cmd/frontend/graphqlbackend/*search*.go @keegancsmith
/cmd/frontend/graphqlbackend/*zoekt*.go @keegancsmith
/cmd/frontend/graphqlbackend/batches.go @eseliger @courier-new
/cmd/frontend/graphqlbackend/insights.go @sourcegraph/code-insights-backend
/cmd/frontend/graphqlbackend/codeintel.go @efritz
/cmd/frontend/graphqlbackend/oobmigrations.go @efritz

/cmd/frontend/internal/search/** @keegancsmith
/cmd/frontend/internal/search/**/* @camdencheek

/cmd/gitserver/server/* @indradhanush @sashaostrikov

/cmd/repo-updater/* @indradhanush @sashaostrikov

/cmd/searcher/**/* @keegancsmith

/cmd/server/internal/goreman/** @keegancsmith

/cmd/server/internal/goremancmd/** @keegancsmith

/cmd/symbols/** @keegancsmith

/cmd/worker/**/* @efritz

/dev/authtest/**/* @unknwon

/dev/codeintel-qa/**/* @efritz

/dev/depgraph/**/* @efritz

/dev/gqltest/**/* @unknwon

/doc/**/admin/** @sourcegraph/delivery
/doc/dev/background-information/adding_ping_data.md @ebrodymoore @dadlerj

/doc/batch_changes/**/* @eseliger @courier-new

/doc/code_navigation/**/* @efritz

/doc/code_search/**/* @rvantonder

/doc/dev/adr/**/* @unknwon

/doc/dev/background-information/codeintel/**/* @efritz

/docker-images/cadvisor/**/* @bobheadxi

/docker-images/grafana/**/* @bobheadxi

/docker-images/postgres-12-alpine/**/* @sourcegraph/delivery

/docker-images/prometheus/**/* @bobheadxi

/enterprise/cmd/executor/**/* @efritz

/enterprise/cmd/frontend/internal/auth/**/* @unknwon

/enterprise/cmd/frontend/internal/authz/**/* @unknwon

/enterprise/cmd/frontend/internal/codeintel/**/* @efritz

/enterprise/cmd/frontend/internal/executorqueue/**/* @efritz @eseliger

/enterprise/cmd/frontend/internal/licensing/**/* @unknwon

/enterprise/cmd/migrator/**/* @efritz

/enterprise/cmd/precise-code-intel-worker/**/* @efritz

/enterprise/cmd/repo-updater/**/* @indradhanush

/enterprise/cmd/repo-updater/internal/authz/**/* @unknwon

/enterprise/cmd/worker/**/* @efritz

/enterprise/cmd/worker/internal/batches/**/* @eseliger

/enterprise/cmd/worker/internal/executorqueue/**/* @efritz @eseliger

/enterprise/cmd/worker/internal/executors/**/* @efritz

/enterprise/dev/ci/**/* @bobheadxi

/enterprise/internal/authz/**/* @unknwon

/enterprise/internal/batches/**/* @eseliger

/enterprise/internal/cloud/**/* @unknwon @michaellzc

/enterprise/internal/codeintel/**/* @efritz @Strum355

/enterprise/internal/database/external_services* @unknwon
/enterprise/internal/database/perms_store* @unknwon

/enterprise/internal/executor/**/* @efritz

/enterprise/internal/insights/**/* @sourcegraph/code-insights-backend

/enterprise/internal/license/**/* @unknwon

/enterprise/internal/licensing/**/* @unknwon

/internal/authz/**/* @unknwon

/internal/codeintel/**/* @efritz

/internal/codeintel/dependencies/**/* @mrnugget

/internal/database/external* @eseliger
/internal/database/namespaces* @eseliger
/internal/database/repos* @eseliger
/internal/database/permissions* @BolajiOlajide
/internal/database/user_roles* @BolajiOlajide
/internal/database/roles* @BolajiOlajide
/internal/database/role_permissions* @BolajiOlajide

/internal/database/basestore/**/* @efritz

/internal/database/batch/**/* @efritz

/internal/database/connections/**/* @efritz

/internal/database/dbconn/**/* @efritz

/internal/database/dbtest/**/* @efritz

/internal/database/dbutil/**/* @efritz

/internal/database/locker/**/* @efritz

/internal/database/migration/**/* @efritz

/internal/database/postgresdsn/**/* @efritz

/internal/debugserver/** @keegancsmith

/internal/diskcache/** @keegancsmith

/internal/endpoint/** @keegancsmith

/internal/env/baseconfig.go @efritz

/internal/extsvc/**/* @eseliger

/internal/extsvc/auth/**/* @unknwon

/internal/gitserver/* @indradhanush @sashaostrikov

/internal/goroutine/**/* @efritz

/internal/gqltestutil/**/* @unknwon

/internal/honey/** @keegancsmith

/internal/httpcli/** @keegancsmith

/internal/lazyregexp/** @keegancsmith

/internal/luasandbox/**/* @efritz

/internal/mutablelimiter/** @keegancsmith

/internal/observation/**/* @sourcegraph/dev-experience

/internal/oobmigration/**/* @efritz

/internal/rcache/** @keegancsmith

/internal/redispool/** @keegancsmith

/internal/repos/* @indradhanush @sashaostrikov

/internal/search/**/* @keegancsmith @camdencheek

/internal/src-cli/**/* @eseliger @BolajiOlajide @courier-new

/internal/symbols/** @keegancsmith

/internal/sysreq/** @keegancsmith

/internal/trace/** @keegancsmith
/internal/trace/**/* @sourcegraph/dev-experience

/internal/tracer/** @keegancsmith
/internal/tracer/**/* @sourcegraph/dev-experience

/internal/usagestats/batches.go @eseliger
/internal/usagestats/batches_test.go @eseliger
/internal/usagestats/*codeintel*.go @efritz

/internal/vcs/** @keegancsmith

/internal/workerutil/**/* @efritz

/lib/codeintel/**/* @efritz

/lib/errors/**/* @sourcegraph/dev-experience

/lib/servicecatalog/**/* @sourcegraph/cloud @sourcegraph/security

/monitoring/**/* @bobheadxi @slimsag @sourcegraph/delivery
/monitoring/frontend.go @efritz
/monitoring/precise_code_intel_* @efritz @sourcegraph/code-intelligence`

	file, _ = codeowners.Parse(strings.NewReader(c))
	forist, _ = codeowners.ParseTrie(strings.NewReader(c))
}

var file *codeowners.Ruleset
var forist codeowners.PatternForist

var cases = []string{
	"/main/src/foo/bar/README.txt",
}

func BenchmarkRuleset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pattern := cases[i%len(cases)]
		file.FindOwners(pattern)
	}
}

func BenchmarkTrie(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pattern := cases[i%len(cases)]
		forist.Find(strings.Split(pattern, "/"))
	}
}
