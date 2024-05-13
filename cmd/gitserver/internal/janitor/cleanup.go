package janitor

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git/gitcli"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/gitserverfs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/connection"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type JanitorConfig struct {
	JanitorInterval time.Duration
	ShardID         string

	DesiredPercentFree             int
	DisableDeleteReposOnWrongShard bool
}

func NewJanitor(ctx context.Context, cfg JanitorConfig, db database.DB, fs gitserverfs.FS, rcf *wrexec.RecordingCommandFactory, logger log.Logger) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			logger.Info("Starting janitor run")
			// On Sourcegraph.com, we clone repos lazily, meaning whatever github.com
			// repo is visited will be cloned eventually. So over time, we would always
			// accumulate terabytes of repos, of which many are probably not visited
			// often. Thus, we have this special cleanup worker for Sourcegraph.com that
			// will remove repos that have not been changed in a long time (thats the
			// best metric we have here today) once our disks are running full.
			// On customer instances, this worker is useless, because repos are always
			// managed by an external service connection and they will be recloned
			// ASAP.
			if dotcom.SourcegraphDotComMode() {
				func() {
					logger := logger.Scoped("dotcom-repo-cleaner")
					start := time.Now()
					logger.Info("Starting dotcom repo cleaner")

					usage, err := fs.DiskUsage()
					if err != nil {
						logger.Error("getting free disk space", log.Error(err))
						return
					}

					toFree := howManyBytesToFree(logger, usage, cfg.DesiredPercentFree)

					if err := freeUpSpace(ctx, logger, db, fs, cfg.ShardID, usage, cfg.DesiredPercentFree, toFree); err != nil {
						logger.Error("error freeing up space", log.Error(err))
					}

					logger.Info("dotcom repo cleaner finished", log.Int64("toFree", toFree), log.Bool("failed", err != nil), log.String("duration", time.Since(start).String()))
				}()
			}

			gitserverAddrs := connection.NewGitserverAddresses(conf.Get())
			// TODO: Should this return an error?
			cleanupRepos(ctx, logger, db, fs, rcf, cfg.ShardID, gitserverAddrs, cfg.DisableDeleteReposOnWrongShard)

			return nil
		}),
		goroutine.WithName("gitserver.janitor"),
		goroutine.WithDescription("cleans up and maintains repositories regularly"),
		goroutine.WithInterval(cfg.JanitorInterval),
	)
}

var (
	wrongShardReposTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_repo_wrong_shard",
		Help: "The number of repos that are on disk on the wrong shard",
	})
	wrongShardReposDeletedCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repo_wrong_shard_deleted",
		Help: "The number of repos on the wrong shard that we deleted",
	})
)

const (
	day = 24 * time.Hour
	// repoTTLGC is how often we should re-clone a repository once it is
	// reporting git gc issues.
	repoTTLGC = 2 * day
	// gitConfigMaybeCorrupt is a key we add to git config to signal that a repo may be
	// corrupt on disk.
	gitConfigMaybeCorrupt = "sourcegraph.maybeCorruptRepo"
	// The name of the log file placed by sg maintenance in case it encountered an
	// error.
	sgmLog = "sgm.log"
)

// Controls if gitserver cleanup tries to remove repos from disk which are not defined in the DB. Defaults to false.
var removeNonExistingRepos, _ = strconv.ParseBool(env.Get("SRC_REMOVE_NON_EXISTING_REPOS", "false", "controls if gitserver cleanup tries to remove repos from disk which are not defined in the DB"))

