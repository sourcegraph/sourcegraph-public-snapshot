package inference

import (
	"context"
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/luasandbox"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/unpack/unpacktest"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

func TestRecognizersEmpty(t *testing.T) {
	testRecognizers(t,
		recognizerTestCase{
			description:        "empty",
			repositoryContents: nil,
			expected:           nil,
		},
	)
}

type recognizerTestCase struct {
	description        string
	repositoryContents map[string]string
	expected           []config.IndexJob
}

func testRecognizers(t *testing.T, testCases ...recognizerTestCase) {
	for _, testCase := range testCases {
		testRecognizer(t, testCase)
	}
}

func testRecognizer(t *testing.T, testCase recognizerTestCase) {
	t.Run(testCase.description, func(t *testing.T) {
		// Real deal
		sandboxService := luasandbox.GetService()

		// Fake deal
		gitService := NewMockGitService()
		gitService.ListFilesFunc.SetDefaultHook(func(ctx context.Context, repo api.RepoName, commit string, pattern *regexp.Regexp) (paths []string, _ error) {
			for path := range testCase.repositoryContents {
				if pattern.MatchString(path) {
					paths = append(paths, path)
				}
			}

			return
		})
		gitService.ArchiveFunc.SetDefaultHook(func(ctx context.Context, repoName api.RepoName, opts gitserver.ArchiveOptions) (io.ReadCloser, error) {
			files := map[string]io.Reader{}
			for _, spec := range opts.Pathspecs {
				if contents, ok := testCase.repositoryContents[strings.TrimPrefix(string(spec), ":(literal)")]; ok {
					files[string(spec)] = strings.NewReader(contents)
				}
			}

			return unpacktest.CreateZipArchive(t, files)
		})

		jobs, err := newService(sandboxService, gitService, &observation.TestContext).InferIndexJobs(
			context.Background(),
			api.RepoName("github.com/test/test"),
			"HEAD",
			"", // TODO
		)
		if err != nil {
			t.Fatalf("unexpected error inferring jobs: %s", err)
		}
		if diff := cmp.Diff(sortIndexJobs(testCase.expected), sortIndexJobs(jobs)); diff != "" {
			t.Errorf("unexpected index jobs (-want +got):\n%s", diff)
		}
	})
}

func sortIndexJobs(s []config.IndexJob) []config.IndexJob {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
