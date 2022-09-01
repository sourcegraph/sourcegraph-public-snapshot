package repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/singleflight"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	Sourcer Sourcer
	Worker  *workerutil.Worker
	Store   Store

	// Synced is sent a collection of Repos that were synced by Sync (only if Synced is non-nil)
	Synced chan Diff

	// Logger if non-nil is logged to.
	Logger log.Logger

	// Now is time.Now. Can be set by tests to get deterministic output.
	Now func() time.Time

	// Registerer is the interface to register / unregister prometheus metrics.
	Registerer prometheus.Registerer

	// UserReposMaxPerUser can be used to override the value read from config.
	// If zero, we'll read from config instead.
	UserReposMaxPerUser int

	// UserReposMaxPerSite can be used to override the value read from config.
	// If zero, we'll read from config instead.
	UserReposMaxPerSite int

	// Ensure that we only run one sync per repo at a time
	syncGroup singleflight.Group
}

// RunOptions contains options customizing Run behaviour.
type RunOptions struct {
	EnqueueInterval func() time.Duration // Defaults to 1 minute
	IsCloud         bool                 // Defaults to false
	MinSyncInterval func() time.Duration // Defaults to 1 minute
	DequeueInterval time.Duration        // Default to 10 seconds
}

// Run runs the Sync at the specified interval.
func (s *Syncer) Run(ctx context.Context, store Store, opts RunOptions) error {
	if opts.EnqueueInterval == nil {
		opts.EnqueueInterval = func() time.Duration { return time.Minute }
	}
	if opts.MinSyncInterval == nil {
		opts.MinSyncInterval = func() time.Duration { return time.Minute }
	}
	if opts.DequeueInterval == 0 {
		opts.DequeueInterval = 10 * time.Second
	}

	if !opts.IsCloud {
		s.initialUnmodifiedDiffFromStore(ctx, store)
	}

	worker, resetter := NewSyncWorker(ctx, s.Logger.Scoped("syncWorker", ""), store.Handle(), &syncHandler{
		syncer:          s,
		store:           store,
		minSyncInterval: opts.MinSyncInterval,
	}, SyncWorkerOptions{
		WorkerInterval:       opts.DequeueInterval,
		NumHandlers:          ConfRepoConcurrentExternalServiceSyncers(),
		PrometheusRegisterer: s.Registerer,
		CleanupOldJobs:       true,
	})

	go worker.Start()
	defer worker.Stop()

	go resetter.Start()
	defer resetter.Stop()

	for ctx.Err() == nil {
		if !conf.Get().DisableAutoCodeHostSyncs {
			err := store.EnqueueSyncJobs(ctx, opts.IsCloud)
			if err != nil && s.Logger != nil {
				s.Logger.Error("enqueuing sync jobs", log.Error(err))
			}
		}
		sleep(ctx, opts.EnqueueInterval())
	}

	return ctx.Err()
}

type syncHandler struct {
	syncer          *Syncer
	store           Store
	minSyncInterval func() time.Duration
}

func (s *syncHandler) Handle(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
	sj, ok := record.(*SyncJob)
	if !ok {
		return errors.Errorf("expected repos.SyncJob, got %T", record)
	}

	return s.syncer.SyncExternalService(ctx, sj.ExternalServiceID, s.minSyncInterval())
}

// sleep is a context aware time.Sleep
func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// TriggerExternalServiceSync will enqueue a sync job for the supplied external
// service
func (s *Syncer) TriggerExternalServiceSync(ctx context.Context, id int64) error {
	return s.Store.EnqueueSingleSyncJob(ctx, id)
}

const (
	ownerUndefined = ""
	ownerSite      = "site"
	ownerUser      = "user"
	ownerOrg       = "org"
)

type ErrUnauthorized struct{}

func (e ErrUnauthorized) Error() string {
	return "bad credentials"
}

func (e ErrUnauthorized) Unauthorized() bool {
	return true
}

type ErrForbidden struct{}

func (e ErrForbidden) Error() string {
	return "forbidden"
}

func (e ErrForbidden) Forbidden() bool {
	return true
}

type ErrAccountSuspended struct{}

func (e ErrAccountSuspended) Error() string {
	return "account suspended"
}

