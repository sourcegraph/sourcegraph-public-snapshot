package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/types"

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

	fileMetrics := r.db.FileMetrics()

	metrics := fileMetrics.GetFileMetrics(context.TODO(), api.RepoID(args.RepoId), args.Path, api.CommitID(args.CommitSHA))

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

	// still ambiguous language detection
	// go for the full tamale
	metrics, err := fileMetrics.CalculateAndStoreFileMetrics(
		context.TODO(),
		types.MinimalRepo{
			ID:   api.RepoID(args.RepoId),
			Name: api.RepoName(args.RepoName),
		},
		args.Path,
		api.CommitID(args.CommitSHA),
	)
	return metrics.FirstLanguage(), err
}