var (
	reposRemoved = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_repos_removed",
		Help: "number of repos removed during cleanup",
	}, []string{"reason"})
	reposRecloned = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repos_recloned",
		Help: "number of repos removed and re-cloned due to age",
	})
	reposRemovedDiskPressure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_repos_removed_disk_pressure",
		Help: "number of repos removed due to not enough disk space",
	})
	janitorRunning = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_gitserver_janitor_running",
		Help: "set to 1 when the gitserver janitor background job is running",
	})
	jobTimer = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_gitserver_janitor_job_duration_seconds",
		Help: "Duration of the individual jobs within the gitserver janitor background job",
	}, []string{"success", "job_name"})
	maintenanceStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_maintenance_status",
		Help: "whether the maintenance run was a success (true/false) and the reason why a cleanup was needed",
	}, []string{"success", "reason"})
	pruneStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_gitserver_prune_status",
		Help: "whether git prune was a success (true/false) and whether it was skipped (true/false)",
	}, []string{"success", "skipped"})
	janitorTimer = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "src_gitserver_janitor_duration_seconds",
		Help:    "Duration of gitserver janitor background job",
		Buckets: []float64{0.1, 1, 10, 60, 300, 3600, 7200},
	})
	nonExistingReposRemoved = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_gitserver_non_existing_repos_removed",
		Help: "number of non existing repos removed during cleanup",
	})
)

