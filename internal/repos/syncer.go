package repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	Sourcer Sourcer
	Worker  *workerutil.Worker
	Store   *Store

	// Synced is sent a collection of Repos that were synced by Sync (only if Synced is non-nil)
	Synced chan Diff

	// Logger if non-nil is logged to.
	Logger log15.Logger

	// Now is time.Now. Can be set by tests to get deterministic output.
	Now func() time.Time

	Registerer prometheus.Registerer

	// UserReposMaxPerUser can be used to override the value read from config.
	// If zero, we'll read from config instead.
	UserReposMaxPerUser int

	// UserReposMaxPerSite can be used to override the value read from config.
	// If zero, we'll read from config instead.
	UserReposMaxPerSite int
}

// RunOptions contains options customizing Run behaviour.
type RunOptions struct {
	EnqueueInterval func() time.Duration // Defaults to 1 minute
	IsCloud         bool                 // Defaults to false
	MinSyncInterval func() time.Duration // Defaults to 1 minute
	DequeueInterval time.Duration        // Default to 10 seconds
	// Run each job in a transaction. We default to false in production
	// because a sync job can take a very long time, and we want incremental updates
	// and inserts to show up. This is used for tests only.
	Transact bool
}

// Run runs the Sync at the specified interval.
func (s *Syncer) Run(ctx context.Context, store *Store, opts RunOptions) error {
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

	worker, resetter := NewSyncWorker(ctx, store.Handle().DB(), &syncHandler{
		syncer:          s,
		store:           store,
		minSyncInterval: opts.MinSyncInterval,
		transact:        opts.Transact,
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
			if err != nil {
				s.log().Error("Enqueuing sync jobs", "error", err)
			}
		}
		sleep(ctx, opts.EnqueueInterval())
	}

	return ctx.Err()
}

var discardLogger = func() log15.Logger {
	l := log15.New()
	l.SetHandler(log15.DiscardHandler())
	return l
}()

func (s *Syncer) log() log15.Logger {
	if s.Logger == nil {
		return discardLogger
	}
	return s.Logger
}

type syncHandler struct {
	syncer          *Syncer
	store           *Store
	transact        bool
	minSyncInterval func() time.Duration
}

