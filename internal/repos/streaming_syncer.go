package repos

import (
	"context"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// StreamingSyncer wraps a Syncer to replace its batch SyncExternalService and SyncRepo methods
// with streaming implementations.
type StreamingSyncer struct{ *Syncer }

// SyncRepo syncs a single repository with the first cloud default external service found
// for its type.
func (s *StreamingSyncer) SyncRepo(ctx context.Context, sourced *types.Repo) (err error) {
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
	_, err = s.sync(ctx, s.Store, svc, sourced)
	return err
}

// SyncExternalService syncs repos using the supplied external service in a streaming fashion, rather than batch.
// This allows very large sync jobs (i.e. that source potentially millions of repos) to incrementally persist changes.
// Deletes of repositories that were not sourced are done at the end.
func (s *StreamingSyncer) SyncExternalService(ctx context.Context, tx *Store, externalServiceID int64, minSyncInterval time.Duration) (err error) {
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

		diff, err := s.sync(ctx, tx, svc, sourced)
		if err != nil {
			multierror.Append(errs, err)
			continue
		}

		for _, r := range diff.Repos() {
			seen[r.ID] = struct{}{}
		}

		modified = modified || len(diff.Modified)+len(diff.Added) > 0
	}

	// Remove associations and any repos that are no longer associated with any external service.
	deleted, err := s.delete(ctx, tx, svc, seen)
	if err != nil {
		multierror.Append(err, errors.Wrap(err, "some repos couldn't be deleted"))
	}

	now := s.Now()
	modified = modified || deleted > 0
	interval := s.calcSyncInterval(now, svc.LastSyncAt, minSyncInterval, modified, errs.ErrorOrNil())

	s.log().Debug("Synced external service", "id", externalServiceID, "backoff duration", interval)
	svc.NextSyncAt = now.Add(interval)
	svc.LastSyncAt = now

	err = tx.ExternalServiceStore.Upsert(ctx, svc)
	if err != nil {
		multierror.Append(errors.Wrap(err, "upserting external service"))
	}

	return errs.ErrorOrNil()
}

// syncs a sourced repo of a given external service, returning a diff with a single repo.
func (s *StreamingSyncer) sync(ctx context.Context, tx *Store, svc *types.ExternalService, sourced *types.Repo) (d Diff, err error) {
	if !tx.InTransaction() {
		tx, err = tx.Transact(ctx)
		if err != nil {
			return Diff{}, errors.Wrap(err, "syncer: opening transaction")
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
		if err = tx.RepoStore.Delete(ctx, conflicting.ID); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to delete conflicting repo")
		}

		stored = types.Repos{existing}

		fallthrough
	case 1: // Existing repo, update.
		if !stored[0].Update(sourced) {
			d.Unmodified = append(d.Unmodified, stored[0])
			break
		}

		if err = tx.UpdateExternalServiceRepo(ctx, svc, stored[0]); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to update external service repo")
		}

		d.Modified = append(d.Modified, stored[0])
	case 0: // New repo, create.
		if err = tx.CreateExternalServiceRepo(ctx, svc, sourced); err != nil {
			return Diff{}, errors.Wrap(err, "syncer: failed to create external service repo")
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

	return d, nil
}

func (s *StreamingSyncer) delete(ctx context.Context, tx *Store, svc *types.ExternalService, seen map[api.RepoID]struct{}) (int, error) {
	// We do deletion in a best effort manner, returning any errors for individual repos that failed to be deleted.
	deleted, err := tx.DeleteExternalServiceReposNotIn(ctx, svc, seen)

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

	return len(deleted), err
}

var discardLogger = func() log15.Logger {
	l := log15.New()
	l.SetHandler(log15.DiscardHandler())
	return l
}()

func (s *StreamingSyncer) log() log15.Logger {
	if s.Logger == nil {
		return discardLogger
	}
	return s.Logger
}

func (StreamingSyncer) calcSyncInterval(now time.Time, lastSync time.Time, minSyncInterval time.Duration, modified bool, err error) time.Duration {
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

func (s *StreamingSyncer) observe(ctx context.Context, family, title string) (context.Context, func(*types.ExternalService, error)) {
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
