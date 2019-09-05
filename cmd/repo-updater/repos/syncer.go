package repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"gopkg.in/inconshreveable/log15.v2"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	// FailFullSync prevents Sync from running. This should only be true for
	// Sourcegraph.com
	FailFullSync bool

	// lastSyncErr contains the last error returned by the Sourcer in each
	// Sync. It's reset with each Sync and if the sync produced no error, it's
	// set to nil.
	lastSyncErr   error
	lastSyncErrMu sync.Mutex

	store   Store
	sourcer Sourcer
	diffs   chan Diff
	now     func() time.Time

	syncSignal chan struct{}
}

// NewSyncer returns a new Syncer that syncs stored repos with
// the repos yielded by the configured sources, retrieved by the given sourcer.
// Each completed sync results in a diff that is sent to the given diffs channel.
func NewSyncer(
	store Store,
	sourcer Sourcer,
	diffs chan Diff,
	now func() time.Time,
) *Syncer {
	return &Syncer{
		store:      store,
		sourcer:    sourcer,
		diffs:      diffs,
		now:        now,
		syncSignal: make(chan struct{}, 1),
	}
}

// Run runs the Sync at the specified interval.
func (s *Syncer) Run(ctx context.Context, interval time.Duration, useStreaming bool) error {
	for ctx.Err() == nil {
		if useStreaming {
			if err := s.StreamingSync(ctx); err != nil {
				log15.Error("Syncer", "error", err)
			}
		} else {
			if _, err := s.Sync(ctx); err != nil {
				log15.Error("Syncer", "error", err)
			}
		}

		select {
		case <-time.After(interval):
		case <-s.syncSignal:
		}
	}

	return ctx.Err()
}

// TriggerSync will run Sync as soon as the current Sync has finished running
// or if no Sync is running.
func (s *Syncer) TriggerSync() {
	select {
	case s.syncSignal <- struct{}{}:
	default:
	}
}

// Sync synchronizes the repositories.
func (s *Syncer) Sync(ctx context.Context) (diff Diff, err error) {
	ctx, save := s.observe(ctx, "Syncer.Sync", "")
	defer save(&diff, &err)
	defer s.setOrResetLastSyncErr(&err)

	if s.FailFullSync {
		return Diff{}, errors.New("Syncer is not enabled")
	}

	var sourced Repos
	if sourced, err = s.sourced(ctx); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.sourced")
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return Diff{}, errors.Wrap(err, "syncer.sync.transact")
		}
		defer txs.Done(&err)
		store = txs
	}

	var stored Repos
	if stored, err = store.ListRepos(ctx, StoreListReposArgs{}); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.list-repos")
	}

	diff = NewDiff(sourced, stored)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.upsert-repos")
	}

	if s.diffs != nil {
		s.diffs <- diff
	}

	return diff, nil
}

type syncRunState struct {
	byID     map[api.ExternalRepoSpec]*Repo
	byName   map[string]*Repo
	seenID   map[api.ExternalRepoSpec]bool
	seenName map[string]bool
}

// StreamingSync streams repositories when syncing
func (s *Syncer) StreamingSync(ctx context.Context) (err error) {
	ctx, save := s.observe(ctx, "Syncer.StreamingSync", "")
	defer save(&Diff{}, &err)
	defer s.setOrResetLastSyncErr(&err)

	if s.FailFullSync {
		return errors.New("Syncer is not enabled")
	}

	sourcedCtx, cancel := context.WithTimeout(ctx, sourceTimeout)
	defer cancel()

	sourced, err := s.asyncSourced(sourcedCtx)
	if err != nil {
		return errors.Wrap(err, "syncer.streaming-sync.async-sourced")
	}

	state := &syncRunState{
		byID:     make(map[api.ExternalRepoSpec]*Repo),
		byName:   make(map[string]*Repo),
		seenID:   make(map[api.ExternalRepoSpec]bool),
		seenName: make(map[string]bool),
	}

	var errs *multierror.Error
	for result := range sourced {
		if result.Err != nil {
			for _, extSvc := range result.Source.ExternalServices() {
				errs = multierror.Append(errs, &SourceError{Err: result.Err, ExtSvc: extSvc})
			}
			continue
		}
		err := s.syncSourcedRepo(ctx, state, result.Repo)
		if err != nil {
			return err
		}
	}
	if err = errs.ErrorOrNil(); err != nil {
		return errors.Wrap(err, "syncer.streaming-sync.sourcing")
	}

	// In `byID` are now all repositories that we want to keep in the database
	// We use `byID` for that instead of a separate slice, because when we come
	// across repositories with the same, we choose a winner and discard the repo
	// by deleting it from `byID`.
	// We _could_ delete the discarded repo with a `DELETE` query directly,
	// but this way, we only do a single `DELETE` at the end.
	reposToKeep := make([]uint32, 0, len(state.byID))
	for _, r := range state.byID {
		if r.ID == 0 {
			panic("repo has 0 id!")
		}
		reposToKeep = append(reposToKeep, r.ID)
	}

	if err = s.store.DeleteReposExcept(ctx, reposToKeep...); err != nil {
		return errors.Wrap(err, "syncer.streaming-sync.store.delete-repos-except")
	}

	return err
}