func (e ErrAccountSuspended) AccountSuspended() bool {
	return true
}

// initialUnmodifiedDiffFromStore creates a diff of all repos present in the
// store and sends it to s.Synced. This is used so that on startup the reader
// of s.Synced will receive a list of repos. In particular this is so that the
// git update scheduler can start working straight away on existing
// repositories.
func (s *Syncer) initialUnmodifiedDiffFromStore(ctx context.Context, store Store) {
	if s.Synced == nil {
		return
	}

	stored, err := store.RepoStore().List(ctx, database.ReposListOptions{})
	if err != nil {
		if s.Logger != nil {
			s.Logger.Warn("initialUnmodifiedDiffFromStore store.ListRepos", log.Error(err))
		}
		return
	}

	// Assuming sources returns no differences from the last sync, the Diff
	// would be just a list of all stored repos Unmodified. This is the steady
	// state, so is the initial diff we choose.
	select {
	case s.Synced <- Diff{Unmodified: stored}:
	case <-ctx.Done():
	}
}

// Diff is the difference found by a sync between what is in the store and
// what is returned from sources.
type Diff struct {
	Added      types.Repos
	Deleted    types.Repos
	Modified   ReposModified
	Unmodified types.Repos
}

// Sort sorts all Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := range []types.Repos{
		d.Added,
		d.Deleted,
		d.Modified.Repos(),
		d.Unmodified,
	} {
		sort.Sort(ds)
	}
}

// Repos returns all repos in the Diff.
func (d Diff) Repos() types.Repos {
	all := make(types.Repos, 0, len(d.Added)+
		len(d.Deleted)+
		len(d.Modified)+
		len(d.Unmodified))

	for _, rs := range []types.Repos{
		d.Added,
		d.Deleted,
		d.Modified.Repos(),
		d.Unmodified,
	} {
		all = append(all, rs...)
	}

	return all
}

func (d Diff) Len() int {
	return len(d.Deleted) + len(d.Modified) + len(d.Added) + len(d.Unmodified)
}

// RepoModified tracks the modifications applied to a single repository after a
// sync.
type RepoModified struct {
	Repo     *types.Repo
	Modified types.RepoModified
}

type ReposModified []RepoModified

// Repos returns all modified repositories.
func (rm ReposModified) Repos() types.Repos {
	repos := make(types.Repos, len(rm))
	for i := range rm {
		repos[i] = rm[i].Repo
	}

	return repos
}

// ReposModified returns only the repositories that had a specific field
// modified in the sync.
func (rm ReposModified) ReposModified(modified types.RepoModified) types.Repos {
	repos := types.Repos{}
	for _, pair := range rm {
		if pair.Modified&modified == modified {
			repos = append(repos, pair.Repo)
		}
	}

	return repos
}

// SyncRepo syncs a single repository by name and associates it with an external service.
//
// It works for repos from:
//
// 1. Public "cloud_default" code hosts since we don't sync them in the background
//    (which would delete lazy synced repos).
// 2. Any package hosts (i.e. npm, Maven, etc) since callers are expected to store
//    repos in the `lsif_dependency_repos` table which is used as the source of truth
//    for the next full sync, so lazy added repos don't get wiped.
//
// The "background" boolean flag indicates that we should run this
// sync in the background vs block and call s.syncRepo synchronously.
func (s *Syncer) SyncRepo(ctx context.Context, name api.RepoName, background bool) (repo *types.Repo, err error) {
	logger := s.Logger.With(log.String("name", string(name)), log.Bool("background", background))
	tr, ctx := trace.New(ctx, "Syncer.SyncRepo", string(name))
	defer tr.Finish()

	repo, err = s.Store.RepoStore().GetByName(ctx, name)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	codehost := extsvc.CodeHostOf(name, extsvc.PublicCodeHosts...)
	if codehost == nil {
		if repo != nil {
			return repo, nil
		}
		return nil, &database.RepoNotFoundErr{Name: name}
	}

	if repo != nil {
		// Only public repos can be individually synced on sourcegraph.com
		if repo.Private {
			return nil, &database.RepoNotFoundErr{Name: name}
		}
		// Don't sync the repo if it's been updated in the past 1 minute.
		if s.Now().Sub(repo.UpdatedAt) < time.Minute {
			return repo, nil
		}
	}

	if background && repo != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			// We don't care about the return value here, but we still want to ensure that
			// only one is in flight at a time.
			_, err, shared := s.syncGroup.Do(string(name), func() (any, error) {
				return s.syncRepo(ctx, codehost, name, repo)
			})
			if err != nil {
				logger.Error("SyncRepo", log.Error(err), log.Bool("shared", shared))
			}
		}()
		return repo, nil
	}

	updatedRepo, err, shared := s.syncGroup.Do(string(name), func() (any, error) {
		return s.syncRepo(ctx, codehost, name, repo)
	})
	if err != nil {
		logger.Error("SyncRepo", log.Error(err), log.Bool("shared", shared))
		return nil, err
	}
	return updatedRepo.(*types.Repo), nil
}

