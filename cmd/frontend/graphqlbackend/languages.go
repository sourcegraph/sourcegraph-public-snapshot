package graphqlbackend

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (r *schemaResolver) GetLanguageForFile(ctx context.Context, args *struct {
	Path      string
	Content   string
	RepoId    int
	RepoName  string
	CommitSHA string
}) (string, error) {
	if args.Path == "" {
		return "", errors.New("missing file path")
	}

	metricsCache := r.db.FileMetrics()

	metrics := metricsCache.GetFileMetrics(context.TODO(), api.RepoID(args.RepoId), api.CommitID(args.CommitSHA), args.Path)

	if metrics != nil {
		return metrics.FirstLanguage(), nil
	}

	metrics = &fileutil.FileMetrics{}
	// try to determine language based on file name or extension first
	metrics.LanguagesByFileNameAndExtension(args.Path)
	if len(metrics.Languages) == 1 {
		return metrics.FirstLanguage(), nil
	}

	// language from file name/extension is ambiguous
	if len(args.Content) > 0 {
		// content was supplied, so use it
		metrics.LanguagesByFileContent(args.Path, []byte(args.Content))
		if len(metrics.Languages) == 1 {
			return metrics.FirstLanguage(), nil
		}
	}

	err := metrics.CalculateFileMetrics(
		context.TODO(),
		args.Path,
		func(ctx context.Context, path string) (io.ReadCloser, error) {
			return gitserver.NewClient().NewFileReader(ctx, authz.DefaultSubRepoPermsChecker, api.RepoName(args.RepoName), api.CommitID(args.CommitSHA), path)
		},
	)

	metricsCache.SetFileMetrics(context.TODO(), api.RepoID(args.RepoId), api.CommitID(args.CommitSHA), args.Path, metrics, err == nil)

	return metrics.FirstLanguage(), err
}