func (s *Syncer) syncSourcedRepo(ctx context.Context, state *syncRunState, r *Repo) (err error) {
	if !r.ExternalRepo.IsSet() {
		panic(fmt.Errorf("%s has no valid external repo spec: %s", r.Name, r.ExternalRepo))
	}

	var (
		txs              TxStore
		closeTransaction = false
		store            = s.store
	)
	if tr, ok := s.store.(Transactor); ok {
		if txs, err = tr.Transact(ctx); err != nil {
			return errors.Wrap(err, "syncer.syncsubset.transact")
		}
		closeTransaction = true
		store = txs
	}

	merged := false
	if old := state.byID[r.ExternalRepo]; old != nil {
		// We use `Less` here in order to deterministically merge repos.
		// e.g. with two repos that have the same name, one lower- one
		// uppercase, the same one would always win
		if r.Less(old) {
			merge(r, old)
		} else {
			merge(old, r)
		}
		merged = true
	}

	// Ensure names are unique case-insensitively. We don't merge when finding
	// a conflict on name, we deterministically pick which sourced repo to
	// keep. Can't merge since they represent different repositories
	// (different external ID).
	k := strings.ToLower(r.Name)
	if old := state.byName[k]; old == nil {
		state.byName[k] = r
	} else if !merged {
		// Only discard repo if we didn't previously merge `r` with another
		// one.

		// If we previously merged, there _is_ going to be a naming
		// confict, in which case we don't want to discard the merged
		// repo.

		keep, discard := pick(r, old)
		state.byName[k] = keep
		delete(state.byID, discard.ExternalRepo)
	}

	storedSubset, err := s.listMatchesForSourcedRepo(ctx, store, r)
	if err != nil {
		return errors.Wrap(err, "syncer.streaming-sync.store.list-repos")
	}

	diff := Diff{}
	for _, old := range storedSubset {
		var src *Repo
		if r.ExternalRepo == old.ExternalRepo {
			src = r
		}

		// We do not want a stored repository without an externalrepo to be set.
		//
		// We are unsure if customer repositories can have ExternalRepo unset. We
		// know it can be unset for Sourcegraph.com. As such, we want to fallback
		// to associating stored repositories by name with the sourced
		// repositories.
		// But only if we didn't previously associate the stored repo with
		// another repo that we sourced.
		if old.ExternalRepo.ID == "" && !state.seenName[old.Name] {
			src = state.byName[strings.ToLower(old.Name)]
		}

		if src == nil {
			diff.Deleted = append(diff.Deleted, old)
			continue
		}

		if old.Update(src) {
			diff.Modified = append(diff.Modified, old)
		} else {
			diff.Unmodified = append(diff.Unmodified, old)
		}
		state.seenID[old.ExternalRepo] = true
		state.seenName[old.Name] = true
	}

	if !state.seenID[r.ExternalRepo] {
		diff.Added = append(diff.Added, r)
	}

	if err = store.UpsertRepos(ctx, s.upserts(diff)...); err != nil {
		return errors.Wrap(err, "syncer.streaming-sync.store.upsert-repos")
	}

	// Added, Unmodified, Modified are all repos that are in the database
	// now.
	// We keep track of them so we can `merge` them with the repos that
	// will come over `results`.
	// That allows us to
	// (1) avoid duplicates (i.e. same repo from the same external service)
	// (2) merge sources of same repo (i.e. same repo from different same external services)
	for _, r := range diff.ReposExceptDeleted() {
		state.byID[r.ExternalRepo] = r
		state.byName[strings.ToLower(r.Name)] = r
	}

	if closeTransaction {
		txs.Done(&err)
	}

	return nil
}

