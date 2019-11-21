package repos

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"gopkg.in/inconshreveable/log15.v2"
)

// NewPreSync takes in dependencies used by the Syncer and returns a function
// that can then be set on the Syncer as a PreSync.
type NewPreSync func(*sql.DB, Store, *httpcli.Factory) func(context.Context) error

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	Store   Store
	Sourcer Sourcer

	// PreSync is called before this Syncer's Sync method.
	PreSync func(context.Context) error

	// DisableStreaming if true will prevent the syncer from streaming in new
	// sourced repositories into the store.
	DisableStreaming bool

	// FailFullSync prevents Sync from running. This should only be true for
	// Sourcegraph.com
	FailFullSync bool

	// Synced is sent Repos that were synced by Sync (only if Synced is non-nil)
	Synced chan Repos

	// SubsetSynced is sent Repos that were synced by SubsetSync (only if SubsetSynced is non-nil)
	SubsetSynced chan Repos

	// Logger if non-nil is logged to.
	Logger log15.Logger

	// Now is time.Now. Can be set by tests to get deterministic output.
	Now func() time.Time

	// lastSyncErr contains the last error returned by the Sourcer in each
	// Sync. It's reset with each Sync and if the sync produced no error, it's
	// set to nil.
	lastSyncErr   error
	lastSyncErrMu sync.Mutex

	syncSignal signal
}

// Run runs the Sync at the specified interval.
func (s *Syncer) Run(pctx context.Context, interval time.Duration) error {
	for pctx.Err() == nil {
		ctx, cancel := contextWithSignalCancel(pctx, s.syncSignal.Watch())

		if err := s.Sync(ctx); err != nil && s.Logger != nil {
			s.Logger.Error("Syncer", "error", err)
		}

		sleep(ctx, interval)

		cancel()
	}

	return pctx.Err()
}

// contextWithSignalCancel will return a context which will be cancelled if
// signal fires. Callers need to call cancel when done.
func contextWithSignalCancel(ctx context.Context, signal <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)

	go func() {
		select {
		case <-ctx.Done():
		case <-signal:
			cancel()
		}
	}()

	return ctx, cancel
}

// sleep is a context aware time.Sleep
func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}

// TriggerSync will run Sync now. If a sync is currently running it is
// cancelled.
func (s *Syncer) TriggerSync() {
	s.syncSignal.Trigger()
}

// Sync synchronizes the repositories.
func (s *Syncer) Sync(ctx context.Context) (err error) {
	if s.PreSync != nil {
		if err := s.PreSync(ctx); err != nil && s.Logger != nil {
			s.Logger.Error("PreSync", "error", err)
		}
	}

	var diff Diff

	ctx, save := s.observe(ctx, "Syncer.Sync", "")
	defer save(&diff, &err)
	defer s.setOrResetLastSyncErr(&err)

	if s.FailFullSync {
		return errors.New("Syncer is not enabled")
	}

	var streamingInserter func(*Repo)
	if s.DisableStreaming {
		streamingInserter = func(*Repo) {} //noop
	} else {
		streamingInserter, err = s.makeNewRepoInserter(ctx)
		if err != nil {
			return errors.Wrap(err, "syncer.sync.streaming")
		}
	}

	var sourced Repos
	if sourced, err = s.sourced(ctx, streamingInserter); err != nil {
		return errors.Wrap(err, "syncer.sync.sourced")
	}

	store := s.Store
	if tr, ok := s.Store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return errors.Wrap(err, "syncer.sync.transact")
		}
		defer txs.Done(&err)
		store = txs
	}

	var stored Repos
	if stored, err = store.ListRepos(ctx, StoreListReposArgs{}); err != nil {
		return errors.Wrap(err, "syncer.sync.store.list-repos")
	}

	diff = NewDiff(sourced, stored)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return errors.Wrap(err, "syncer.sync.store.upsert-repos")
	}

	if s.Synced != nil {
		s.Synced <- diff.Repos()
	}

	return nil
}