func (s *Syncer) syncRepo(
	ctx context.Context,
	codehost *extsvc.CodeHost,
	name api.RepoName,
	stored *types.Repo,
) (repo *types.Repo, err error) {
	var svc *types.ExternalService
	ctx, save := s.observeSync(ctx, "Syncer.syncRepo", string(name))
	defer func() { save(svc, err) }()

	svcs, err := s.Store.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{
		Kinds: []string{extsvc.TypeToKind(codehost.ServiceType)},
		// Since package host external services have the set of repositories to sync in
		// the lsif_dependency_repos table, we can lazy-sync individual repos without wiping them
		// out in the next full background sync as long as we add them to that table.
		//
		// This permits lazy-syncing of package repos in on-prem instances as well as in cloud.
		OnlyCloudDefault: !codehost.IsPackageHost(),
		LimitOffset:      &database.LimitOffset{Limit: 1},
	})
	if err != nil {
		return nil, errors.Wrap(err, "listing external services")
	}

	if len(svcs) != 1 {
		return nil, errors.Wrapf(
			&database.RepoNotFoundErr{Name: name},
			"cloud default external service of type %q not found", codehost.ServiceType,
		)
	}

	svc = svcs[0]

	src, err := s.Sourcer(ctx, svc)
	if err != nil {
		return nil, err
	}

	rg, ok := src.(RepoGetter)
	if !ok {
		return nil, errors.Wrapf(
			&database.RepoNotFoundErr{Name: name},
			"can't get repo metadata for service of type %q", codehost.ServiceType,
		)
	}

	path := strings.TrimPrefix(string(name), strings.TrimPrefix(codehost.ServiceID, "https://"))

	if stored != nil {
		defer func() {
			if isDeleteableRepoError(err) {
				err2 := s.Store.DeleteExternalServiceRepo(ctx, svc, stored.ID)
				if err2 != nil {
					s.Logger.Error(
						"SyncRepo failed to delete",
						log.Object("svc", log.String("name", svc.DisplayName), log.Int64("id", svc.ID)),
						log.String("repo", string(name)),
						log.NamedError("cause", err),
						log.Error(err2),
					)
				}
				s.notifyDeleted(ctx, stored.ID)
			}
		}()
	}

	repo, err = rg.GetRepo(ctx, path)
	if err != nil {
		return nil, err
	}

	if repo.Private {
		return nil, &database.RepoNotFoundErr{Name: name}
	}

	if _, err = s.sync(ctx, svc, repo); err != nil {
		return nil, err
	}

	return repo, nil
}

// isDeleteableRepoError checks whether the error returned from a repo sync
// signals that we can safely delete the repo
func isDeleteableRepoError(err error) bool {
	return errcode.IsNotFound(err) || errcode.IsUnauthorized(err) ||
		errcode.IsForbidden(err) || errcode.IsAccountSuspended(err) || errcode.IsUnavailableForLegalReasons(err)
}

// RepoLimitError is produced by Syncer.ExternalServiceSync when a user's sync job
// exceeds the user added repo limits.
type RepoLimitError struct {
	// Number of repos added to site
	SiteAdded uint64

	// Limit of repos that can be added to one site
	SiteLimit uint64

	// Number of repos added by user or org
	ReposCount uint64

	// Limit of repos that can be added by one user or org
	ReposLimit uint64

	// NamespaceUserID of an external service
	UserID int32

	// NamespaceUserID of an external service
	OrgID int32
}

