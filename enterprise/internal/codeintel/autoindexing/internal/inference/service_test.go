package inference

import (
	"context"
	"io"
	"sort"
	"strings"
	"testing"

	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/paths"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
)

func testService(t *testing.T, repositoryContents map[string]string) *Service {
	repositoryPaths := make([]string, 0, len(repositoryContents))
	for path := range repositoryContents {
		repositoryPaths = append(repositoryPaths, path)
	}
	sort.Strings(repositoryPaths)

	// Real deal
	sandboxService := luasandbox.NewService()

	// Fake deal
	gitService := NewMockGitService()
	gitService.LsFilesFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit string, pathspecs ...gitdomain.Pathspec) ([]string, error) {
		var patterns []*paths.GlobPattern
		for _, spec := range pathspecs {
			pattern, err := paths.Compile(string(spec))
			if err != nil {
				return nil, err
			}

			patterns = append(patterns, pattern)
		}

		return filterPaths(repositoryPaths, patterns, nil), nil
	})
	gitService.ArchiveFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
		files := map[string]string{}
		for _, spec := range opts.Pathspecs {
			if contents, ok := repositoryContents[strings.TrimPrefix(string(spec), ":(literal)")]; ok {
				files[string(spec)] = contents
			}
		}

		return unpacktest.CreateTarArchive(t, files), nil
	})

	return newService(&observation.TestContext, sandboxService, gitService, ratelimit.NewInstrumentedLimiter("TestInference", rate.NewLimiter(rate.Limit(100), 1)), 100, 1024*1024)
}