// cleanupRepos walks the repos directory and performs maintenance tasks:
//
// 1. Compute the amount of space used by the repo
// 2. Remove corrupt repos.
// 3. Remove stale lock files.
// 4. Ensure correct git attributes
// 5. Ensure gc.auto=0 or unset depending on gitGCMode
// 6. Perform garbage collection
// 7. Re-clone repos after a while. (simulate git gc)
// 8. Remove repos based on disk pressure.
// 9. Perform sg-maintenance
// 10. Git prune
// 11. Set sizes of repos
func cleanupRepos(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	fs gitserverfs.FS,
	rcf *wrexec.RecordingCommandFactory,
	shardID string,
	gitServerAddrs connection.GitserverAddresses,
	disableDeleteReposOnWrongShard bool,
) {
	logger = logger.Scoped("cleanup")

	start := time.Now()

	janitorRunning.Set(1)
	defer janitorRunning.Set(0)

	janitorStart := time.Now()
	defer func() {
		janitorTimer.Observe(time.Since(janitorStart).Seconds())
	}()

	knownGitServerShard := false
	for _, addr := range gitServerAddrs.Addresses {
		if hostnameMatch(shardID, addr) {
			knownGitServerShard = true
			break
		}
	}
	if !knownGitServerShard {
		logger.Warn("current shard is not included in the list of known gitserver shards, will not delete repos", log.String("current-hostname", shardID), log.Strings("all-shards", gitServerAddrs.Addresses))
	}

	repoToSize := make(map[api.RepoName]int64)
	var wrongShardRepoCount int64
	defer func() {
		// We want to set the gauge only at the end when we know the total
		wrongShardReposTotal.Set(float64(wrongShardRepoCount))
	}()

	var wrongShardReposDeleted int64
	defer func() {
		// We want to set the gauge only when wrong shard clean-up is enabled
		if disableDeleteReposOnWrongShard {
			wrongShardReposDeletedCounter.Add(float64(wrongShardReposDeleted))
		}
	}()

	maybeDeleteWrongShardRepos := func(repoName api.RepoName, dir common.GitDir) (done bool, err error) {
		// Record the number of repos that should not belong on this instance and
		// remove up to SRC_WRONG_SHARD_DELETE_LIMIT in a single Janitor run.
		addr := gitServerAddrs.AddrForRepo(ctx, repoName)

		if hostnameMatch(shardID, addr) {
			return false, nil
		}

		wrongShardRepoCount++

		// If we're on a shard not currently known, basically every repo would
		// be considered on the wrong shard. This is probably a configuration
		// error and we don't want to completely empty our disk in that case,
		// so skip.
		if !knownGitServerShard {
			return false, nil
		}

		// Check that wrong shard deletion has not been disabled.
		if disableDeleteReposOnWrongShard {
			return false, nil
		}

		logger.Info(
			"removing repo cloned on the wrong shard",
			log.String("dir", string(dir)),
			log.String("target-shard", addr),
			log.String("current-shard", shardID),
		)
		if err := fs.RemoveRepo(repoName); err != nil {
			return true, err
		}

		wrongShardReposDeleted++

		// Note: We just deleted the repo. So we're done with any further janitor tasks!
		return true, nil
	}

	maybeRemoveCorrupt := func(repoName api.RepoName, dir common.GitDir) (done bool, _ error) {
		corrupt, reason, err := checkRepoDirCorrupt(rcf, repoName, dir)
		if !corrupt || err != nil {
			return false, err
		}

		err = db.GitserverRepos().LogCorruption(ctx, repoName, fmt.Sprintf("sourcegraph detected corrupt repo: %s", reason), shardID)
		if err != nil {
			logger.Warn("failed to log repo corruption", log.String("repo", string(repoName)), log.Error(err))
		}

		logger.Info("removing corrupt repo", log.String("repo", string(dir)), log.String("reason", reason))
		if err := fs.RemoveRepo(repoName); err != nil {
			return true, err
		}
		reposRemoved.WithLabelValues(reason).Inc()

		// Set as not_cloned in the database.
		if err := db.GitserverRepos().SetCloneStatus(ctx, repoName, types.CloneStatusNotCloned, shardID); err != nil {
			return true, errors.Wrap(err, "failed to update clone status")
		}

		return true, nil
	}

	maybeRemoveNonExisting := func(repoName api.RepoName, dir common.GitDir) (bool, error) {
		if !removeNonExistingRepos {
			return false, nil
		}

		_, err := db.GitserverRepos().GetByName(ctx, repoName)
		// Repo still exists, nothing to do.
		if err == nil {
			return false, nil
		}

		// Failed to talk to DB, skip this repo.
		if !errcode.IsNotFound(err) {
			logger.Warn("failed to look up repo", log.Error(err), log.String("repo", string(repoName)))
			return false, nil
		}

		// The repo does not exist in the DB (or is soft-deleted), continue deleting it.
		// TODO: For soft-deleted, it might be nice to attempt to update the clone status,
		// but that can only work when we can map a repo on disk back to a repo in DB
		// when the name has been modified to have the DELETED- prefix.
		err = fs.RemoveRepo(repoName)
		if err == nil {
			nonExistingReposRemoved.Inc()
		}
		return true, err
	}

	ensureGitAttributes := func(repoName api.RepoName, dir common.GitDir) (done bool, err error) {
		return false, git.SetGitAttributes(dir)
	}

	maybeReclone := func(repoName api.RepoName, dir common.GitDir) (done bool, err error) {
		backend := gitcli.NewBackend(logger, rcf, dir, repoName)

		repoType, err := git.GetRepositoryType(ctx, backend.Config())
		if err != nil {
			return false, err
		}

		// Add a jitter to spread out re-cloning of repos cloned at the same time.
		var reason string
		const maybeCorrupt = "maybeCorrupt"

		if maybeCorrupt, _ := backend.Config().Get(ctx, gitConfigMaybeCorrupt); maybeCorrupt != "" {
			// Set the reason so that the repo cleaned up
			reason = maybeCorrupt
			// We don't log the corruption here, since the corruption *should* have already been
			// logged when this config setting was set in the repo.
			// When the repo is recloned, the corrupted_at status should be cleared, which means
			// the repo is not considered corrupted anymore.
			//
			// unset flag to stop constantly re-cloning if it fails.
			_ = backend.Config().Unset(ctx, gitConfigMaybeCorrupt)
		}

		// Check if we marked GC as failed and if so, if it's been too long.
		gcFailedAt, err := backend.Config().Get(ctx, gitConfigGCFailed)
		if err != nil {
			return false, errors.Wrap(err, "failed to read git gc fail time")
		}
		if gcFailedAt != "" {
			gcFailedAtInt, err := strconv.Atoi(gcFailedAt)
			if err != nil {
				return false, errors.Wrap(err, "failed to parse git gc fail time")
			}
			firstGCFailure := time.Unix(int64(gcFailedAtInt), 0)
			if time.Since(firstGCFailure) > repoTTLGC+jitterDuration(string(dir), repoTTLGC/4) {
				if gclog, err := os.ReadFile(dir.Path("gc.log")); err == nil && len(gclog) > 0 {
					reason = fmt.Sprintf("git gc %s", string(bytes.TrimSpace(gclog)))
				}
			}
		}

		if (sgmRetries >= 0) && (bestEffortReadFailed(dir) > sgmRetries) {
			if sgmLog, err := os.ReadFile(dir.Path(sgmLog)); err == nil && len(sgmLog) > 0 {
				reason = fmt.Sprintf("sg maintenance, too many retries: %s", string(bytes.TrimSpace(sgmLog)))
			}
		}

		// We believe converting a Perforce depot to a Git repository is generally a
		// very expensive operation, therefore we do not try to re-clone/redo the
		// conversion only because it is old or slow to do "git gc".
		if repoType == "perforce" && reason != maybeCorrupt {
			reason = ""
		}

		if reason == "" {
			return false, nil
		}

		recloneLogger := logger.With(
			log.String("repo", string(repoName)),
			log.String("reason", reason),
		)

		recloneLogger.Info("re-cloning potentially broken repo")

		// We trigger a reclone by removing the repo from disk and marking it as
		// uncloned in the DB. The reclone will then be performed as if this repo
		// was newly added to Sourcegraph.
		// This will make the repo inaccessible for a bit, but we consider the
		// repo completely broken at this stage anways.
		if err := fs.RemoveRepo(repoName); err != nil {
			return true, errors.Wrap(err, "failed to remove repo")
		}
		// Set as not_cloned in the database.
		if err := db.GitserverRepos().SetCloneStatus(ctx, repoName, types.CloneStatusNotCloned, shardID); err != nil {
			return true, errors.Wrap(err, "failed to update clone status")
		}

		reposRecloned.Inc()

		return true, nil
	}

	type cleanupFn struct {
		Name string
		Do   func(api.RepoName, common.GitDir) (bool, error)
	}
	cleanups := []cleanupFn{
		// First, check if we should even be having this repo on disk anymore,
		// maybe there's been a resharding event and we can actually remove it
		// and not spend further CPU cycles fixing it.
		{"delete wrong shard repos", maybeDeleteWrongShardRepos},
		// Do some sanity checks on the repository.
		{"maybe remove corrupt", maybeRemoveCorrupt},
		// Remove repo if DB does not contain it anymore
		{"maybe remove non existing", maybeRemoveNonExisting},
		// We always want to have the same git attributes file at info/attributes.
		{"ensure git attributes", ensureGitAttributes},
		// Enable or disable background garbage collection depending on
		// gitGCMode. The purpose is to avoid repository corruption which can
		// happen if several git-gc operations are running at the same time.
		// We only disable if sg is managing gc.
		{"auto gc config", ensureAutoGC},
	}

	if !conf.Get().DisableAutoGitUpdates {
		// Old git clones accumulate loose git objects that waste space and slow down git
		// operations. Periodically do a fresh clone to avoid these problems. git gc is
		// slow and resource intensive. It is cheaper and faster to just re-clone the
		// repository. We don't do this if DisableAutoGitUpdates is set as it could
		// potentially kick off a clone operation.
		cleanups = append(cleanups, cleanupFn{
			Name: "maybe re-clone",
			Do:   maybeReclone,
		})
	}

	// Compute the amount of space used by the repo. We do this last, because
	// we want it to reflect the improvements that previous GC methods had.
	cleanups = append(cleanups, cleanupFn{"compute stats", collectSize})

	reposCleaned := 0

	err := fs.ForEachRepo(func(repo api.RepoName, gitDir common.GitDir) (done bool) {
		for _, cfn := range cleanups {
			// Check if context has been canceled, if so skip the rest of the repos.
			select {
			case <-ctx.Done():
				logger.Warn("aborting janitor run", log.Error(ctx.Err()))
				return true
			default:
			}

			start := time.Now()
			done, err := cfn.Do(repo, gitDir)
			if err != nil {
				logger.Error("error running cleanup command",
					log.String("name", cfn.Name),
					log.String("repo", string(gitDir)),
					log.Error(err))
			}
			jobTimer.WithLabelValues(strconv.FormatBool(err == nil), cfn.Name).Observe(time.Since(start).Seconds())
			if done {
				break
			}
		}

		reposCleaned++

		// Every 1000 repos, log a progress message.
		if reposCleaned%1000 == 0 {
			logger.Info("Janitor progress", log.Int("repos_cleaned", reposCleaned))
		}

		return false
	})
	if err != nil {
		logger.Error("error iterating over repositories", log.Error(err))
	}

	if len(repoToSize) > 0 {
		_, err := db.GitserverRepos().UpdateRepoSizes(ctx, logger, shardID, repoToSize)
		if err != nil {
			logger.Error("setting repo sizes", log.Error(err))
		}
	}

	logger.Info("Janitor run finished", log.String("duration", time.Since(start).String()))
}