func (s *Syncer) listMatchesForSourcedRepo(ctx context.Context, store Store, r *Repo) (Repos, error) {
	var storedSubset Repos
	args := StoreListReposArgs{
		Names:         []string{r.Name},
		ExternalRepos: []api.ExternalRepoSpec{r.ExternalRepo},
		UseOr:         true,
	}

	storedSubset, err := store.ListRepos(ctx, args)
	if err != nil {
		return storedSubset, err
	}

	sort.Stable(byExternalRepoSpecSet(storedSubset))

	return storedSubset, nil
}

// SyncSubset runs the syncer on a subset of the stored repositories. It will
// only sync the repositories with the same name or external service spec as
// sourcedSubset repositories.
func (s *Syncer) SyncSubset(ctx context.Context, sourcedSubset ...*Repo) (diff Diff, err error) {
	ctx, save := s.observe(ctx, "Syncer.SyncSubset", strings.Join(Repos(sourcedSubset).Names(), " "))
	defer save(&diff, &err)

	if len(sourcedSubset) == 0 {
		return Diff{}, nil
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return Diff{}, errors.Wrap(err, "syncer.syncsubset.transact")
		}
		defer txs.Done(&err)
		store = txs
	}

	var storedSubset Repos
	args := StoreListReposArgs{
		Names:         Repos(sourcedSubset).Names(),
		ExternalRepos: Repos(sourcedSubset).ExternalRepos(),
		UseOr:         true,
	}
	if storedSubset, err = store.ListRepos(ctx, args); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.syncsubset.store.list-repos")
	}

	diff = NewDiff(sourcedSubset, storedSubset)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.syncsubset.store.upsert-repos")
	}

	if s.diffs != nil {
		s.diffs <- diff
	}

	return diff, nil
}