func (s *syncHandler) Handle(ctx context.Context, tx dbworkerstore.Store, record workerutil.Record) (err error) {
	sj, ok := record.(*SyncJob)
	if !ok {
		return fmt.Errorf("expected repos.SyncJob, got %T", record)
	}

	store := s.store
	if s.transact {
		store = store.With(tx)
	}

	return s.syncer.SyncExternalService(ctx, store, sj.ExternalServiceID, s.minSyncInterval())
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

type externalServiceOwnerType string

const (
	ownerUndefined externalServiceOwnerType = ""
	ownerSite      externalServiceOwnerType = "site"
	ownerUser      externalServiceOwnerType = "user"
)

// SyncExternalService syncs repos using the supplied external service.
func (s *Syncer) SyncExternalService(ctx context.Context, tx *Store, externalServiceID int64, minSyncInterval time.Duration) (err error) {
	s.log().Debug("Syncing external service", "serviceID", externalServiceID)

	var svc *types.ExternalService
	ctx, save := s.observe(ctx, "Syncer.SyncExternalService", "")
	defer func() { save(svc, err) }()

	// We don't use tx here as the sourcing process below can be slow and we don't
	// want to hold a lock on the external_services table for too long.
	svc, err = s.Store.ExternalServiceStore.GetByID(ctx, externalServiceID)
	if err != nil {
		return errors.Wrap(err, "fetching external services")
	}

	// Unless our site config explicitly allows private code or the user has the
	// "AllowUserExternalServicePrivate" tag, user added external services should
	// only sync public code.
	allowed := func(*types.Repo) bool { return true }
	if svc.NamespaceUserID != 0 {
		if mode, err := database.UsersWith(tx).UserAllowedExternalServices(ctx, svc.NamespaceUserID); err != nil {
			return errors.Wrap(err, "checking if user can add private code")
		} else if mode != conf.ExternalServiceModeAll {
			allowed = func(r *types.Repo) bool { return !r.Private }
		}
		// If we are over our limit for user added repos we abort the sync
		totalAllowed := uint64(s.UserReposMaxPerSite)
		if totalAllowed == 0 {
			totalAllowed = uint64(conf.UserReposMaxPerSite())
		}
		userAdded, err := tx.CountUserAddedRepos(ctx)
		if err != nil {
			return errors.Wrap(err, "counting user added repos")
		}
		if userAdded >= totalAllowed {
			return errors.Errorf("reached maximum allowed user added repos: %d", userAdded)
		}
	}

	limit := s.UserReposMaxPerUser
	if limit == 0 {
		limit = conf.UserReposMaxPerUser()
	}

	src, err := s.Sourcer(svc)
	if err != nil {
		return err
	}

	results := make(chan SourceResult)
	go func() {
		src.ListRepos(ctx, results)
		close(results)
	}()

	modified := false
	seen := make(map[api.RepoID]struct{})
	errs := new(multierror.Error)
	count := 0

	// Insert or update repos as they are sourced. Keep track of what was seen
	// so we can remove anything else at the end.
	for res := range results {
		if err := res.Err; err != nil {
			multierror.Append(errs, errors.Wrapf(err, "fetching from code host %s", svc.DisplayName))
			if errcode.IsUnauthorized(errs) || errcode.IsForbidden(errs) || errcode.IsAccountSuspended(errs) {
				seen = map[api.RepoID]struct{}{} // Delete all external service repos of this external service
				break
			}
			continue
		}

		sourced := res.Repo
		if !allowed(sourced) {
			continue
		}

		if count++; count > limit {
			multierror.Append(errs, errors.Errorf("syncer: per user repo count has exceeded allowed limit: %d", limit))
			break
		}

		id, mod, err := s.sync(ctx, tx, svc, sourced)
		if err != nil {
			multierror.Append(errs, err)
			continue
		}

		seen[id] = struct{}{}
		modified = modified || mod
	}

	// Remove associations and any repos that are no longer associated with any external service.
	deleted := s.delete(ctx, tx, svc, seen)

	now := s.Now()
	shouldSyncSoon := modified || len(deleted) > 0 || errs.Len() > 0
	interval := calcSyncInterval(now, svc.LastSyncAt, minSyncInterval, shouldSyncSoon)

	s.log().Debug("Synced external service", "id", externalServiceID, "backoff duration", interval)
	svc.NextSyncAt = now.Add(interval)
	svc.LastSyncAt = now

	err = tx.ExternalServiceStore.Upsert(ctx, svc)
	if err != nil {
		multierror.Append(errors.Wrap(err, "upserting external service"))
	}

	return errs.ErrorOrNil()
}

func (s *Syncer) delete(ctx context.Context, tx *Store, svc *types.ExternalService, seen map[api.RepoID]struct{}) (_ []api.RepoID) {
	deleted, err := tx.DeleteExternalServiceReposNotIn(ctx, svc, seen)
	if err != nil {
		s.log().Error("some external service repos couldn't be deleted", "external-service-id", svc.ID, "err", err)
	}

	var d Diff
	for _, id := range deleted {
		d.Deleted = append(d.Deleted, &types.Repo{ID: id})
	}

	if s.Synced != nil && d.Len() > 0 {
		select {
		case <-ctx.Done():
		case s.Synced <- d:
		}
	}

	return deleted
}

// syncs a sourced repo of a given external service, returning the id of the repo, if it was modified, and an error if any.
func (s *Syncer) sync(ctx context.Context, tx *Store, svc *types.ExternalService, sourced *types.Repo) (_ api.RepoID, modified bool, err error) {
	if !tx.InTransaction() {
		tx, err = tx.Transact(ctx)
		if err != nil {
			return 0, false, errors.Wrap(err, "syncer: opening transaction")
		}
		defer func() { tx.Done(err) }()
	}

	stored, err := tx.RepoStore.List(ctx, database.ReposListOptions{
		Names:          []string{string(sourced.Name)},
		ExternalRepos:  []api.ExternalRepoSpec{sourced.ExternalRepo},
		IncludeBlocked: true,
		IncludeDeleted: true,
		UseOr:          true,
	})
	if err != nil {
		return 0, false, errors.Wrap(err, "syncer: getting repo from the database")
	}

	var d Diff

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
		if err = tx.RepoStore.Delete(ctx, conflicting.ID); err != nil {
			return 0, false, errors.Wrap(err, "syncer: failed to delete conflicting repo")
		}

		stored = types.Repos{existing}

		fallthrough
	case 1: // Existing repo, update.
		if !stored[0].Update(sourced) {
			d.Unmodified = append(d.Unmodified, stored[0])
			break
		}

		if err = tx.UpdateExternalServiceRepo(ctx, svc, stored[0]); err != nil {
			return 0, false, errors.Wrap(err, "syncer: failed to update external service repo")
		}

		d.Modified = append(d.Modified, stored[0])
	case 0: // New repo, create.
		if err = tx.CreateExternalServiceRepo(ctx, svc, sourced); err != nil {
			return 0, false, errors.Wrap(err, "syncer: failed to create external service repo")
		}

		stored = append(stored, sourced)
		d.Added = append(d.Added, sourced)
	default: // Impossible since we have two separate unique constraints on name and external repo spec
		panic("unreachable")
	}

	if s.Synced != nil && d.Len() > 0 {
		select {
		case <-ctx.Done():
		case s.Synced <- d:
		}
	}

	return stored[0].ID, len(d.Modified)+len(d.Added) > 0, nil
}

