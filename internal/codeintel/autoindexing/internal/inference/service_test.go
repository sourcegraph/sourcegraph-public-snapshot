package inference

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/grafana/regexp"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
)

func testService(t *testing.T, repositoryContents map[string]string) *Service {
	// Real deal
	sandboxService := luasandbox.GetService()

	// Fake deal
	gitService := NewMockGitService()
	gitService.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (paths []string, _ error) {
		for path := range repositoryContents {
			if pattern.MatchString(path) {
				paths = append(paths, path)
			}
		}

		return
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

	return newService(sandboxService, gitService, rate.NewLimiter(rate.Limit(100), 1), 100, 1024*1024, &observation.TestContext)
}
