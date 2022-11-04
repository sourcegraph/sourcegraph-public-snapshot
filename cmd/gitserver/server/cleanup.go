package server

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

//go:embed sg_maintenance.sh
var sgMaintenanceScript string

const (
	// repoTTL is how often we should re-clone a repository.
	repoTTL = time.Hour * 24 * 45
	// repoTTLGC is how often we should re-clone a repository once it is
	// reporting git gc issues.
	repoTTLGC = time.Hour * 24 * 2
	// gitConfigMaybeCorrupt is a key we add to git config to signal that a repo may be
	// corrupt on disk.
	gitConfigMaybeCorrupt = "sourcegraph.maybeCorruptRepo"
	// The name of the log file placed by sg maintenance in case it encountered an
	// error.
	sgmLog = "sgm.log"
)

const (
	// gitGCModeGitAutoGC is when we rely on git running auto gc.
	gitGCModeGitAutoGC int = 1
	// gitGCModeJanitorAutoGC is when during janitor jobs we run git gc --auto.
	gitGCModeJanitorAutoGC = 2
	// gitGCModeMaintenance is when during janitor jobs we run sg maintenance.
	gitGCModeMaintenance = 3
)

// gitGCMode describes which mode we should be running git gc.
var gitGCMode = func() int {
	// EnableGCAuto is a temporary flag that allows us to control whether or not
	// `git gc --auto` is invoked during janitorial activities. This flag will
	// likely evolve into some form of site config value in the future.
	enableGCAuto, _ := strconv.ParseBool(env.Get("SRC_ENABLE_GC_AUTO", "false", "Use git-gc during janitorial cleanup phases"))

	// sg maintenance and git gc must not be enabled at the same time. However, both
	// might be disabled at the same time, hence we need both SRC_ENABLE_GC_AUTO and
	// SRC_ENABLE_SG_MAINTENANCE.
	enableSGMaintenance, _ := strconv.ParseBool(env.Get("SRC_ENABLE_SG_MAINTENANCE", "true", "Use sg maintenance during janitorial cleanup phases"))

	if enableGCAuto && !enableSGMaintenance {
		return gitGCModeJanitorAutoGC
	}

	if enableSGMaintenance && !enableGCAuto {
		return gitGCModeMaintenance
	}

	return gitGCModeGitAutoGC
}()

// The limit of 50 mirrors Git's gc_auto_pack_limit
var autoPackLimit, _ = strconv.Atoi(env.Get("SRC_GIT_AUTO_PACK_LIMIT", "50", "the maximum number of pack files we tolerate before we trigger a repack"))

// Our original Git gc job used 1 as limit, while git's default is 6700. We
// don't want to be too aggressive to avoid unnecessary IO, hence we choose a
// value somewhere in the middle. https://gitlab.com/gitlab-org/gitaly uses a
// limit of 1024, which corresponds to an average of 4 loose objects per folder.
// We can tune this parameter once we gain more experience.
var looseObjectsLimit, _ = strconv.Atoi(env.Get("SRC_GIT_LOOSE_OBJECTS_LIMIT", "1024", "the maximum number of loose objects we tolerate before we trigger a repack"))

// A failed sg maintenance run will place a log file in the git directory.
// Subsequent sg maintenance runs are skipped unless the log file is old.
//
// Based on how https://github.com/git/git handles the gc.log file.
var sgmLogExpire = env.MustGetDuration("SRC_GIT_LOG_FILE_EXPIRY", 24*time.Hour, "the number of hours after which sg maintenance runs even if a log file is present")

// Each failed sg maintenance run increments a counter in the sgmLog file.
// We reclone the repository if the number of retries exceeds sgmRetries.
// Setting SRC_SGM_RETRIES to -1 disables recloning due to sgm failures.
// Default value is 3 (reclone after 3 failed sgm runs).
//
// We mention this ENV variable in the header message of the sgmLog files. Make
// sure that changes here are reflected in sgmLogHeader, too.
var sgmRetries, _ = strconv.Atoi(env.Get("SRC_SGM_RETRIES", "3", "the maximum number of times we retry sg maintenance before triggering a reclone."))

// The limit of repos cloned on the wrong shard to delete in one janitor run - value <=0 disables delete.
var wrongShardReposDeleteLimit, _ = strconv.Atoi(env.Get("SRC_WRONG_SHARD_DELETE_LIMIT", "10", "the maximum number of repos not assigned to this shard we delete in one run"))

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

const reposStatsName = "repos-stats.json"