func checkRepoDirCorrupt(rcf *wrexec.RecordingCommandFactory, repoName api.RepoName, dir common.GitDir) (bool, string, error) {
	// We treat repositories missing HEAD to be corrupt. Both our cloning
	// and fetching ensure there is a HEAD file.
	if _, err := os.Stat(dir.Path("HEAD")); os.IsNotExist(err) {
		return true, "missing-head", nil
	} else if err != nil {
		return false, "", err
	}

	// We have seen repository corruption fail in such a way that the git
	// config is missing the bare repo option but everything else looks
	// like it works. This leads to failing fetches, so treat non-bare
	// repos as corrupt. Since we often fetch with ensureRevision, this
	// leads to most commands failing against the repository. It is safer
	// to remove now than try a safe reclone.
	if gitIsNonBareBestEffort(rcf, repoName, dir) {
		return true, "non-bare", nil
	}

	return false, "", nil
}

// gitIsNonBareBestEffort returns true if the repository is not a bare
// repo. If we fail to check or the repository is bare we return false.
//
// Note: it is not always possible to check if a repository is bare since a
// lock file may prevent the check from succeeding. We only want bare
// repositories and want to avoid transient false positives.
func gitIsNonBareBestEffort(rcf *wrexec.RecordingCommandFactory, repoName api.RepoName, dir common.GitDir) bool {
	cmd := exec.Command("git", "-C", dir.Path(), "rev-parse", "--is-bare-repository")
	dir.Set(cmd)
	wrappedCmd := rcf.WrapWithRepoName(context.Background(), log.NoOp(), repoName, cmd)
	b, _ := wrappedCmd.Output()
	b = bytes.TrimSpace(b)
	return bytes.Equal(b, []byte("false"))
}

