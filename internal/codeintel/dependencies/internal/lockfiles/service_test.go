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

		got, err := TestService(gitSvc).ListDependencies(ctx, "foo", "")
		if err != nil {
			t.Fatal(err)
		}

		if len(got) != 0 {
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
			// promise@8.0.3 is also in yarn.lock. We test that it gets de-duplicated.
			"package-lock.json": `{"dependencies": { "promise": {"version": "8.0.3"} }}`,
			"yarn.lock":         string(yarnLock),
		}))

		deps, err := TestService(gitSvc).ListDependencies(ctx, "foo", "HEAD")
		if err != nil {
			t.Fatal(err)
		}

		have := make([]string, 0, len(deps))
		for _, dep := range deps {
			have = append(have, dep.PackageManagerSyntax())
		}

		sort.Strings(have)

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
			// github.com/google/uuid v1.0.0 is also in go.mod. We test that it gets de-duplicated.
			"subpkg/go.mod": `
require modernc.org/cc v1.0.0
require modernc.org/golex v1.0.0
require github.com/google/uuid v1.0.0
`,
			"go.mod": `
require github.com/google/uuid v1.0.0
require github.com/pborman/uuid v1.2.1
`,
		}))

		deps, err := TestService(gitSvc).ListDependencies(ctx, "foo", "HEAD")
		if err != nil {
			t.Fatal(err)
		}

		have := make([]string, 0, len(deps))
		for _, dep := range deps {
			have = append(have, dep.PackageManagerSyntax())
		}

		sort.Strings(have)

		g := goldie.New(t, goldie.WithFixtureDir("testdata/svc"))
		g.AssertJson(t, t.Name(), have)
	})
}

func zipArchive(t testing.TB, files map[string]string) func(context.Context, api.RepoName, gitserver.ArchiveOptions) (io.ReadCloser, error) {
	return func(ctx context.Context, name api.RepoName, options gitserver.ArchiveOptions) (io.ReadCloser, error) {
		return unpacktest.CreateZipArchive(t, files), nil
	}
}