// cleanupRepos walks the repos directory and performs maintenance tasks:
//
// 1. Compute the amount of space used by the repo
// 2. Remove corrupt repos.
// 3. Remove stale lock files.
// 4. Ensure correct git attributes
// 5. Ensure gc.auto=0 or unset depending on gitGCMode
// 6. Scrub remote URLs
// 7. Perform garbage collection
// 8. Re-clone repos after a while. (simulate git gc)
// 9. Remove repos based on disk pressure.
// 10. Perform sg-maintenance
// 11. Git prune
// 12. Only during first run: Set sizes of repos which don't have it in a database.
func (s *Server) cleanupRepos(gitServerAddrs gitserver.GitServerAddresses) {
	janitorRunning.Set(1)
	janitorStart := time.Now()
	defer func() {
		janitorTimer.Observe(time.Since(janitorStart).Seconds())
	}()
	defer janitorRunning.Set(0)

	logger := s.Logger.Scoped("cleanup", "repositories cleanup operation")

	knownGitServerShard := false
	for _, addr := range gitServerAddrs.Addresses {
		if s.hostnameMatch(addr) {
			knownGitServerShard = true
			break
		}
	}
	if !knownGitServerShard {
		s.Logger.Warn("current shard is not included in the list of known gitserver shards, will not delete repos", log.String("current-hostname", s.Hostname), log.Strings("all-shards", gitServerAddrs.Addresses))
	}

	bCtx, bCancel := s.serverContext()
	defer bCancel()

	stats := protocol.ReposStats{
		UpdatedAt: time.Now(),
	}

	repoToSize := make(map[api.RepoName]int64)
	var wrongShardRepoCount int64
	var wrongShardRepoSize int64
	defer func() {
		// We want to set the gauge only at the end when we know the total
		wrongShardReposTotal.Set(float64(wrongShardRepoCount))
		wrongShardReposSizeTotalBytes.Set(float64(wrongShardRepoSize))
	}()

	var wrongShardReposDeleted int64
	defer func() {
		// We want to set the gauge only when wrong shard clean-up is enabled
		if wrongShardReposDeleteLimit > 0 {
			wrongShardReposDeletedCounter.Add(float64(wrongShardReposDeleted))
		}
	}()

	collectSizeAndMaybeDeleteWrongShardRepos := func(dir GitDir) (done bool, err error) {
		size := dirSize(dir.Path("."))
		stats.GitDirBytes += size
		name := s.name(dir)
		repoToSize[name] = size

		// Record the number and disk usage used of repos that should
		// not belong on this instance and remove up to SRC_WRONG_SHARD_DELETE_LIMIT in a single Janitor run.
		addr, err := s.addrForRepo(bCtx, name, gitServerAddrs)
		if !s.hostnameMatch(addr) {
			wrongShardRepoCount++
			wrongShardRepoSize += size

			if knownGitServerShard && wrongShardReposDeleteLimit > 0 && wrongShardReposDeleted < int64(wrongShardReposDeleteLimit) {
				logger.Info(
					"removing repo cloned on the wrong shard",
					log.String("dir", string(dir)),
					log.String("target-shard", addr),
					log.String("current-shard", s.Hostname),
					log.Int64("size-bytes", size),
				)
				if err := s.removeRepoDirectory(dir, false); err != nil {
					return false, err
				}
				wrongShardReposDeleted++
			}
		}
		return false, nil
	}

	maybeRemoveCorrupt := func(dir GitDir) (done bool, _ error) {
		var reason string

		// We treat repositories missing HEAD to be corrupt. Both our cloning
		// and fetching ensure there is a HEAD file.
		if _, err := os.Stat(dir.Path("HEAD")); os.IsNotExist(err) {
			reason = "missing-head"
		} else if err != nil {
			return false, err
		}

		// We have seen repository corruption fail in such a way that the git
		// config is missing the bare repo option but everything else looks
		// like it works. This leads to failing fetches, so treat non-bare
		// repos as corrupt. Since we often fetch with ensureRevision, this
		// leads to most commands failing against the repository. It is safer
		// to remove now than try a safe reclone.
		if reason == "" && gitIsNonBareBestEffort(dir) {
			reason = "non-bare"
		}

		if reason == "" {
			return false, nil
		}

		s.Logger.Info("removing corrupt repo", log.String("repo", string(dir)), log.String("reason", reason))
		if err := s.removeRepoDirectory(dir, true); err != nil {
			return true, err
		}
		reposRemoved.WithLabelValues(reason).Inc()
		return true, nil
	}

	maybeRemoveNonExisting := func(dir GitDir) (bool, error) {
		if !removeNonExistingRepos {
			return false, nil
		}

		repo, _ := s.DB.GitserverRepos().GetByName(bCtx, s.name(dir))
		if repo == nil {
			err := s.removeRepoDirectory(dir, false)
			if err == nil {
				nonExistingReposRemoved.Inc()
			} else {
				s.Logger.Warn("failed removing repo that is not in DB", log.String("repo", string(dir)))
			}
			return true, err
		}
		return false, nil
	}

	ensureGitAttributes := func(dir GitDir) (done bool, err error) {
		return false, setGitAttributes(dir)
	}

	ensureAutoGC := func(dir GitDir) (done bool, err error) {
		return false, gitSetAutoGC(dir)
	}

	scrubRemoteURL := func(dir GitDir) (done bool, err error) {
		cmd := exec.Command("git", "remote", "remove", "origin")
		dir.Set(cmd)
		// ignore error since we fail if the remote has already been scrubbed.
		_ = cmd.Run()
		return false, nil
	}

	maybeReclone := func(dir GitDir) (done bool, err error) {
		repoType, err := getRepositoryType(dir)
		if err != nil {
			return false, err
		}

		recloneTime, err := getRecloneTime(dir)
		if err != nil {
			return false, err
		}

		// Add a jitter to spread out re-cloning of repos cloned at the same time.
		var reason string
		const maybeCorrupt = "maybeCorrupt"
		if maybeCorrupt, _ := gitConfigGet(dir, gitConfigMaybeCorrupt); maybeCorrupt != "" {
			reason = maybeCorrupt
			// unset flag to stop constantly re-cloning if it fails.
			_ = gitConfigUnset(dir, gitConfigMaybeCorrupt)
		}
		if time.Since(recloneTime) > repoTTL+jitterDuration(string(dir), repoTTL/4) {
			reason = "old"
		}
		if time.Since(recloneTime) > repoTTLGC+jitterDuration(string(dir), repoTTLGC/4) {
			if gclog, err := os.ReadFile(dir.Path("gc.log")); err == nil && len(gclog) > 0 {
				reason = fmt.Sprintf("git gc %s", string(bytes.TrimSpace(gclog)))
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

		ctx, cancel := context.WithTimeout(bCtx, conf.GitLongCommandTimeout())
		defer cancel()

		// name is the relative path to ReposDir, but without the .git suffix.
		repo := s.name(dir)
		recloneLogger := logger.With(
			log.String("repo", string(repo)),
			log.Time("cloned", recloneTime),
			log.String("reason", reason),
		)

		recloneLogger.Info("re-cloning expired repo")

		// update the re-clone time so that we don't constantly re-clone if cloning fails.
		// For example if a repo fails to clone due to being large, we will constantly be
		// doing a clone which uses up lots of resources.
		if err := setRecloneTime(dir, recloneTime.Add(time.Since(recloneTime)/2)); err != nil {
			recloneLogger.Warn("setting backed off re-clone time failed", log.Error(err))
		}

		if _, err := s.cloneRepo(ctx, repo, &cloneOptions{Block: true, Overwrite: true}); err != nil {
			return true, err
		}
		reposRecloned.Inc()
		return true, nil
	}

	removeStaleLocks := func(gitDir GitDir) (done bool, err error) {
		// if removing a lock fails, we still want to try the other locks.
		var multi error

		// config.lock should be held for a very short amount of time.
		if _, err := removeFileOlderThan(logger, gitDir.Path("config.lock"), time.Minute); err != nil {
			multi = errors.Append(multi, err)
		}
		// packed-refs can be held for quite a while, so we are conservative
		// with the age.
		if _, err := removeFileOlderThan(logger, gitDir.Path("packed-refs.lock"), time.Hour); err != nil {
			multi = errors.Append(multi, err)
		}
		// we use the same conservative age for locks inside of refs
		if err := bestEffortWalk(gitDir.Path("refs"), func(path string, fi fs.FileInfo) error {
			if fi.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, ".lock") {
				return nil
			}

			_, err := removeFileOlderThan(logger, path, time.Hour)
			return err
		}); err != nil {
			multi = errors.Append(multi, err)
		}
		// We have seen that, occasionally, commit-graph.locks prevent a git repack from
		// succeeding. Benchmarks on our dogfood cluster have shown that a commit-graph
		// call for a 5GB bare repository takes less than 1 min. The lock is only held
		// during a short period during this time. A 1-hour grace period is very
		// conservative.
		if _, err := removeFileOlderThan(logger, gitDir.Path("objects", "info", "commit-graph.lock"), time.Hour); err != nil {
			multi = errors.Append(multi, err)
		}

		// gc.pid is set by git gc and our sg maintenance script. 24 hours is twice the
		// time git gc uses internally.
		gcPIDMaxAge := 24 * time.Hour
		if foundStale, err := removeFileOlderThan(logger, gitDir.Path(gcLockFile), gcPIDMaxAge); err != nil {
			multi = errors.Append(multi, err)
		} else if foundStale {
			logger.Warn(
				"removeStaleLocks found a stale gc.pid lockfile and removed it. This should not happen and points to a problem with garbage collection. Monitor the repo for possible corruption and verify if this error reoccurs",
				log.String("path", string(gitDir)),
				log.Duration("age", gcPIDMaxAge))
		}

		return false, multi
	}

	performGC := func(dir GitDir) (done bool, err error) {
		return false, gitGC(dir)
	}

	performSGMaintenance := func(dir GitDir) (done bool, err error) {
		return false, sgMaintenance(s.Logger, dir)
	}

	performGitPrune := func(dir GitDir) (done bool, err error) {
		return false, pruneIfNeeded(dir, looseObjectsLimit)
	}

	type cleanupFn struct {
		Name string
		Do   func(GitDir) (bool, error)
	}
	cleanups := []cleanupFn{
		// Compute the amount of space used by the repo
		{"compute stats and delete wrong shard repos", collectSizeAndMaybeDeleteWrongShardRepos},
		// Do some sanity checks on the repository.
		{"maybe remove corrupt", maybeRemoveCorrupt},
		// Remove repo if DB does not contain it anymore
		{"maybe remove non existing", maybeRemoveNonExisting},
		// If git is interrupted it can leave lock files lying around. It does not clean
		// these up, and instead fails commands.
		{"remove stale locks", removeStaleLocks},
		// We always want to have the same git attributes file at info/attributes.
		{"ensure git attributes", ensureGitAttributes},
		// 2021-03-01 (tomas,keegan) we used to store an authenticated remote URL on
		// disk. We no longer need it so we can scrub it.
		{"scrub remote URL", scrubRemoteURL},
		// Enable or disable background garbage collection depending on
		// gitGCMode. The purpose is to avoid repository corruption which can
		// happen if several git-gc operations are running at the same time.
		// We only disable if sg is managing gc.
		{"auto gc config", ensureAutoGC},
	}

	if gitGCMode == gitGCModeJanitorAutoGC {
		// Runs a number of housekeeping tasks within the current repository, such as
		// compressing file revisions (to reduce disk space and increase performance),
		// removing unreachable objects which may have been created from prior
		// invocations of git add, packing refs, pruning reflog, rerere metadata or stale
		// working trees. May also update ancillary indexes such as the commit-graph.
		cleanups = append(cleanups, cleanupFn{"garbage collect", performGC})
	}

	if gitGCMode == gitGCModeMaintenance {
		// Run tasks to optimize Git repository data, speeding up other Git commands and
		// reducing storage requirements for the repository. Note: "garbage collect" and
		// "sg maintenance" must not be enabled at the same time.
		cleanups = append(cleanups, cleanupFn{"sg maintenance", performSGMaintenance})
		cleanups = append(cleanups, cleanupFn{"git prune", performGitPrune})
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

	err := bestEffortWalk(s.ReposDir, func(dir string, fi fs.FileInfo) error {
		if s.ignorePath(dir) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Look for $GIT_DIR
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}

		// We are sure this is a GIT_DIR after the above check
		gitDir := GitDir(dir)

		for _, cfn := range cleanups {
			start := time.Now()
			done, err := cfn.Do(gitDir)
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
		return filepath.SkipDir
	})
	if err != nil {
		logger.Error("error iterating over repositories", log.Error(err))
	}

	if b, err := json.Marshal(stats); err != nil {
		logger.Error("failed to marshal periodic stats", log.Error(err))
	} else if err = os.WriteFile(filepath.Join(s.ReposDir, reposStatsName), b, 0666); err != nil {
		logger.Error("failed to write periodic stats", log.Error(err))
	}

	err = s.setRepoSizes(context.Background(), repoToSize)
	if err != nil {
		logger.Error("setting repo sizes", log.Error(err))
	}

	if s.DiskSizer == nil {
		s.DiskSizer = &StatDiskSizer{}
	}
	b, err := s.howManyBytesToFree()
	if err != nil {
		logger.Error("ensuring free disk space", log.Error(err))
	}
	if err := s.freeUpSpace(b); err != nil {
		logger.Error("error freeing up space", log.Error(err))
	}
}

// setRepoSizes uses calculated sizes of repos to update database entries of repos
// with actual sizes, but only up to 10,000 in one run.
func (s *Server) setRepoSizes(ctx context.Context, repoToSize map[api.RepoName]int64) error {
	logger := s.Logger.Scoped("setRepoSizes", "setRepoSizes does cleanup of database entries")

	reposNumber := len(repoToSize)
	if reposNumber == 0 {
		logger.Info("file system walk didn't yield any directory sizes")
		return nil
	}

	logger.Debug("directory sizes calculated during file system walk",
		log.Int("repoToSize", reposNumber))

	// repos number is limited in order not to overwhelm the database with massive batch updates
	// of every single row of `gitserver_repos` table. This will lead to eventual consistency of
	// repo sizes in the database, but this is totally acceptable.
	if reposNumber > 10000 {
		reposNumber = 10000
	}

	// getting repo IDs for given repo names
	foundRepos, err := s.fetchRepos(ctx, repoToSize, reposNumber)
	if err != nil {
		return err
	}

	reposToUpdate := make(map[api.RepoID]int64)
	for _, repo := range foundRepos {
		if size, exists := repoToSize[repo.Name]; exists {
			reposToUpdate[repo.ID] = size
		}
	}

	// updating repos
	updatedRepos, err := s.DB.GitserverRepos().UpdateRepoSizes(ctx, s.Hostname, reposToUpdate)
	if err != nil {
		return err
	}
	if updatedRepos > 0 {
		logger.Info("repos had their sizes updated", log.Int("updatedRepos", updatedRepos))
	}

	return nil
}

// fetchRepos returns up to count random repos found by names (i.e. keys) in repoToSize map
func (s *Server) fetchRepos(ctx context.Context, repoToSize map[api.RepoName]int64, count int) ([]types.MinimalRepo, error) {
	reposToUpdateNames := make([]string, count)
	idx := 0
	// random nature of map traversal yields a different subset of repos every time this function is called
	for repoName := range repoToSize {
		if idx >= count {
			break
		}
		reposToUpdateNames[idx] = string(repoName)
		idx++
	}

	foundRepos, err := s.DB.Repos().ListMinimalRepos(ctx, database.ReposListOptions{
		Names:          reposToUpdateNames,
		LimitOffset:    &database.LimitOffset{Limit: count},
		IncludeBlocked: true,
	})
	if err != nil {
		return nil, err
	}
	return foundRepos, nil
}

// DiskSizer gets information about disk size and free space.
type DiskSizer interface {
	BytesFreeOnDisk(mountPoint string) (uint64, error)
	DiskSizeBytes(mountPoint string) (uint64, error)
}

// howManyBytesToFree returns the number of bytes that should be freed to make sure
// there is sufficient disk space free to satisfy s.DesiredPercentFree.
func (s *Server) howManyBytesToFree() (int64, error) {
	actualFreeBytes, err := s.DiskSizer.BytesFreeOnDisk(s.ReposDir)
	if err != nil {
		return 0, errors.Wrap(err, "finding the amount of space free on disk")
	}

	// Free up space if necessary.
	diskSizeBytes, err := s.DiskSizer.DiskSizeBytes(s.ReposDir)
	if err != nil {
		return 0, errors.Wrap(err, "getting disk size")
	}
	desiredFreeBytes := uint64(float64(s.DesiredPercentFree) / 100.0 * float64(diskSizeBytes))
	howManyBytesToFree := int64(desiredFreeBytes - actualFreeBytes)
	if howManyBytesToFree < 0 {
		howManyBytesToFree = 0
	}
	const G = float64(1024 * 1024 * 1024)

	s.Logger.Debug("howManyBytesToFree",
		log.Int("desired percent free", s.DesiredPercentFree),
		log.Float64("actual percent free", float64(actualFreeBytes)/float64(diskSizeBytes)*100.0),
		log.Float64("amount to free in GiB", float64(howManyBytesToFree)/G))

	return howManyBytesToFree, nil
}

type StatDiskSizer struct{}

func (s *StatDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	var statFS syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &statFS); err != nil {
		return 0, errors.Wrap(err, "statting")
	}
	free := statFS.Bavail * uint64(statFS.Bsize)
	return free, nil
}

func (s *StatDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	var statFS syscall.Statfs_t
	if err := syscall.Statfs(mountPoint, &statFS); err != nil {
		return 0, errors.Wrap(err, "statting")
	}
	free := statFS.Blocks * uint64(statFS.Bsize)
	return free, nil
}

// freeUpSpace removes git directories under ReposDir, in order from least
// recently to most recently used, until it has freed howManyBytesToFree.
func (s *Server) freeUpSpace(howManyBytesToFree int64) error {
	if howManyBytesToFree <= 0 {
		return nil
	}

	logger := s.Logger.Scoped("cleanup.freeUpSpace", "removes git directories under ReposDir")

	// Get the git directories and their mod times.
	gitDirs, err := s.findGitDirs()
	if err != nil {
		return errors.Wrap(err, "finding git dirs")
	}
	dirModTimes := make(map[GitDir]time.Time, len(gitDirs))
	for _, d := range gitDirs {
		mt, err := gitDirModTime(d)
		if err != nil {
			return errors.Wrap(err, "computing mod time of git dir")
		}
		dirModTimes[d] = mt
	}

	// Sort the repos from least to most recently used.
	sort.Slice(gitDirs, func(i, j int) bool {
		return dirModTimes[gitDirs[i]].Before(dirModTimes[gitDirs[j]])
	})

	// Remove repos until howManyBytesToFree is met or exceeded.
	var spaceFreed int64
	diskSizeBytes, err := s.DiskSizer.DiskSizeBytes(s.ReposDir)
	if err != nil {
		return errors.Wrap(err, "getting disk size")
	}
	for _, d := range gitDirs {
		if spaceFreed >= howManyBytesToFree {
			return nil
		}
		delta := dirSize(d.Path("."))
		if err := s.removeRepoDirectory(d, true); err != nil {
			return errors.Wrap(err, "removing repo directory")
		}
		spaceFreed += delta
		reposRemovedDiskPressure.Inc()

		// Report the new disk usage situation after removing this repo.
		actualFreeBytes, err := s.DiskSizer.BytesFreeOnDisk(s.ReposDir)
		if err != nil {
			return errors.Wrap(err, "finding the amount of space free on disk")
		}
		G := float64(1024 * 1024 * 1024)

		logger.Warn("removed least recently used repo",
			log.String("repo", string(d)),
			log.Duration("how old", time.Since(dirModTimes[d])),
			log.Float64("free space in GiB", float64(actualFreeBytes)/G),
			log.Float64("actual percent of disk space free", float64(actualFreeBytes)/float64(diskSizeBytes)*100.0),
			log.Float64("desired percent of disk space free", float64(s.DesiredPercentFree)),
			log.Float64("space freed in GiB", float64(spaceFreed)/G),
			log.Float64("how much space to free in GiB", float64(howManyBytesToFree)/G))
	}

	// Check.
	if spaceFreed < howManyBytesToFree {
		return errors.Errorf("only freed %d bytes, wanted to free %d", spaceFreed, howManyBytesToFree)
	}
	return nil
}

func gitDirModTime(d GitDir) (time.Time, error) {
	head, err := os.Stat(d.Path("HEAD"))
	if err != nil {
		return time.Time{}, errors.Wrap(err, "getting repository modification time")
	}
	return head.ModTime(), nil
}

func (s *Server) findGitDirs() ([]GitDir, error) {
	var dirs []GitDir
	err := bestEffortWalk(s.ReposDir, func(path string, fi fs.FileInfo) error {
		if s.ignorePath(path) {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !fi.IsDir() || fi.Name() != ".git" {
			return nil
		}
		dirs = append(dirs, GitDir(path))
		return filepath.SkipDir
	})
	if err != nil {
		return nil, errors.Wrap(err, "findGitDirs")
	}
	return dirs, nil
}

// dirSize returns the total size in bytes of all the files under d.
func dirSize(d string) int64 {
	var size int64
	// We don't return an error, so we know that err is always nil and can be
	// ignored.
	_ = bestEffortWalk(d, func(path string, fi fs.FileInfo) error {
		if fi.IsDir() {
			return nil
		}
		size += fi.Size()
		return nil
	})
	return size
}

// removeRepoDirectory atomically removes a directory from s.ReposDir.
//
// It first moves the directory to a temporary location to avoid leaving
// partial state in the event of server restart or concurrent modifications to
// the directory.
//
// Additionally, it removes parent empty directories up until s.ReposDir.
func (s *Server) removeRepoDirectory(gitDir GitDir, updateCloneStatus bool) error {
	ctx := context.Background()
	dir := string(gitDir)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// If directory doesn't exist we can avoid all the work below and treat it as if
		// it was removed.
		return nil
	}

	// Rename out of the location so we can atomically stop using the repo.
	tmp, err := s.tempDir("delete-repo")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	if err := fileutil.RenameAndSync(dir, filepath.Join(tmp, "repo")); err != nil {
		return err
	}

	// Everything after this point is just cleanup, so any error that occurs
	// should not be returned, just logged.

	// Set as not_cloned in the database.
	if updateCloneStatus {
		s.setCloneStatusNonFatal(ctx, s.name(gitDir), types.CloneStatusNotCloned)
	}

	// Cleanup empty parent directories. We just attempt to remove and if we
	// have a failure we assume it's due to the directory having other
	// children. If we checked first we could race with someone else adding a
	// new clone.
	rootInfo, err := os.Stat(s.ReposDir)
	if err != nil {
		s.Logger.Warn("Failed to stat ReposDir", log.Error(err))
		return nil
	}
	current := dir
	for {
		parent := filepath.Dir(current)
		if parent == current {
			// This shouldn't happen, but protecting against escaping
			// ReposDir.
			break
		}
		current = parent
		info, err := os.Stat(current)
		if os.IsNotExist(err) {
			// Someone else beat us to it.
			break
		}
		if err != nil {
			s.Logger.Warn("failed to stat parent directory", log.String("dir", current), log.Error(err))
			return nil
		}
		if os.SameFile(rootInfo, info) {
			// Stop, we are at the parent.
			break
		}

		if err := os.Remove(current); err != nil {
			// Stop, we assume remove failed due to current not being empty.
			break
		}
	}

	// Delete the atomically renamed dir. We do this last since if it fails we
	// will rely on a janitor job to clean up for us.
	if err := os.RemoveAll(filepath.Join(tmp, "repo")); err != nil {
		s.Logger.Warn("failed to cleanup after removing dir", log.String("dir", dir), log.Error(err))
	}

	return nil
}

// cleanTmpFiles tries to remove tmp_pack_* files from .git/objects/pack.
// These files can be created by an interrupted fetch operation,
// and would be purged by `git gc --prune=now`, but `git gc` is
// very slow. Removing these files while they're in use will cause
// an operation to fail, but not damage the repository.
func (s *Server) cleanTmpFiles(dir GitDir) {
	now := time.Now()
	packdir := dir.Path("objects", "pack")
	err := bestEffortWalk(packdir, func(path string, info fs.FileInfo) error {
		if path != packdir && info.IsDir() {
			return filepath.SkipDir
		}
		file := filepath.Base(path)
		if strings.HasPrefix(file, "tmp_pack_") {
			if now.Sub(info.ModTime()) > conf.GitLongCommandTimeout() {
				err := os.Remove(path)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		s.Logger.Error("error removing tmp_pack_* files", log.Error(err))
	}
}

// SetupAndClearTmp sets up the the tempdir for ReposDir as well as clearing it
// out. It returns the temporary directory location.
func (s *Server) SetupAndClearTmp() (string, error) {
	logger := s.Logger.Scoped("cleanup.SetupAndClearTmp", "sets up the the tempdir for ReposDir as well as clearing it out")

	// Additionally we create directories with the prefix .tmp-old which are
	// asynchronously removed. We do not remove in place since it may be a
	// slow operation to block on. Our tmp dir will be ${s.ReposDir}/.tmp
	dir := filepath.Join(s.ReposDir, tempDirName) // .tmp
	oldPrefix := tempDirName + "-old"
	if _, err := os.Stat(dir); err == nil {
		// Rename the current tmp file so we can asynchronously remove it. Use
		// a consistent pattern so if we get interrupted, we can clean it
		// another time.
		oldTmp, err := os.MkdirTemp(s.ReposDir, oldPrefix)
		if err != nil {
			return "", err
		}
		// oldTmp dir exists, so we need to use a child of oldTmp as the
		// rename target.
		if err := os.Rename(dir, filepath.Join(oldTmp, tempDirName)); err != nil {
			return "", err
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}

	// Asynchronously remove old temporary directories
	files, err := os.ReadDir(s.ReposDir)
	if err != nil {
		logger.Error("failed to do tmp cleanup", log.Error(err))
	} else {
		for _, f := range files {
			// Remove older .tmp directories as well as our older tmp-
			// directories we would place into ReposDir. In September 2018 we
			// can remove support for removing tmp- directories.
			if !strings.HasPrefix(f.Name(), oldPrefix) && !strings.HasPrefix(f.Name(), "tmp-") {
				continue
			}
			go func(path string) {
				if err := os.RemoveAll(path); err != nil {
					logger.Error("failed to remove old temporary directory", log.String("path", path), log.Error(err))
				}
			}(filepath.Join(s.ReposDir, f.Name()))
		}
	}

	return dir, nil
}

// setRepositoryType sets the type of the repository.
func setRepositoryType(dir GitDir, typ string) error {
	return gitConfigSet(dir, "sourcegraph.type", typ)
}

// getRepositoryType returns the type of the repository.
func getRepositoryType(dir GitDir) (string, error) {
	val, err := gitConfigGet(dir, "sourcegraph.type")
	if err != nil {
		return "", err
	}
	return val, nil
}

// setRecloneTime sets the time a repository is cloned.
func setRecloneTime(dir GitDir, now time.Time) error {
	err := gitConfigSet(dir, "sourcegraph.recloneTimestamp", strconv.FormatInt(now.Unix(), 10))
	if err != nil {
		ensureHEAD(dir)
		return errors.Wrap(err, "failed to update recloneTimestamp")
	}
	return nil
}

// getRecloneTime returns an approximate time a repository is cloned. If the
// value is not stored in the repository, the re-clone time for the repository is
// set to now.
func getRecloneTime(dir GitDir) (time.Time, error) {
	// We store the time we re-cloned the repository. If the value is missing,
	// we store the current time. This decouples this timestamp from the
	// different ways a clone can appear in gitserver.
	update := func() (time.Time, error) {
		now := time.Now()
		return now, setRecloneTime(dir, now)
	}

	value, err := gitConfigGet(dir, "sourcegraph.recloneTimestamp")
	if err != nil {
		return time.Unix(0, 0), errors.Wrap(err, "failed to determine clone timestamp")
	}
	if value == "" {
		return update()
	}

	sec, err := strconv.ParseInt(value, 10, 0)
	if err != nil {
		// If the value is bad update it to the current time
		now, err2 := update()
		if err2 != nil {
			err = err2
		}
		return now, err
	}

	return time.Unix(sec, 0), nil
}

func checkMaybeCorruptRepo(logger log.Logger, repo api.RepoName, dir GitDir, stderr string) {
	if !stdErrIndicatesCorruption(stderr) {
		return
	}

	logger = logger.With(log.String("repo", string(repo)), log.String("dir", string(dir)))

	logger.Warn("marking repo for re-cloning due to stderr output indicating repo corruption",
		log.String("stderr", stderr))

	// We set a flag in the config for the cleanup janitor job to fix. The janitor
	// runs every minute.
	err := gitConfigSet(dir, gitConfigMaybeCorrupt, strconv.FormatInt(time.Now().Unix(), 10))
	if err != nil {
		logger.Error("failed to set maybeCorruptRepo config", log.Error(err))
	}
}

// stdErrIndicatesCorruption returns true if the provided stderr output from a git command indicates
// that there might be repository corruption.
func stdErrIndicatesCorruption(stderr string) bool {
	return objectOrPackFileCorruptionRegex.MatchString(stderr) || commitGraphCorruptionRegex.MatchString(stderr)
}

var (
	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate that
	// that a repository's packfiles or commit objects might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/6676 for more
	// context.
	objectOrPackFileCorruptionRegex = lazyregexp.NewPOSIX(`^error: (Could not read|packfile) `)

	// objectOrPackFileCorruptionRegex matches stderr lines from git which indicate that
	// git's supplemental commit-graph might be corrupted.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/37872 for more
	// context.
	commitGraphCorruptionRegex = lazyregexp.NewPOSIX(`^fatal: commit-graph requires overflow generation data but has none`)
)

// gitIsNonBareBestEffort returns true if the repository is not a bare
// repo. If we fail to check or the repository is bare we return false.
//
// Note: it is not always possible to check if a repository is bare since a
// lock file may prevent the check from succeeding. We only want bare
// repositories and want to avoid transient false positives.
func gitIsNonBareBestEffort(dir GitDir) bool {
	cmd := exec.Command("git", "-C", dir.Path(), "rev-parse", "--is-bare-repository")
	dir.Set(cmd)
	b, _ := cmd.Output()
	b = bytes.TrimSpace(b)
	return bytes.Equal(b, []byte("false"))
}

// gitGC will invoke `git-gc` to clean up any garbage in the repo. It will
// operate synchronously and be aggressive with its internal heuristics when
// deciding to act (meaning it will act now at lower thresholds).
func gitGC(dir GitDir) error {
	cmd := exec.Command("git", "-c", "gc.auto=1", "-c", "gc.autoDetach=false", "gc", "--auto")
	dir.Set(cmd)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to git-gc")
	}
	return nil
}

const (
	sgmLogPrefix = "failed="

	sgmLogHeader = `DO NOT EDIT: generated by gitserver.
This file records the number of failed runs of sg maintenance and the
last error message. The number of failed attempts is compared to the
number of allowed retries (see SRC_SGM_RETRIES) to decide whether a
repository should be recloned.`
)

// writeSGMLog writes a log file with the format
// 		<header>
//
// 		<sgmLogPrefix>=<int>
//
// 		<error message>
//
func writeSGMLog(dir GitDir, m []byte) error {
	return os.WriteFile(
		dir.Path(sgmLog),
		[]byte(fmt.Sprintf("%s\n\n%s%d\n\n%s\n", sgmLogHeader, sgmLogPrefix, bestEffortReadFailed(dir)+1, m)),
		0600,
	)
}

func bestEffortReadFailed(dir GitDir) int {
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

// sgMaintenance runs a set of git cleanup tasks in dir. This must not be run
// concurrently with git gc. sgMaintenance will check the state of the repository
// to avoid running the cleanup tasks if possible. If a sgmLog file is present in
// dir, sgMaintenance will not run unless the file is old.
func sgMaintenance(logger log.Logger, dir GitDir) (err error) {
	// Don't run if sgmLog file is younger than sgmLogExpire hours. There is no need
	// to report an error, because the error has already been logged in a previous
	// run.
	if fi, err := os.Stat(dir.Path(sgmLog)); err == nil {
		if fi.ModTime().After(time.Now().Add(-sgmLogExpire)) {
			return nil
		}
	}
	needed, reason, err := needsMaintenance(dir)
	defer func() {
		maintenanceStatus.WithLabelValues(strconv.FormatBool(err == nil), reason).Inc()
	}()
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}

	cmd := exec.Command("sh")
	dir.Set(cmd)

	cmd.Stdin = strings.NewReader(sgMaintenanceScript)

	err, unlock := lockRepoForGC(dir)
	if err != nil {
		logger.Debug(
			"could not lock repository for sg maintenance",
			log.String("dir", string(dir)),
			log.Error(err),
		)
		return nil
	}
	defer unlock()

	b, err := cmd.CombinedOutput()
	if err != nil {
		if err := writeSGMLog(dir, b); err != nil {
			logger.Debug("sg maintenance failed to write log file", log.String("file", dir.Path(sgmLog)), log.Error(err))
		}
		logger.Debug("sg maintenance", log.String("dir", string(dir)), log.String("out", string(b)))
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to run sg maintenance")
	}
	// Remove the log file after a successful run.
	_ = os.Remove(dir.Path(sgmLog))
	return nil
}

const gcLockFile = "gc.pid"

func lockRepoForGC(dir GitDir) (error, func() error) {
	// Setting permissions to 644 to mirror the permissions that git gc sets for gc.pid.
	f, err := os.OpenFile(dir.Path(gcLockFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		content, err1 := os.ReadFile(dir.Path(gcLockFile))
		if err1 != nil {
			return err, nil
		}
		pidMachine := strings.Split(string(content), " ")
		if len(pidMachine) < 2 {
			return err, nil
		}
		return errors.Wrapf(err, "process %s on machine %s is already running a gc operation", pidMachine[0], pidMachine[1]), nil
	}

	// We cut the hostname to 256 bytes, just like git gc does. See HOST_NAME_MAX in
	// github.com/git/git.
	name := hostname.Get()
	hostNameMax := 256
	if len(name) > hostNameMax {
		name = name[0:hostNameMax]
	}

	_, err = fmt.Fprintf(f, "%d %s", os.Getpid(), name)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}

	return err, func() error {
		return os.Remove(dir.Path(gcLockFile))
	}
}

// We run git-prune only if there are enough loose objects. This approach is
// adapted from https://gitlab.com/gitlab-org/gitaly.
func pruneIfNeeded(dir GitDir, limit int) (err error) {
	needed, err := tooManyLooseObjects(dir, limit)
	defer func() {
		pruneStatus.WithLabelValues(strconv.FormatBool(err == nil), strconv.FormatBool(!needed)).Inc()
	}()
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}

	// "--expire now" will remove all unreachable, loose objects from the store. The
	// default setting is 2 weeks. We choose a more aggressive setting because
	// unreachable, loose objects count towards the threshold that triggers a
	// repack. In the worst case, IE all loose objects are unreachable, we would
	// continuously trigger repacks until the loose objects expire.
	cmd := exec.Command("git", "prune", "--expire", "now")
	dir.Set(cmd)
	err = cmd.Run()
	if err != nil {
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to git-prune")
	}
	return nil
}

func needsMaintenance(dir GitDir) (bool, string, error) {
	// Bitmaps store reachability information about the set of objects in a
	// packfile which speeds up clone and fetch operations.
	hasBm, err := hasBitmap(dir)
	if err != nil {
		return false, "", err
	}
	if !hasBm {
		return true, "bitmap", nil
	}

	// The commit-graph file is a supplemental data structure that accelerates
	// commit graph walks triggered EG by git-log.
	hasCg, err := hasCommitGraph(dir)
	if err != nil {
		return false, "", err
	}
	if !hasCg {
		return true, "commit_graph", nil
	}

	tooManyPf, err := tooManyPackfiles(dir, autoPackLimit)
	if err != nil {
		return false, "", err
	}
	if tooManyPf {
		return true, "packfiles", nil
	}

	tooManyLO, err := tooManyLooseObjects(dir, looseObjectsLimit)
	if err != nil {
		return false, "", err
	}
	if tooManyLO {
		return tooManyLO, "loose_objects", nil
	}
	return false, "skipped", nil
}

var reHexadecimal = lazyregexp.New("^[0-9a-f]+$")

// tooManyLooseObjects follows Git's approach of estimating the number of
// loose objects by counting the objects in a sentinel folder and extrapolating
// based on the assumption that loose objects are randomly distributed in the
// 256 possible folders.
func tooManyLooseObjects(dir GitDir, limit int) (bool, error) {
	// We use the same folder git uses to estimate the number of loose objects.
	objs, err := os.ReadDir(filepath.Join(dir.Path(), "objects", "17"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, errors.Wrap(err, "tooManyLooseObjects")
	}

	count := 0
	for _, obj := range objs {
		// Git checks if the file names are hexadecimal and that they have the right
		// length depending on the chosen hash algorithm. Since the hash algorithm might
		// change over time, checking the length seems too brittle. Instead, we just
		// count all files with hexadecimal names.
		if obj.IsDir() {
			continue
		}
		if matches := reHexadecimal.MatchString(obj.Name()); !matches {
			continue
		}
		count++
	}
	return count*256 > limit, nil
}

func hasBitmap(dir GitDir) (bool, error) {
	bitmaps, err := filepath.Glob(dir.Path("objects", "pack", "*.bitmap"))
	if err != nil {
		return false, err
	}
	return len(bitmaps) > 0, nil
}

func hasCommitGraph(dir GitDir) (bool, error) {
	if _, err := os.Stat(dir.Path("objects", "info", "commit-graph")); err == nil {
		return true, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// tooManyPackfiles counts the packfiles in objects/pack. Packfiles with an
// accompanying .keep file are ignored.
func tooManyPackfiles(dir GitDir, limit int) (bool, error) {
	packs, err := filepath.Glob(dir.Path("objects", "pack", "*.pack"))
	if err != nil {
		return false, err
	}
	count := 0
	for _, p := range packs {
		// Because we know p has the extension .pack, we can slice it off directly
		// instead of using strings.TrimSuffix and filepath.Ext. Benchmarks showed that
		// this option is 20x faster than strings.TrimSuffix(file, filepath.Ext(file))
		// and 17x faster than file[:strings.LastIndex(file, ".")]. However, the runtime
		// of all options is dominated by adding the extension ".keep".
		keepFile := p[:len(p)-5] + ".keep"
		if _, err := os.Stat(keepFile); err == nil {
			continue
		}
		count++
	}
	return count > limit, nil
}

// gitSetAutoGC will set the value of gc.auto. If GC is managed by Sourcegraph
// the value will be 0 (disabled), otherwise if managed by git we will unset
// it to rely on default (on) or global config.
//
// The purpose is to avoid repository corruption which can happen if several
// git-gc operations are running at the same time.
func gitSetAutoGC(dir GitDir) error {
	switch gitGCMode {
	case gitGCModeGitAutoGC:
		return gitConfigUnset(dir, "gc.auto")

	case gitGCModeJanitorAutoGC, gitGCModeMaintenance:
		return gitConfigSet(dir, "gc.auto", "0")

	default:
		// should not happen
		panic(fmt.Sprintf("non exhaustive switch for gitGCMode: %d", gitGCMode))
	}
}

func gitConfigGet(dir GitDir, key string) (string, error) {
	cmd := exec.Command("git", "config", "--get", key)
	dir.Set(cmd)
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 1 {
			return "", nil
		}
		return "", errors.Wrapf(wrapCmdError(cmd, err), "failed to get git config %s", key)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitConfigSet(dir GitDir, key, value string) error {
	cmd := exec.Command("git", "config", key, value)
	dir.Set(cmd)
	err := cmd.Run()
	if err != nil {
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to set git config %s", key)
	}
	return nil
}

func gitConfigUnset(dir GitDir, key string) error {
	cmd := exec.Command("git", "config", "--unset-all", key)
	dir.Set(cmd)
	err := cmd.Run()
	if err != nil {
		// Exit code 5 means the key is not set.
		var e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscall.WaitStatus).ExitStatus() == 5 {
			return nil
		}
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to unset git config %s", key)
	}
	return nil
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

// wrapCmdError will wrap errors for cmd to include the arguments. If the error
// is an exec.ExitError and cmd was invoked with Output(), it will also include
// the captured stderr.
func wrapCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		return errors.Wrapf(err, "%s %s failed with stderr: %s", cmd.Path, strings.Join(cmd.Args, " "), string(e.Stderr))
	}
	return errors.Wrapf(err, "%s %s failed", cmd.Path, strings.Join(cmd.Args, " "))
}

// removeFileOlderThan removes path if its mtime is older than maxAge. If the
// file is missing, no error is returned. The first argument indicates whether a
// stale file was present.
func removeFileOlderThan(logger log.Logger, path string, maxAge time.Duration) (bool, error) {
	fi, err := os.Stat(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	age := time.Since(fi.ModTime())
	if age < maxAge {
		return false, nil
	}

	logger.Debug("removing stale lock file", log.String("path", path), log.Duration("age", age))
	err = os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return true, err
	}
	return true, nil
}

func mockRemoveNonExistingReposConfig(value bool) {
	removeNonExistingRepos = value
}