func (s *Syncer) upserts(diff Diff) []*Repo {
	now := s.now()
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	for _, repo := range diff.Deleted {
		repo.UpdatedAt, repo.DeletedAt = now, now
		repo.Sources = map[string]*SourceInfo{}
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Modified {
		repo.UpdatedAt, repo.DeletedAt = now, time.Time{}
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Added {
		repo.CreatedAt, repo.UpdatedAt, repo.DeletedAt = now, now, time.Time{}
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	return upserts
}

// A Diff of two sets of Diffables.
type Diff struct {
	Added      Repos
	Deleted    Repos
	Modified   Repos
	Unmodified Repos
}

// Sort sorts all Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := range []Repos{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		sort.Sort(ds)
	}
}

// Repos returns all repos in the Diff.
func (d Diff) Repos() Repos {
	all := make(Repos, 0, len(d.Added)+
		len(d.Deleted)+
		len(d.Modified)+
		len(d.Unmodified))

	for _, rs := range []Repos{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		all = append(all, rs...)
	}

	return all
}

// ReposExceptDeleted returns all repos in the Diff except the repos in Deleted.
func (d Diff) ReposExceptDeleted() Repos {
	all := make(Repos, 0, len(d.Added)+
		len(d.Modified)+
		len(d.Unmodified))

	for _, rs := range []Repos{d.Added, d.Modified, d.Unmodified} {
		all = append(all, rs...)
	}

	return all
}

// NewDiff returns a diff from the given sourced and stored repos.
func NewDiff(sourced, stored []*Repo) (diff Diff) {
	// Sort sourced so we merge determinstically
	sort.Sort(Repos(sourced))

	byID := make(map[api.ExternalRepoSpec]*Repo, len(sourced))
	for _, r := range sourced {
		if !r.ExternalRepo.IsSet() {
			panic(fmt.Errorf("%s has no valid external repo spec: %s", r.Name, r.ExternalRepo))
		} else if old := byID[r.ExternalRepo]; old != nil {
			merge(old, r)
		} else {
			byID[r.ExternalRepo] = r
		}
	}

	// Ensure names are unique case-insensitively. We don't merge when finding
	// a conflict on name, we deterministically pick which sourced repo to
	// keep. Can't merge since they represent different repositories
	// (different external ID).
	byName := make(map[string]*Repo, len(byID))
	for _, r := range byID {
		k := strings.ToLower(r.Name)
		if old := byName[k]; old == nil {
			byName[k] = r
		} else {
			keep, discard := pick(r, old)
			byName[k] = keep
			delete(byID, discard.ExternalRepo)
		}
	}

	seenID := make(map[api.ExternalRepoSpec]bool, len(stored))
	seenName := make(map[string]bool, len(stored))

	// We are unsure if customer repositories can have ExternalRepo unset. We
	// know it can be unset for Sourcegraph.com. As such, we want to fallback
	// to associating stored repositories by name with the sourced
	// repositories.
	//
	// We do not want a stored repository without an externalrepo to be set
	sort.Stable(byExternalRepoSpecSet(stored))

	for _, old := range stored {
		src := byID[old.ExternalRepo]
		if src == nil && old.ExternalRepo.ID == "" && !seenName[old.Name] {
			src = byName[strings.ToLower(old.Name)]
		}

		if src == nil {
			diff.Deleted = append(diff.Deleted, old)
		} else if old.Update(src) {
			diff.Modified = append(diff.Modified, old)
		} else {
			diff.Unmodified = append(diff.Unmodified, old)
		}

		seenID[old.ExternalRepo] = true
		seenName[old.Name] = true
	}

	for _, r := range byID {
		if !seenID[r.ExternalRepo] {
			diff.Added = append(diff.Added, r)
		}
	}

	return diff
}

func merge(o, n *Repo) {
	for id, src := range o.Sources {
		n.Sources[id] = src
	}
	o.Update(n)
}

func (s *Syncer) sourced(ctx context.Context) ([]*Repo, error) {
	svcs, err := s.store.ListExternalServices(ctx, StoreListExternalServicesArgs{})
	if err != nil {
		return nil, err
	}

	srcs, err := s.sourcer(svcs...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, sourceTimeout)
	defer cancel()

	return listAll(ctx, srcs)
}

func (s *Syncer) asyncSourced(ctx context.Context) (chan SourceResult, error) {
	svcs, err := s.store.ListExternalServices(ctx, StoreListExternalServicesArgs{})
	if err != nil {
		return nil, err
	}

	srcs, err := s.sourcer(svcs...)
	if err != nil {
		return nil, err
	}

	results := make(chan SourceResult)

	go func() {
		srcs.ListRepos(ctx, results)
		close(results)
	}()

	return results, nil
}

func (s *Syncer) setOrResetLastSyncErr(perr *error) {
	var err error
	if perr != nil {
		err = *perr
	}

	s.lastSyncErrMu.Lock()
	s.lastSyncErr = err
	s.lastSyncErrMu.Unlock()
}

// LastSyncError returns the error that was produced in the last Sync run. If
// no error was produced, this returns nil.
func (s *Syncer) LastSyncError() error {
	s.lastSyncErrMu.Lock()
	defer s.lastSyncErrMu.Unlock()

	return s.lastSyncErr
}

func (s *Syncer) observe(ctx context.Context, family, title string) (context.Context, func(*Diff, *error)) {
	began := s.now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(d *Diff, err *error) {
		now := s.now()
		took := s.now().Sub(began).Seconds()

		fields := make([]otlog.Field, 0, 7)
		for state, repos := range map[string]Repos{
			"added":      d.Added,
			"modified":   d.Modified,
			"deleted":    d.Deleted,
			"unmodified": d.Unmodified,
		} {
			fields = append(fields, otlog.Int(state+".count", len(repos)))
			if state != "unmodified" {
				fields = append(fields,
					otlog.Object(state+".repos", repos.Names()))
			}
			syncedTotal.WithLabelValues(state).Add(float64(len(repos)))
		}

		tr.LogFields(fields...)

		lastSync.WithLabelValues().Set(float64(now.Unix()))

		success := err == nil || *err == nil
		syncDuration.WithLabelValues(strconv.FormatBool(success)).Observe(took)

		if !success {
			tr.SetError(*err)
			syncErrors.WithLabelValues().Add(1)
		}

		tr.Finish()
	}
}

type byExternalRepoSpecSet []*Repo

func (rs byExternalRepoSpecSet) Len() int      { return len(rs) }
func (rs byExternalRepoSpecSet) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs byExternalRepoSpecSet) Less(i, j int) bool {
	iSet := rs[i].ExternalRepo.IsSet()
	jSet := rs[j].ExternalRepo.IsSet()
	if iSet == jSet {
		return false
	}
	return iSet
}