func (e *RepoLimitError) Error() string {
	if e.UserID > 0 {
		return fmt.Sprintf(
			"reached maximum allowed user added repos: site:%d/%d, user:%d/%d (user-id: %d)",
			e.SiteAdded,
			e.SiteLimit,
			e.ReposCount,
			e.ReposLimit,
			e.UserID,
		)
	} else if e.OrgID > 0 {
		return fmt.Sprintf(
			"reached maximum allowed organization added repos: site:%d/%d, organization:%d/%d (org-id: %d)",
			e.SiteAdded,
			e.SiteLimit,
			e.ReposCount,
			e.ReposLimit,
			e.OrgID,
		)
	} else {
		return "expected either userID or orgID to be defined"
	}
}

func (s *Syncer) notifyDeleted(ctx context.Context, deleted ...api.RepoID) {
	var d Diff
	for _, id := range deleted {
		d.Deleted = append(d.Deleted, &types.Repo{ID: id})
	}
	observeDiff(d)

	if s.Synced != nil && d.Len() > 0 {
		select {
		case <-ctx.Done():
		case s.Synced <- d:
		}
	}
}

// ErrCloudDefaultSync is returned by SyncExternalService if an attempt to
// sync a cloud default external service is done. We can't sync these external services
// because their repos are added via the lazy-syncing mechanism on sourcegraph.com
// instead of config (which is empty), so attempting to sync them would delete all of
// the lazy-added repos.
var ErrCloudDefaultSync = errors.New("cloud default external services can't be synced")

