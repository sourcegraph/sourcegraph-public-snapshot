package lockfiles

import (
	"context"
	"io"
	"os"
	"sort"
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
)

func TestListDependencies(t *testing.T) {
	ctx := context.Background()

	t.Run("empty ls-files return", func(t *testing.T) {
		gitSvc := NewMockGitService()
		gitSvc.LsFilesFunc.SetDefaultReturn([]string{}, nil)

		results, err := TestService(gitSvc).ListDependencies(ctx, "foo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(results) != 0 {
			t.Fatalf("expected no dependencies")
		}
	})

	t.Run("npm", func(t *testing.T) {
		gitSvc := NewMockGitService()
		gitSvc.LsFilesFunc.SetDefaultReturn([]string{
			"client/package-lock.json",
			"package-lock.json",
			"yarn.lock",
		}, nil)

		yarnLock, err := os.ReadFile("testdata/parse/yarn.lock/yarn_normal.lock")
		if err != nil {
			t.Fatal(err)
		}

		gitSvc.ArchiveFunc.SetDefaultHook(zipArchive(t, map[string]string{
			"client/package-lock.json": `{"dependencies": { "@octokit/request": {"version": "5.6.2"} }}`,
			"package-lock.json":        `{"dependencies": { "promise": {"version": "8.0.3"} }}`,
			"yarn.lock":                string(yarnLock),
		}))

		results, err := TestService(gitSvc).ListDependencies(ctx, "foo", "HEAD")
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(results), 3; have != want {
			t.Fatalf("wrong number of results. want=%d, have=%d", want, have)
		}

		have := make([]serializableResult, 0, len(results))
		for _, res := range results {
			have = append(have, serializeResult(res))
		}

		sort.Slice(have, func(i, j int) bool { return have[i].Lockfile < have[j].Lockfile })

		g := goldie.New(t, goldie.WithFixtureDir("testdata/svc"))
		g.AssertJson(t, t.Name(), have)
	})

	t.Run("go", func(t *testing.T) {
		gitSvc := NewMockGitService()
		gitSvc.LsFilesFunc.SetDefaultReturn([]string{
			"subpkg/go.mod",
			"go.mod",
		}, nil)

		gitSvc.ArchiveFunc.SetDefaultHook(zipArchive(t, map[string]string{
			"subpkg/go.mod": `
require modernc.org/cc v1.0.0
require modernc.org/golex v1.0.0
require github.com/google/uuid v1.0.0
`,
			// google/uuid is in here twice, we want to make sure we de-duplicate, but only per lockfile
			"go.mod": `
require github.com/google/uuid v1.0.0
require github.com/google/uuid v1.0.0
require github.com/pborman/uuid v1.2.1
`,
		}))

		results, err := TestService(gitSvc).ListDependencies(ctx, "foo", "HEAD")
		if err != nil {
			t.Fatal(err)
		}

		if have, want := len(results), 2; have != want {
			t.Fatalf("wrong number of results. want=%d, have=%d", want, have)
		}

		have := make([]serializableResult, 0, len(results))
		for _, res := range results {
			have = append(have, serializeResult(res))
		}

		sort.Slice(have, func(i, j int) bool { return have[i].Lockfile < have[j].Lockfile })

		g := goldie.New(t, goldie.WithFixtureDir("testdata/svc"))
		g.AssertJson(t, t.Name(), have)
	})
}

func zipArchive(t testing.TB, files map[string]string) func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return func(ctx context.Context, name api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error) {
		return unpacktest.CreateZipArchive(t, files), nil
	}
}

type serializableResult struct {
	Deps     []string
	Lockfile string
	Graph    string
}

func serializeResult(res Result) serializableResult {
	serializable := serializableResult{Lockfile: res.Lockfile}

	if res.Graph != nil {
		serializable.Graph = res.Graph.String()
	} else {
		serializable.Graph = "NO-GRAPH"
	}

	for _, dep := range res.Deps {
		serializable.Deps = append(serializable.Deps, dep.VersionedPackageSyntax())
	}

	sort.Strings(serializable.Deps)
	return serializable
}
