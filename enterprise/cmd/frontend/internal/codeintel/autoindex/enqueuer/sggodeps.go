package enqueuer

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

// versionPattern matches the form vX.Y.Z-yyyymmddhhmmss-abcdefabcdef
var versionPattern = lazyregexp.New(`^(.*)-(.+)-([a-f0-9]{12})$`)

func (s *IndexEnqueuer) enqueueSourcegraphGoRootDependencies(ctx context.Context, repositoryID int, commit string) error {
	contents, err := s.gitserverClient.RawContents(ctx, repositoryID, commit, "go.mod")
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(contents), "\n") {
		repositoryID, commit, ok, err := s.extractTargetFromGoMod(ctx, line)
		if err != nil {
			log15.Error("Failed to extract dependency", "error", err)
			continue
		}
		if !ok {
			continue
		}

		traceLog := func(fields ...log.Field) {}
		log15.Warn("Queueing dependency for auto-indexing", "repositoryID", repositoryID, "commit", commit)
		if err := s.queueIndexForCommit(ctx, repositoryID, commit, false, traceLog, true); err != nil {
			return err
		}
	}

	return nil
}

func (s *IndexEnqueuer) extractTargetFromGoMod(ctx context.Context, line string) (int, string, bool, error) {
	if parts := strings.Split(line, "=>"); len(parts) > 1 {
		return s.extractTargetFromGoMod(ctx, parts[1])
	}

	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "github.com/") {
		return 0, "", false, nil
	}

	for _, suffix := range []string{"// indirect", "+incompatible"} {
		if strings.HasSuffix(line, suffix) {
			line = strings.TrimSpace(line[:len(line)-len(suffix)])
		}
	}

	parts := strings.Split(line, " ")
	if len(parts) < 2 {
		return 0, "", false, nil
	}

	repoName := api.RepoName(parts[0])
	gitTagOrCommit := parts[1]
	if matches := versionPattern.FindStringSubmatch(gitTagOrCommit); len(matches) > 0 {
		gitTagOrCommit = matches[3]
	}

	repo, err := database.Repos(s.dbStore.Handle().DB()).GetByName(ctx, repoName)
	if err != nil {
		if errcode.IsNotFound(err) {
			log15.Warn("Unknown repository", "repoName", parts[0])
			return 0, "", false, nil
		}

		return 0, "", false, err
	}

	commit, err := git.ResolveRevision(ctx, repoName, gitTagOrCommit, git.ResolveRevisionOptions{})
	if err != nil {
		if errcode.IsNotFound(err) {
			log15.Warn("Unknown revision", "repoName", parts[0], "gitTagOrCommit", gitTagOrCommit)
			return 0, "", false, nil
		}

		return 0, "", false, err
	}

	return int(repo.ID), string(commit), true, nil
}