// SyncExternalService syncs repos using the supplied external service in a streaming fashion, rather than batch.
// This allows very large sync jobs (i.e. that source potentially millions of repos) to incrementally persist changes.
// Deletes of repositories that were not sourced are done at the end.
func (s *Syncer) SyncExternalService(
	ctx context.Context,
	externalServiceID int64,
	minSyncInterval time.Duration,
) (err error) {
	logger := s.Logger.With(log.Int64("externalServiceID", externalServiceID))
	logger.Info("syncing external service")

	// Ensure the job field is recorded when monitoring external API calls
	ctx = metrics.ContextWithTask(ctx, "SyncExternalService")

	var svc *types.ExternalService
	ctx, save := s.observeSync(ctx, "Syncer.SyncExternalService", "")
	defer func() { save(svc, err) }()

	// We don't use tx here as the sourcing process below can be slow and we don't
	// want to hold a lock on the external_services table for too long.
	svc, err = s.Store.ExternalServiceStore().GetByID(ctx, externalServiceID)
	if err != nil {
		return errors.Wrap(err, "fetching external services")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// From this point we always want to make a best effort attempt to update the
	// service timestamps
	var modified bool
	defer func() {
		now := s.Now()
		interval := calcSyncInterval(now, svc.LastSyncAt, minSyncInterval, modified, err)

		svc.NextSyncAt = now.Add(interval)
		svc.LastSyncAt = now

		// We only want to log this error, not return it
		if err := s.Store.ExternalServiceStore().Upsert(ctx, svc); err != nil {
			logger.Error("upserting external service", log.Error(err))
		}

		logger.Debug("synced external service", log.Duration("backoff duration", interval))
	}()

	// We have fail-safes in place to prevent enqueuing sync jobs for cloud default
	// external services, but in case those fail to prevent a sync for any reason,
	// we have this additional check here. Cloud default external services have their
	// repos added via the lazy-syncing mechanism on sourcegraph.com instead of config
	// (which is empty), so attempting to sync them would delete all of the lazy-added repos.
	if svc.CloudDefault {
		return ErrCloudDefaultSync
	}

	// Unless our site config explicitly allows private code or the user has the
	// "AllowUserExternalServicePrivate" tag, user added external services should
	// only sync public code.
	// Organization owned external services are always considered allowed.
	allowed := func(*types.Repo) bool { return true }
	if svc.NamespaceUserID != 0 {
		if mode, err := database.UsersWith(s.Logger, s.Store).UserAllowedExternalServices(ctx, svc.NamespaceUserID); err != nil {
			return errors.Wrap(err, "checking if user can add private code")
		} else if mode != conf.ExternalServiceModeAll {
			allowed = func(r *types.Repo) bool { return !r.Private }
		}
	}

	src, err := s.Sourcer(ctx, svc)
	if err != nil {
		return err
	}

	results := make(chan SourceResult)

	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	seen := make(map[api.RepoID]struct{})
	var errs error
	fatal := func(err error) bool {
		// If the error is just a warning, then it is not fatal.
		if errors.IsWarning(err) && !errcode.IsAccountSuspended(err) {
			return false
		}

		return errcode.IsUnauthorized(err) ||
			errcode.IsForbidden(err) ||
			errcode.IsAccountSuspended(err)
	}

	logger = s.Logger.With(log.Object("svc", log.String("name", svc.DisplayName), log.Int64("id", svc.ID)))
	// Insert or update repos as they are sourced. Keep track of what was seen
	// so we can remove anything else at the end.
	for res := range results {
		if err := res.Err; err != nil {
			logger.Error("error from codehost", log.Int("seen", len(seen)), log.Error(err))

			errs = errors.Append(errs, errors.Wrapf(err, "fetching from code host %s", svc.DisplayName))
			if fatal(err) {
				// Delete all external service repos of this external service
				logger.Error("stopping external service sync due to fatal error from codehost", log.Error(err))
				seen = map[api.RepoID]struct{}{}
				break
			}

			continue
		}

		sourced := res.Repo
		if !allowed(sourced) {
			continue
		}

		var diff Diff
		if diff, err = s.sync(ctx, svc, sourced); err != nil {
			logger.Error("failed to sync, skipping", log.String("repo", string(sourced.Name)), log.Error(err))
			errs = errors.Append(errs, err)

			// Stop syncing this external service as soon as we know repository limits for user or
			// site level has been exceeded. We want to avoid generating spurious errors here
			// because all subsequent syncs will continue failing unless the limits are increased.
			if errors.HasType(err, &RepoLimitError{}) {
				break
			}

			continue
		}
		for _, r := range diff.Repos() {
			seen[r.ID] = struct{}{}
		}

		modified = modified || len(diff.Modified)+len(diff.Added) > 0

		if conf.Get().ExperimentalFeatures != nil && conf.Get().ExperimentalFeatures.EnableWebhookRepoSync {
			job := &webhookworker.Job{
				RepoID:     int32(sourced.ID),
				RepoName:   string(sourced.Name),
				Org:        getOrgFromRepoName(sourced.Name),
				ExtSvcID:   svc.ID,
				ExtSvcKind: svc.Kind,
			}

			id, err := webhookworker.EnqueueJob(ctx, basestore.NewWithHandle(s.Store.Handle()), job)
			if err != nil {
				logger.Error("unable to enqueue webhook build job")
			}
			logger.Info("enqueued webhook build job", log.Int("ID", id))
		}
	}

	// We don't delete any repos of site-level external services if there were any
	// non-warning errors during a sync.
	//
	// Only user or organization external services will delete
	// repos in a sync run with fatal errors.
	//
	// Site-level external services can own lots of repos and are managed by site admins.
	// It's preferable to have them fix any invalidated token manually rather than deleting the repos automatically.
	deleted := 0

	// If all of our errors are warnings and either Forbidden or Unauthorized,
	// we want to proceed with the deletion. This is to be able to properly sync
	// repos (by removing ones if code-host permissions have changed).
	abortDeletion := false
	if errs != nil {
		var ref errors.MultiError
		if errors.As(errs, &ref) {
			for _, e := range ref.Errors() {
				if errors.IsWarning(e) {
					baseError := errors.Unwrap(e)
					if !errcode.IsForbidden(baseError) && !errcode.IsUnauthorized(baseError) {
						abortDeletion = true
						break
					}
					continue
				}
				abortDeletion = true
				break
			}
		}
	}

	if !abortDeletion || (!svc.IsSiteOwned() && fatal(errs)) {
		// Remove associations and any repos that are no longer associated with any
		// external service.
		//
		// We don't want to delete all repos that weren't seen if we had a lot of
		// spurious errors since that could cause lots of repos to be deleted, only to be
		// added the next sync. We delete only if we had no errors,
		// or all of our errors are warnings and either Forbidden or Unauthorized,
		// or we had one of the fatal errors and the service is not site owned.
		var deletedErr error
		deleted, deletedErr = s.delete(ctx, svc, seen)
		if deletedErr != nil {
			logger.Warn("failed to delete some repos",
				log.Int("seen", len(seen)),
				log.Int("deteled", deleted),
				log.Error(deletedErr),
			)

			errs = errors.Append(errs, errors.Wrap(deletedErr, "some repos couldn't be deleted"))
		}

		if deleted > 0 {
			logger.Warn("deleted not seen repos",
				log.Int("seen", len(seen)),
				log.Int("deleted", deleted),
				log.Error(err),
			)
		}
	}

	modified = modified || deleted > 0

	return errs
}

func (s *Syncer) userReposMaxPerSite() uint64 {
	if n := uint64(s.UserReposMaxPerSite); n > 0 {
		return n
	}
	return uint64(conf.UserReposMaxPerSite())
}

func (s *Syncer) userReposMaxPerUser() uint64 {
	if s.UserReposMaxPerUser == 0 {
		return uint64(conf.UserReposMaxPerUser())
	}
	return uint64(s.UserReposMaxPerUser)
}

// syncs a sourced repo of a given external service, returning a diff with a single repo.
func (s *Syncer) sync(ctx context.Context, svc *types.ExternalService, sourced *types.Repo) (d Diff, err error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return Diff{}, errors.Wrap(err, "syncer: opening transaction")
	}

	defer func() {
		observeDiff(d)
		// We must commit the transaction before publishing to s.Synced
		// so that gitserver finds the repo in the database.
		err = tx.Done(err)
		if err != nil {
			return
		}

		if s.Synced != nil && d.Len() > 0 {
			select {
			case <-ctx.Done():
			case s.Synced <- d:
			}
		}
	}()

	stored, err := tx.RepoStore().List(ctx, database.ReposListOptions{
		Names:          []string{string(sourced.Name)},
		ExternalRepos:  []api.ExternalRepoSpec{sourced.ExternalRepo},
		IncludeBlocked: true,
		IncludeDeleted: true,
		UseOr:          true,
	})
	if err != nil {
		return Diff{}, errors.Wrap(err, "syncer: getting repo from the database")
	}

	switch len(stored) {
	case 2: // Existing repo with a naming conflict
		// Pick this sourced repo to own the name by deleting the other repo. If it still exists, it'll have a different
		// name when we source it from the same code host, and it will be re-created.
		var conflicting, existing *types.Repo
		for _, r := range stored {
			if r.ExternalRepo.Equal(&sourced.ExternalRepo) {
				existing = r
			} else {
				conflicting = r
			}
		}

		// invariant: conflicting can't be nil due to our database constraints
		if err = tx.RepoStore().Delete(ctx, conflicting.ID); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to delete conflicting repo")
		}

		// We fallthrough to the next case after removing the conflicting repo in order to update
		// the winner (i.e. existing). This works because we mutate stored to contain it, which the case expects.
		stored = types.Repos{existing}
		fallthrough
	case 1: // Existing repo, update.
		modified := stored[0].Update(sourced)
		if modified == types.RepoUnmodified {
			d.Unmodified = append(d.Unmodified, stored[0])
			break
		}

		if err = tx.UpdateExternalServiceRepo(ctx, svc, stored[0]); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to update external service repo")
		}

		*sourced = *stored[0]
		d.Modified = append(d.Modified, RepoModified{Repo: stored[0], Modified: modified})
	case 0: // New repo, create.
		if !svc.IsSiteOwned() { // enforce user and org repo limits
			siteAdded, err := tx.CountNamespacedRepos(ctx, 0, 0)
			if err != nil {
				return Diff{}, errors.Wrap(err, "counting total user added repos")
			}

			// get either user ID or org ID. We cannot have both defined at the same time,
			// so this naive addition should work
			userAdded, err := tx.CountNamespacedRepos(ctx, svc.NamespaceUserID, svc.NamespaceOrgID)
			if err != nil {
				return Diff{}, errors.Wrap(err, "counting repos added by user or organization")
			}

			// TODO: For now we are using the same limit for users as for organizations
			userLimit, siteLimit := s.userReposMaxPerUser(), s.userReposMaxPerSite()
			if siteAdded >= siteLimit || userAdded >= userLimit {
				return Diff{}, &RepoLimitError{
					SiteAdded:  siteAdded,
					SiteLimit:  siteLimit,
					ReposCount: userAdded,
					ReposLimit: userLimit,
					UserID:     svc.NamespaceUserID,
					OrgID:      svc.NamespaceOrgID,
				}
			}
		}

		if err = tx.CreateExternalServiceRepo(ctx, svc, sourced); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to create external service repo")
		}

		d.Added = append(d.Added, sourced)
	default: // Impossible since we have two separate unique constraints on name and external repo spec
		panic("unreachable")
	}

	return d, nil
}