func calcSyncInterval(now time.Time, lastSync time.Time, minSyncInterval time.Duration, shouldSyncSoon bool) time.Duration {
	const maxSyncInterval = 8 * time.Hour

	// Special case, we've never synced
	if lastSync.IsZero() {
		return minSyncInterval
	}

	if shouldSyncSoon {
		return minSyncInterval
	}

	// No change, back off
	interval := now.Sub(lastSync) * 2
	if interval < minSyncInterval {
		return minSyncInterval
	}
	if interval > maxSyncInterval {
		return maxSyncInterval
	}
	return interval
}

// SyncRepo runs the syncer on a single repository.
func (s *Syncer) SyncRepo(ctx context.Context, sourced *types.Repo) (err error) {
	var svc *types.ExternalService
	ctx, save := s.observe(ctx, "Syncer.SyncRepo", string(sourced.Name))
	defer func() { save(svc, err) }()

	svcs, err := s.Store.ExternalServiceStore.List(ctx, database.ExternalServicesListOptions{
		Kinds:            []string{extsvc.TypeToKind(sourced.ExternalRepo.ServiceType)},
		OnlyCloudDefault: true,
		LimitOffset:      &database.LimitOffset{Limit: 1},
	})
	if err != nil {
		return errors.Wrap(err, "listing external services")
	}

	if len(svcs) != 1 {
		return errors.Wrapf(err, "cloud default external service of type %q not found", sourced.ExternalRepo.ServiceType)
	}

	svc = svcs[0]

	_, _, err = s.sync(ctx, s.Store, svc, sourced)
	return err
}

// initialUnmodifiedDiffFromStore creates a diff of all repos present in the
// store and sends it to s.Synced. This is used so that on startup the reader
// of s.Synced will receive a list of repos. In particular this is so that the
// git update scheduler can start working straight away on existing
// repositories.
func (s *Syncer) initialUnmodifiedDiffFromStore(ctx context.Context, store *Store) {
	if s.Synced == nil {
		return
	}

	stored, err := store.RepoStore.List(ctx, database.ReposListOptions{})
	if err != nil {
		s.log().Warn("initialUnmodifiedDiffFromStore store.ListRepos", "error", err)
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
	Modified   types.Repos
	Unmodified types.Repos
}

// Sort sorts all Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := range []types.Repos{
		d.Added,
		d.Deleted,
		d.Modified,
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
		d.Modified,
		d.Unmodified,
	} {
		all = append(all, rs...)
	}

	return all
}

