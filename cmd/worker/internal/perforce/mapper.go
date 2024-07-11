package perforce

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitserverClient interface {
	GetDefaultBranch(ctx context.Context, repo api.RepoName, short bool) (refName string, commit api.CommitID, err error)
	Commits(ctx context.Context, repo api.RepoName, opt gitserver.CommitsOptions) ([]*gitdomain.Commit, error)
}

type perforceChangelistMapper struct {
	cfg    *Config
	db     database.DB
	logger log.Logger
	gs     GitserverClient
}

func (m *perforceChangelistMapper) Handle(ctx context.Context) (errs error) {
	if c := conf.Get(); c.ExperimentalFeatures == nil || c.ExperimentalFeatures.PerforceChangelistMapping != "enabled" {
		return nil
	}

	m.logger.Debug("Indexing Perforce changelist IDs")

	// Iterate over the list of all perforce repos.
	page := 0
	// Initially empty to consider all repos once.
	var lastCheck time.Time
	for {
		nextLastCheck := time.Now()
		rs, err := m.db.Repos().ListMinimalRepos(ctx, database.ReposListOptions{
			ExternalServiceType: extsvc.TypePerforce,
			MinLastChanged:      lastCheck,
			LimitOffset:         &database.LimitOffset{Limit: m.cfg.RepositoryBatchSize, Offset: page * m.cfg.RepositoryBatchSize},
		})
		if err != nil {
			return err
		}
		lastCheck = nextLastCheck

		for _, r := range rs {
			m.logger.Debug("Indexing changelist IDs for repo", log.String("repo", string(r.Name)))
			// Then attempt to index each repo. This is a best effort busy loop
			// for now. Most instances have very few Perforce depots, so this
			// is most likely fine for a good while.
			err := processRepo(ctx, m.logger, m.gs, m.db.RepoCommitsChangelists(), r)
			if err != nil {
				m.logger.Error("failed to process repo", log.Error(err), log.String("repo", string(r.Name)))
				errs = errors.Append(errs, err)
			}
		}

		if len(rs) != m.cfg.RepositoryBatchSize {
			break
		}
		page++
	}

	return errs
}

func processRepo(ctx context.Context, logger log.Logger, gs GitserverClient, store database.RepoCommitsChangelistsStore, repo types.MinimalRepo) error {
	start := time.Now()

	commitsMap, err := getCommitsToInsert(ctx, logger, gs, store, repo)
	if err != nil {
		return errors.Wrap(err, "failed to map perforce changelists")
	}

	// We want to write all the commits or nothing at all in a single transaction to avoid partially
	// successful mapping jobs which will make it difficult to determine missing commits that need to
	// be mapped. This makes it easy to have a reliable start point for the next time this job is
	// attempted, knowing for sure that the latest commit in the DB is indeed the last point from
	// which we need to resume the mapping.
	err = store.BatchInsertCommitSHAsWithPerforceChangelistID(ctx, repo.ID, commitsMap)
	if err != nil {
		return errors.Wrap(err, "failed to insert perforce changelist mappings")
	}

	timeTaken := time.Since(start)
	// NOTE: Hardcoded to log for tasks that run longer than 1 minute. Will revisit this if it
	// becomes noisy under production loads.
	if timeTaken > time.Duration(time.Second*60) {
		logger.Warn("mapping job took long to complete", log.Duration("duration", timeTaken))
	}

	return nil
}

// getCommitsToInsert returns a list of commitsSHA -> changelistID for each commit that is yet to
// be "mapped" in the DB. For new repos, this will contain all the commits and for existing repos it
// will only return the commits yet to be mapped in the DB.
//
// It returns an error if any.
func getCommitsToInsert(ctx context.Context, logger log.Logger, gs GitserverClient, store database.RepoCommitsChangelistsStore, repo types.MinimalRepo) (commitsMap []types.PerforceChangelist, err error) {
	latestRowCommit, err := store.GetLatestForRepo(ctx, repo.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This repo has not been imported into the RepoCommits table yet. Start from the beginning.
			results, err := extractChangelistsFromCommits(ctx, gs, repo.Name, "", "")
			return results, errors.Wrap(err, "failed to import new repo (perforce changelists will have limited functionality)")
		}

		return nil, errors.Wrap(err, "RepoCommits.GetLatestForRepo")
	}

	_, headSHA, err := gs.GetDefaultBranch(ctx, repo.Name, false)
	if err != nil {
		return nil, errors.Wrap(err, "GetDefaultBranch")
	}

	if latestRowCommit != nil && string(latestRowCommit.CommitSHA) == string(headSHA) {
		logger.Info("repo commits already mapped upto HEAD, skipping", log.String("HEAD", string(headSHA)))
		return nil, nil
	}

	results, err := extractChangelistsFromCommits(ctx, gs, repo.Name, string(latestRowCommit.CommitSHA), headSHA)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to import existing repo's commits after HEAD: %q", headSHA)
	}

	return results, nil
}

func extractChangelistsFromCommits(ctx context.Context, gs GitserverClient, repo api.RepoName, lastMappedCommit string, headSHA api.CommitID) ([]types.PerforceChangelist, error) {
	// FIXME: When lastMappedCommit..head is an invalid range.
	// TODO: Follow up in a separate PR.
	ranges := []string{"HEAD"}
	if lastMappedCommit != "" {
		ranges[0] = fmt.Sprintf("%s..%s", lastMappedCommit, headSHA)
	}

	mapped := []types.PerforceChangelist{}
	const pageSize = 50
	page := 0
	for {
		commits, err := gs.Commits(ctx, repo, gitserver.CommitsOptions{
			Ranges: ranges,
			N:      pageSize,
			Skip:   uint(page * pageSize),
		})
		if err != nil {
			return nil, err
		}
		for _, c := range commits {
			cl, err := parseChangelistID(string(c.Message))
			if err != nil {
				return nil, err
			}
			mapped = append(mapped, types.PerforceChangelist{
				CommitSHA:    c.ID,
				ChangelistID: cl,
			})
		}
		if len(commits) < pageSize {
			break
		}
		page++
	}

	return mapped, nil
}

func parseChangelistID(commitMessage string) (changelistID int64, _ error) {
	parsedCID, err := perforce.GetP4ChangelistID(commitMessage)
	if err != nil {
		return 0, err
	}

	cid, err := strconv.ParseInt(parsedCID, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse changelist ID to int64")
	}

	if cid < 0 {
		return 0, errors.New("changelist ID cannot be negative")
	}

	return cid, nil
}