func (s *Syncer) delete(ctx context.Context, svc *types.ExternalService, seen map[api.RepoID]struct{}) (int, error) {
	// We do deletion in a best effort manner, returning any errors for individual repos that failed to be deleted.
	deleted, err := s.Store.DeleteExternalServiceReposNotIn(ctx, svc, seen)

	s.notifyDeleted(ctx, deleted...)

	return len(deleted), err
}

func observeDiff(diff Diff) {
	for state, repos := range map[string]types.Repos{
		"added":      diff.Added,
		"modified":   diff.Modified.Repos(),
		"deleted":    diff.Deleted,
		"unmodified": diff.Unmodified,
	} {
		syncedTotal.WithLabelValues(state).Add(float64(len(repos)))
	}
}

func calcSyncInterval(
	now time.Time,
	lastSync time.Time,
	minSyncInterval time.Duration,
	modified bool,
	err error,
) time.Duration {
	const maxSyncInterval = 8 * time.Hour

	// Special case, we've never synced
	if err == nil && (lastSync.IsZero() || modified) {
		return minSyncInterval
	}

	// No change or there were errors, back off
	interval := now.Sub(lastSync) * 2
	if interval < minSyncInterval {
		return minSyncInterval
	}

	if interval > maxSyncInterval {
		return maxSyncInterval
	}

	return interval
}