func (d Diff) Len() int {
	return len(d.Deleted) + len(d.Modified) + len(d.Added) + len(d.Unmodified)
}

// NewDiff returns a diff from the given sourced and stored repos.
func NewDiff(sourced, stored []*types.Repo) (diff Diff) {
	return newDiff(nil, sourced, stored)
}

func newDiff(svc *types.ExternalService, sourced, stored []*types.Repo) (diff Diff) {
	// Sort sourced so we merge deterministically
	sort.Sort(types.Repos(sourced))

	byID := make(map[api.ExternalRepoSpec]*types.Repo, len(sourced))
	for _, r := range sourced {
		if old := byID[r.ExternalRepo]; old != nil {
			merge(old, r)
		} else {
			byID[r.ExternalRepo] = r
		}
	}

	// Ensure names are unique case-insensitively. We don't merge when finding
	// a conflict on name, we deterministically pick which sourced repo to
	// keep. Can't merge since they represent different repositories
	// (different external ID).
	byName := make(map[string]*types.Repo, len(byID))
	for _, r := range byID {
		k := strings.ToLower(string(r.Name))
		if old := byName[k]; old == nil {
			byName[k] = r
		} else {
			keep, discard := pick(r, old)
			byName[k] = keep
			delete(byID, discard.ExternalRepo)
		}
	}

	seenID := make(map[api.ExternalRepoSpec]bool, len(stored))

	for _, old := range stored {
		src := byID[old.ExternalRepo]

		// if the repo hasn't been found in the sourced repo list
		// we add it to the Deleted slice and, if the service is provided
		// we remove the service from its source map.
		if src == nil {
			if svc != nil {
				if _, ok := old.Sources[svc.URN()]; ok {
					old = old.Clone()
					delete(old.Sources, svc.URN())
				}
			}

			diff.Deleted = append(diff.Deleted, old)
		} else if old.Update(src) {
			diff.Modified = append(diff.Modified, old)
		} else {
			diff.Unmodified = append(diff.Unmodified, old)
		}

		seenID[old.ExternalRepo] = true
	}

	for _, r := range byID {
		if !seenID[r.ExternalRepo] {
			diff.Added = append(diff.Added, r)
		}
	}

	return diff
}

func merge(o, n *types.Repo) {
	for id, src := range o.Sources {
		n.Sources[id] = src
	}
	o.Update(n)
}

func (s *Syncer) sourced(ctx context.Context, svc *types.ExternalService, onSourced ...func(*types.Repo) error) ([]*types.Repo, error) {
	srcs, err := s.Sourcer(svc)
	if err != nil {
		return nil, err
	}

	return listAll(ctx, srcs, onSourced...)
}

func (s *Syncer) observe(ctx context.Context, family, title string) (context.Context, func(*types.ExternalService, error)) {
	began := s.Now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(svc *types.ExternalService, err error) {
		var owner string
		if svc == nil {
			owner = string(ownerUndefined)
		} else if svc.NamespaceUserID > 0 {
			owner = string(ownerUser)
		} else {
			owner = string(ownerSite)
		}

		syncStarted.WithLabelValues(family, owner).Inc()

		now := s.Now()
		took := s.Now().Sub(began).Seconds()

		lastSync.WithLabelValues(family).Set(float64(now.Unix()))

		success := err == nil
		syncDuration.WithLabelValues(strconv.FormatBool(success), family).Observe(took)

		if !success {
			tr.SetError(err)
			syncErrors.WithLabelValues(family, owner).Add(1)
		}

		tr.Finish()
	}
}