// SyncSubset runs the syncer on a subset of the stored repositories. It will
// only sync the repositories with the same name or external service spec as
// sourcedSubset repositories.
func (s *Syncer) SyncSubset(ctx context.Context, sourcedSubset ...*Repo) (err error) {
	var diff Diff

	ctx, save := s.observe(ctx, "Syncer.SyncSubset", strings.Join(Repos(sourcedSubset).Names(), " "))
	defer save(&diff, &err)

	if len(sourcedSubset) == 0 {
		return nil
	}

	diff, err = s.syncSubset(ctx, false, sourcedSubset...)
	return err
}

// insertIfNew is a specialization of SyncSubset. It will insert sourcedRepo
// if there are no related repositories, otherwise does nothing.
func (s *Syncer) insertIfNew(ctx context.Context, sourcedRepo *Repo) (err error) {
	var diff Diff

	ctx, save := s.observe(ctx, "Syncer.InsertIfNew", sourcedRepo.Name)
	defer save(&diff, &err)

	diff, err = s.syncSubset(ctx, true, sourcedRepo)
	return err
}

func (s *Syncer) syncSubset(ctx context.Context, insertOnly bool, sourcedSubset ...*Repo) (diff Diff, err error) {
	if insertOnly && len(sourcedSubset) != 1 {
		return Diff{}, errors.Errorf("syncer.syncsubset.insertOnly can only handle one sourced repo, given %d repos", len(sourcedSubset))
	}

	store := s.Store
	if tr, ok := s.Store.(Transactor); ok {
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

	if insertOnly && len(storedSubset) > 0 {
		return Diff{}, nil
	}

	diff = NewDiff(sourcedSubset, storedSubset)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.syncsubset.store.upsert-repos")
	}

	if s.SubsetSynced != nil {
		s.SubsetSynced <- diff.Repos()
	}

	return diff, nil
}

func (s *Syncer) upserts(diff Diff) []*Repo {
	now := s.Now()
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

func (s *Syncer) sourced(ctx context.Context, observe ...func(*Repo)) ([]*Repo, error) {
	svcs, err := s.Store.ListExternalServices(ctx, StoreListExternalServicesArgs{})
	if err != nil {
		return nil, err
	}

	srcs, err := s.Sourcer(svcs...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, sourceTimeout)
	defer cancel()

	return listAll(ctx, srcs, observe...)
}

func (s *Syncer) makeNewRepoInserter(ctx context.Context) (func(*Repo), error) {
	// syncSubset requires querying the store for related repositories, and
	// will do nothing if `insertOnly` is set and there are any related repositories. Most
	// repositories will already have related repos, so to avoid that cost we
	// ask the store for all repositories and only do syncsubset if it might
	// be an insert.
	ids, err := s.storedExternalIDs(ctx)
	if err != nil {
		return nil, err
	}

	return func(r *Repo) {
		// We know this won't be an insert.
		if _, ok := ids[r.ExternalRepo]; ok {
			return
		}

		err := s.insertIfNew(ctx, r)
		if err != nil && s.Logger != nil {
			// Best-effort, final syncer will handle this repo if this failed.
			s.Logger.Warn("streaming insert failed", "external_id", r.ExternalRepo, "error", err)
		}
	}, nil
}

func (s *Syncer) storedExternalIDs(ctx context.Context) (map[api.ExternalRepoSpec]struct{}, error) {
	stored, err := s.Store.ListRepos(ctx, StoreListReposArgs{})
	if err != nil {
		return nil, errors.Wrap(err, "syncer.storedExternalIDs")
	}
	ids := make(map[api.ExternalRepoSpec]struct{}, len(stored))
	for _, r := range stored {
		ids[r.ExternalRepo] = struct{}{}
	}
	return ids, nil
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
	began := s.Now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(d *Diff, err *error) {
		now := s.Now()
		took := s.Now().Sub(began).Seconds()

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

				if len(repos) > 0 && s.Logger != nil {
					s.Logger.Debug(family, "diff."+state, repos.Names())
				}
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

type signal struct {
	once sync.Once
	c    chan struct{}
}

func (s *signal) init() {
	s.once.Do(func() {
		s.c = make(chan struct{}, 1)
	})
}

func (s *signal) Trigger() {
	s.init()
	select {
	case s.c <- struct{}{}:
	default:
	}
}

func (s *signal) Watch() <-chan struct{} {
	s.init()
	return s.c
}