func (s *Syncer) observeSync(
	ctx context.Context,
	family, title string,
) (context.Context, func(*types.ExternalService, error)) {
	began := s.Now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(svc *types.ExternalService, err error) {
		var owner string
		if svc == nil {
			owner = ownerUndefined
		} else if svc.NamespaceUserID > 0 {
			owner = ownerUser
		} else if svc.NamespaceOrgID > 0 {
			owner = ownerOrg
		} else {
			owner = ownerSite
		}

		syncStarted.WithLabelValues(family, owner).Inc()

		now := s.Now()
		took := s.Now().Sub(began).Seconds()

		lastSync.WithLabelValues(family).Set(float64(now.Unix()))

		success := err == nil
		syncDuration.WithLabelValues(strconv.FormatBool(success), family).Observe(took)

		if !success {
			tr.SetError(err)
			syncErrors.WithLabelValues(family, owner, syncErrorReason(err)).Inc()
		}

		tr.Finish()
	}
}

func syncErrorReason(err error) string {
	switch {
	case err == nil:
		return ""
	case errcode.IsNotFound(err):
		return "not_found"
	case errcode.IsUnauthorized(err):
		return "unauthorized"
	case errcode.IsForbidden(err):
		return "forbidden"
	case errcode.IsTemporary(err):
		return "temporary"
	case strings.Contains(err.Error(), "expected path in npm/(scope/)?name"):
		// This is a known issue which we can filter out for now
		return "invalid_npm_path"
	case strings.Contains(err.Error(), "internal rate limit exceeded"):
		// We want to identify these as it's not an issue communicating with the code
		// host and is most likely caused by temporary traffic spikes.
		return "internal_rate_limit"
	default:
		return "unknown"
	}
}

func getOrgFromRepoName(repoName api.RepoName) string {
	parts := strings.Split(string(repoName), "/")
	if len(parts) == 1 {
		return string(repoName)
	}
	return parts[1]
}
