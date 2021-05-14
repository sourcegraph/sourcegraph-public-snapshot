package inference

import (
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

const lsifGoImage = "sourcegraph/lsif-go:latest"

type lsifGoJobRecognizer struct{}

var _ IndexJobRecognizer = lsifGoJobRecognizer{}

func (r lsifGoJobRecognizer) CanIndexRepo(paths []string, gitserver GitserverClientWrapper) bool {
	for _, path := range paths {
		if r.canIndexPath(path) {
			return true
		}
	}

	return false
}

func (r lsifGoJobRecognizer) InferIndexJobs(paths []string, gitserver GitserverClientWrapper) (indexes []config.IndexJob) {
	for _, path := range paths {
		if !r.canIndexPath(path) {
			continue
		}

		root := dirWithoutDot(path)

		dockerSteps := []config.DockerStep{
			{
				Root:     root,
				Image:    lsifGoImage,
				Commands: []string{"go mod download"},
			},
		}

		indexes = append(indexes, config.IndexJob{
			Steps:       dockerSteps,
			Root:        root,
			Indexer:     lsifGoImage,
			IndexerArgs: []string{"lsif-go", "--no-animation"},
			Outfile:     "",
		})
	}

	return indexes
}

func (lsifGoJobRecognizer) Patterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		suffixPattern("go.mod"),
		segmentPattern("vendor"),
	}
}

func (r lsifGoJobRecognizer) canIndexPath(path string) bool {
	// TODO(efritz) - support glide, dep, other historic package managers
	// TODO(efritz) - support projects without go.mod but a vendor dir and go files
	return filepath.Base(path) == "go.mod" && containsNoSegments(path, goSegmentBlockList...)
}

var goSegmentBlockList = append([]string{
	"vendor",
}, segmentBlockList...)

var versionPattern = lazyregexp.New(`^(.+)-([a-f0-9]{12})$`)

func (lsifGoJobRecognizer) EnsurePackageRepo(ctx context.Context, pkg semantic.Package, repoUpdater RepoUpdaterClient, gitserver GitserverClient) (int, string, bool, error) {
	if pkg.Scheme != "gomod" || !strings.HasPrefix(pkg.Name, "github.com/") {
		return 0, "", false, nil
	}

	versionString := pkg.Version
	for _, suffix := range []string{"// indirect", "+incompatible"} {
		if strings.HasSuffix(versionString, suffix) {
			versionString = strings.TrimSpace(versionString[:len(versionString)-len(suffix)])
		}
	}

	if matches := versionPattern.FindStringSubmatch(versionString); len(matches) > 0 {
		versionString = matches[2]
	}

	splitRepo := strings.Split(pkg.Name, "/")
	repoName := api.RepoName(splitRepo[0] + "/" + splitRepo[1] + "/" + splitRepo[2])
	repoUpdateResponse, err := repoUpdater.EnqueueRepoUpdate(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			log15.Warn("Unknown repository", "repoName", repoName)
			return 0, "", false, nil
		}

		return 0, "", false, err
	}

	commit, err := gitserver.ResolveRevision(ctx, int(repoUpdateResponse.ID), versionString)
	if err != nil {
		if errcode.IsNotFound(err) {
			log15.Warn("Unknown revision", "repoName", repoName, "gitTagOrCommit", versionString)
			return 0, "", false, nil
		}

		return 0, "", false, err
	}

	return int(repoUpdateResponse.ID), string(commit), true, nil
}

func (r lsifGoJobRecognizer) InferPackageIndexJobs(ctx context.Context, pkg semantic.Package, gitserver GitserverClientWrapper) ([]config.IndexJob, error) {
	paths, err := gitserver.ListFiles(ctx, Patterns)
	if err != nil {
		return nil, err
	}

	return r.InferIndexJobs(paths, gitserver), nil
}