const gitConfigGCFailed = "sourcegraph.gcFailedAt"

const (
	sgmLogPrefix = "failed="

	sgmLogHeader = `DO NOT EDIT: generated by gitserver.
This file records the number of failed runs of sg maintenance and the
last error message. The number of failed attempts is compared to the
number of allowed retries (see SRC_SGM_RETRIES) to decide whether a
repository should be recloned.`
)

// writeSGMLog writes a log file with the format
//
//	<header>
//
//	<sgmLogPrefix>=<int>
//
//	<error message>
func writeSGMLog(dir common.GitDir, m []byte) error {
	return os.WriteFile(
		dir.Path(sgmLog),
		[]byte(fmt.Sprintf("%s\n\n%s%d\n\n%s\n", sgmLogHeader, sgmLogPrefix, bestEffortReadFailed(dir)+1, m)),
		0600,
	)
}

func bestEffortReadFailed(dir common.GitDir) int {
	b, err := os.ReadFile(dir.Path(sgmLog))
	if err != nil {
		return 0
	}

	return bestEffortParseFailed(b)
}

func bestEffortParseFailed(b []byte) int {
	prefix := []byte(sgmLogPrefix)
	from := bytes.Index(b, prefix)
	if from < 0 {
		return 0
	}

	b = b[from+len(prefix):]
	if to := bytes.IndexByte(b, '\n'); to > 0 {
		b = b[:to]
	}

	n, _ := strconv.Atoi(string(b))
	return n
}

// jitterDuration returns a duration between [0, d) based on key. This is like
// a random duration, but instead of a random source it is computed via a hash
// on key.
func jitterDuration(key string, d time.Duration) time.Duration {
	h := fnv.New64()
	_, _ = io.WriteString(h, key)
	r := time.Duration(h.Sum64())
	if r < 0 {
		// +1 because we have one more negative value than positive. ie
		// math.MinInt64 == -math.MinInt64.
		r = -(r + 1)
	}
	return r % d
}
