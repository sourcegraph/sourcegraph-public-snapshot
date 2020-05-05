package authz

import (
	"container/heap"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/inconshreveable/log15"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	edb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// PermsSyncer is a permissions syncing manager that is in charge of keeping
// permissions up-to-date for users and repositories.
//
// It is meant to be running in the background.
type PermsSyncer struct {
	// The priority queue to maintain the permissions syncing requests.
	queue *requestQueue
	// The database interface for any repos and external services operations.
	reposStore repos.Store
	// The database interface for any permissions operations.
	permsStore *edb.PermsStore
	// The mockable function to return the current time.
	clock func() time.Time
	// The time duration of how often to re-compute schedule for users and repositories.
	scheduleInterval time.Duration
	// The metrics that are exposed to Prometheus.
	metrics struct {
		noPerms      *prometheus.GaugeVec
		stalePerms   *prometheus.GaugeVec
		permsGap     *prometheus.GaugeVec
		syncErrors   *prometheus.CounterVec
		syncDuration *prometheus.HistogramVec
		queueSize    prometheus.Gauge
	}
}

// NewPermsSyncer returns a new permissions syncing manager.
func NewPermsSyncer(
	reposStore repos.Store,
	permsStore *edb.PermsStore,
	clock func() time.Time,
) *PermsSyncer {
	s := &PermsSyncer{
		queue:            newRequestQueue(),
		reposStore:       reposStore,
		permsStore:       permsStore,
		clock:            clock,
		scheduleInterval: time.Minute,
	}
	return s
}

// ScheduleUsers schedules new permissions syncing requests for given users
// in desired priority.
//
// This method implements the authz.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleUsers(ctx context.Context, priority Priority, userIDs ...int32) {
	users := make([]scheduledUser, len(userIDs))
	for i := range userIDs {
		users[i] = scheduledUser{
			priority: priority,
			userID:   userIDs[i],
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleUsers(ctx, users...)
}

func (s *PermsSyncer) scheduleUsers(ctx context.Context, users ...scheduledUser) {
	for _, u := range users {
		select {
		case <-ctx.Done():
			log15.Debug("PermsSyncer.scheduleUsers.canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   u.priority,
			Type:       requestTypeUser,
			ID:         u.userID,
			NextSyncAt: u.nextSyncAt,
			NoPerms:    u.noPerms,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "userID", u.userID, "updated", updated)
	}
}

// ScheduleRepos schedules new permissions syncing requests for given repositories
// in desired priority.
//
// This method implements the authz.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleRepos(ctx context.Context, priority Priority, repoIDs ...api.RepoID) {
	repos := make([]scheduledRepo, len(repoIDs))
	for i := range repoIDs {
		repos[i] = scheduledRepo{
			priority: priority,
			repoID:   repoIDs[i],
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleRepos(ctx, repos...)
}

func (s *PermsSyncer) scheduleRepos(ctx context.Context, repos ...scheduledRepo) {
	for _, r := range repos {
		select {
		case <-ctx.Done():
			log15.Debug("PermsSyncer.scheduleRepos.canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   r.priority,
			Type:       requestTypeRepo,
			ID:         int32(r.repoID),
			NextSyncAt: r.nextSyncAt,
			NoPerms:    r.noPerms,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "repoID", r.repoID, "updated", updated)
	}
}

// providers returns a list of authz.Provider configured in the external services.
// Keys are ServiceID, e.g. "https://gitlab.com/".
func (s *PermsSyncer) providers() map[string]authz.Provider {
	_, ps := authz.GetProviders()
	providers := make(map[string]authz.Provider, len(ps))
	for _, p := range ps {
		providers[p.ServiceID()] = p
	}
	return providers
}

// syncUserPerms processes permissions syncing request in user-centric way. When noPerms is true,
// the method will use partial results to update permissions tables when error occurs.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32, noPerms bool) (err error) {
	ctx, save := s.observe(ctx, "PermsSyncer.syncUserPerms", "")
	defer save(requestTypeUser, userID, &err)

	accts, err := s.permsStore.ListExternalAccounts(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "list external accounts")
	}

	var repoSpecs []api.ExternalRepoSpec
	for _, acct := range accts {
		provider := s.providers()[acct.ServiceID]
		if provider == nil {
			// We have no authz provider configured for this external account.
			continue
		}

		extIDs, err := provider.FetchUserPerms(ctx, acct)
		if err != nil {
			// Process partial results if this is an initial fetch.
			if !noPerms {
				return errors.Wrap(err, "fetch user permissions")
			}
			log15.Debug("PermsSyncer.syncUserPerms.proceedWithPartialResults", "userID", userID, "err", err)
		}

		for i := range extIDs {
			repoSpecs = append(repoSpecs, api.ExternalRepoSpec{
				ID:          string(extIDs[i]),
				ServiceType: provider.ServiceType(),
				ServiceID:   provider.ServiceID(),
			})
		}
	}

	var rs []*repos.Repo
	if len(repoSpecs) > 0 {
		// Get corresponding internal database IDs
		rs, err = s.reposStore.ListRepos(ctx, repos.StoreListReposArgs{
			ExternalRepos: repoSpecs,
			PrivateOnly:   true,
		})
		if err != nil {
			return errors.Wrap(err, "list external repositories")
		}
	}

	// Save permissions to database
	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
		IDs:    roaring.NewBitmap(),
	}
	for i := range rs {
		p.IDs.Add(uint32(rs[i].ID))
	}

	err = s.permsStore.SetUserPermissions(ctx, p)
	if err != nil {
		return errors.Wrap(err, "set user permissions")
	}

	log15.Info("PermsSyncer.syncUserPerms.synced", "userID", userID)
	return nil
}

// syncRepoPerms processes permissions syncing request in repository-centric way.
// It discards requests that are made for non-private repositories based on the
// value of "repo.private" column. When noPerms is true, the method will use partial
// results to update permissions tables when error occurs.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID api.RepoID, noPerms bool) (err error) {
	ctx, save := s.observe(ctx, "PermsSyncer.syncRepoPerms", "")
	defer save(requestTypeRepo, int32(repoID), &err)

	rs, err := s.reposStore.ListRepos(ctx, repos.StoreListReposArgs{
		IDs: []api.RepoID{repoID},
	})
	if err != nil {
		return errors.Wrap(err, "list repositories")
	} else if len(rs) == 0 {
		return nil
	}

	repo := rs[0]
	if !repo.Private {
		return nil
	}

	provider := s.providers()[repo.ExternalRepo.ServiceID]
	if provider == nil {
		// We have no authz provider configured for this repository.
		return nil
	}

	extAccountIDs, err := provider.FetchRepoPerms(ctx, &extsvc.Repository{
		URI:              repo.URI,
		ExternalRepoSpec: repo.ExternalRepo,
	})
	if err != nil {
		// Process partial results if this is an initial fetch.
		if !noPerms {
			return errors.Wrap(err, "fetch repository permissions")
		}
		log15.Debug("PermsSyncer.syncRepoPerms.proceedWithPartialResults", "repoID", repo.ID, "err", err)
	}

	pendingAccountIDsSet := make(map[string]struct{})
	var userIDs map[string]int32 // Account ID -> User ID
	if len(extAccountIDs) > 0 {
		accountIDs := make([]string, len(extAccountIDs))
		for i := range extAccountIDs {
			accountIDs[i] = string(extAccountIDs[i])
		}

		// Get corresponding internal database IDs
		userIDs, err = s.permsStore.GetUserIDsByExternalAccounts(ctx, &extsvc.Accounts{
			ServiceType: provider.ServiceType(),
			ServiceID:   provider.ServiceID(),
			AccountIDs:  accountIDs,
		})
		if err != nil {
			return errors.Wrap(err, "get user IDs by external accounts")
		}

		// Set up the set of all account IDs that need to be bound to permissions
		pendingAccountIDsSet = make(map[string]struct{}, len(accountIDs))
		for i := range accountIDs {
			pendingAccountIDsSet[accountIDs[i]] = struct{}{}
		}
	}

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: roaring.NewBitmap(),
	}

	for aid, uid := range userIDs {
		// Add existing user to permissions
		p.UserIDs.Add(uint32(uid))

		// Remove existing user from the set of pending users
		delete(pendingAccountIDsSet, aid)
	}

	pendingAccountIDs := make([]string, 0, len(pendingAccountIDsSet))
	for aid := range pendingAccountIDsSet {
		pendingAccountIDs = append(pendingAccountIDs, aid)
	}

	txs, err := s.permsStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer txs.Done(&err)

	accounts := &extsvc.Accounts{
		ServiceType: provider.ServiceType(),
		ServiceID:   provider.ServiceID(),
		AccountIDs:  pendingAccountIDs,
	}

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
		return errors.Wrap(err, "set repository pending permissions")
	}

	log15.Info("PermsSyncer.syncRepoPerms.synced", "repoID", repo.ID, "name", repo.Name)
	return nil
}

// syncPerms processes the permissions syncing request and remove the request from
// the queue once it is done (independent of success or failure).
func (s *PermsSyncer) syncPerms(ctx context.Context, request *syncRequest) error {
	defer s.queue.remove(request.Type, request.ID, true)

	var err error
	switch request.Type {
	case requestTypeUser:
		err = s.syncUserPerms(ctx, request.ID, request.NoPerms)
	case requestTypeRepo:
		err = s.syncRepoPerms(ctx, api.RepoID(request.ID), request.NoPerms)
	default:
		err = fmt.Errorf("unexpected request type: %v", request.Type)
	}

	return err
}

func (s *PermsSyncer) runSync(ctx context.Context) {
	log15.Debug("PermsSyncer.runSync.started")
	defer log15.Info("PermsSyncer.runSync.stopped")

	// To unblock the "select" on the next loop iteration if no enqueue happened in between.
	notifyDequeued := make(chan struct{}, 1)
	for {
		select {
		case <-notifyDequeued:
		case <-s.queue.notifyEnqueue:
		case <-ctx.Done():
			return
		}

		request := s.queue.acquireNext()
		if request == nil {
			// No waiting request is in the queue
			continue
		}

		// Check if it's the time to sync the request
		if wait := request.NextSyncAt.Sub(s.clock()); wait > 0 {
			s.queue.release(request.Type, request.ID)
			time.AfterFunc(wait, func() {
				notify(s.queue.notifyEnqueue)
			})

			log15.Debug("PermsSyncer.Run.waitForNextSync", "duration", wait)
			continue
		}

		notify(notifyDequeued)

		err := s.syncPerms(ctx, request)
		if err != nil {
			log15.Warn("Failed to sync permissions", "type", request.Type, "id", request.ID, "err", err)
			continue
		}
	}
}

// scheduleUsersWithNoPerms returns computed schedules for users who have no permissions
// found in database.
func (s *PermsSyncer) scheduleUsersWithNoPerms(ctx context.Context) ([]scheduledUser, error) {
	ids, err := s.permsStore.UserIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}
	s.metrics.noPerms.WithLabelValues("user").Set(float64(len(ids)))

	users := make([]scheduledUser, len(ids))
	for i, id := range ids {
		users[i] = scheduledUser{
			priority: PriorityLow,
			userID:   id,
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority.
			noPerms: true,
		}
	}
	return users, nil
}

// scheduleReposWithNoPerms returns computed schedules for private repositories that
// have no permissions found in database.
func (s *PermsSyncer) scheduleReposWithNoPerms(ctx context.Context) ([]scheduledRepo, error) {
	ids, err := s.permsStore.RepoIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}
	s.metrics.noPerms.WithLabelValues("repo").Set(float64(len(ids)))

	repos := make([]scheduledRepo, len(ids))
	for i, id := range ids {
		repos[i] = scheduledRepo{
			priority: PriorityLow,
			repoID:   id,
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority.
			noPerms: true,
		}
	}
	return repos, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have oldest
// permissions in database and capped results by the limit.
func (s *PermsSyncer) scheduleUsersWithOldestPerms(ctx context.Context, limit int) ([]scheduledUser, error) {
	results, err := s.permsStore.UserIDsWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, 0, len(results))
	for id, t := range results {
		users = append(users, scheduledUser{
			priority:   PriorityLow,
			userID:     id,
			nextSyncAt: t,
		})
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
func (s *PermsSyncer) scheduleReposWithOldestPerms(ctx context.Context, limit int) ([]scheduledRepo, error) {
	results, err := s.permsStore.ReposIDsWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, err
	}

	repos := make([]scheduledRepo, 0, len(results))
	for id, t := range results {
		repos = append(repos, scheduledRepo{
			priority:   PriorityLow,
			repoID:     id,
			nextSyncAt: t,
		})
	}
	return repos, nil
}

// schedule contains information for scheduling users and repositories.
type schedule struct {
	Users []scheduledUser
	Repos []scheduledRepo
}

// scheduledUser contains information for scheduling a user.
type scheduledUser struct {
	priority   Priority
	userID     int32
	nextSyncAt time.Time

	// Whether the user has no permissions when scheduled. Currently used to
	// accept partial results from authz provider in case of error.
	noPerms bool
}

// scheduledRepo contains for scheduling a repository.
type scheduledRepo struct {
	priority   Priority
	repoID     api.RepoID
	nextSyncAt time.Time

	// Whether the repository has no permissions when scheduled. Currently used
	// to accept partial results from authz provider in case of error.
	noPerms bool
}

// schedule computes schedule four lists in the following order:
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsSyncer) schedule(ctx context.Context) (*schedule, error) {
	schedule := new(schedule)

	users, err := s.scheduleUsersWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule users with no permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err := s.scheduleReposWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule repositories with no permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	// TODO(jchen): Predict a limit taking account into:
	//   1. Based on total repos and users that make sense to finish syncing before
	//      next schedule call, so we don't waste database bandwidth.
	//   2. How we're doing in terms of rate limiting.
	// Formula (in worse case scenario, at the pace of 1 req/s):
	//   initial limit  = <predicted from the previous step>
	//	 consumed by users = <initial limit> / (<total repos> / <page size>)
	//   consumed by repos = (<initial limit> - <consumed by users>) / (<total users> / <page size>)
	// Hard coded both to 10 for now.
	const limit = 10

	// TODO(jchen): Use better heuristics for setting NextSyncAt, the initial version
	// just uses the value of LastUpdatedAt get from the perms tables.

	users, err = s.scheduleUsersWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, errors.Wrap(err, "load users with oldest permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err = s.scheduleReposWithOldestPerms(ctx, limit)
	if err != nil {
		return nil, errors.Wrap(err, "scan repositories with oldest permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	return schedule, nil
}

func (s *PermsSyncer) runSchedule(ctx context.Context) {
	log15.Debug("PermsSyncer.runSchedule.started")
	defer log15.Info("PermsSyncer.runSchedule.stopped")

	ticker := time.NewTicker(s.scheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		if !globals.PermissionsBackgroundSync().Enabled {
			continue
		}

		schedule, err := s.schedule(ctx)
		if err != nil {
			log15.Error("Failed to compute schedule", "err", err)
			continue
		}

		s.scheduleUsers(ctx, schedule.Users...)
		s.scheduleRepos(ctx, schedule.Repos...)
	}
}

// DebugDump returns the state of the permissions syncer for debugging.
func (s *PermsSyncer) DebugDump() interface{} {
	type requestInfo struct {
		Meta     *requestMeta
		Acquired bool
	}
	data := struct {
		Name  string
		Size  int
		Queue []*requestInfo
	}{
		Name: "permissions",
	}

	queue := requestQueue{
		heap: make([]*syncRequest, len(s.queue.heap)),
	}

	s.queue.mu.RLock()
	defer s.queue.mu.RUnlock()

	for i, request := range s.queue.heap {
		// Copy the syncRequest as a value so that poping off the heap here won't
		// update the index value of the real heap, and we don't do a racy read on
		// the repo pointer which may change concurrently in the real heap.
		requestCopy := *request
		queue.heap[i] = &requestCopy
	}

	for len(queue.heap) > 0 {
		// Copy values of the syncRequest so that the requestMeta pointer
		// won't change concurrently after we release the lock.
		request := heap.Pop(&queue).(*syncRequest)
		data.Queue = append(data.Queue, &requestInfo{
			Meta: &requestMeta{
				Priority:   request.Priority,
				Type:       request.Type,
				ID:         request.ID,
				NextSyncAt: request.NextSyncAt,
			},
			Acquired: request.acquired,
		})
	}
	data.Size = len(data.Queue)

	return &data
}

func (s *PermsSyncer) observe(ctx context.Context, family, title string) (context.Context, func(requestType, int32, *error)) {
	began := s.clock()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(typ requestType, id int32, err *error) {
		defer tr.Finish()
		tr.LogFields(otlog.Int32("id", id))

		var typLabel string
		switch typ {
		case requestTypeRepo:
			typLabel = "repo"
		case requestTypeUser:
			typLabel = "user"
		default:
			tr.SetError(fmt.Errorf("unexpected request type: %v", typ))
			return
		}

		success := err == nil || *err == nil
		s.metrics.syncDuration.WithLabelValues(typLabel, strconv.FormatBool(success)).Observe(time.Since(began).Seconds())

		if !success {
			tr.SetError(*err)
			s.metrics.syncErrors.WithLabelValues(typLabel).Add(1)
		}
	}
}

// registerMetrics registers exposed metrics with Prometheus default register.
func (s *PermsSyncer) registerMetrics() {
	s.metrics.noPerms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_no_perms",
		Help: "The number of records that do not have any permissions",
	}, []string{"type"})
	s.metrics.stalePerms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_stale_perms",
		Help: "The number of records that have stale permissions",
	}, []string{"type"})
	s.metrics.permsGap = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_perms_gap_seconds",
		Help: "The time gap between least and most up to date permissions",
	}, []string{"type"})
	s.metrics.syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_repoupdater_perms_syncer_sync_duration_seconds",
		Help:    "Time spent on syncing permissions",
		Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"type", "success"})
	s.metrics.syncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_perms_syncer_sync_errors_total",
		Help: "Total number of permissions sync errors",
	}, []string{"type"})
	s.metrics.queueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_queue_size",
		Help: "The size of the sync request queue",
	})
}

// collectMetrics periodically collecting metrics values from both database and memory objects.
func (s *PermsSyncer) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		m, err := s.permsStore.Metrics(ctx, 3*24*time.Hour)
		if err != nil {
			log15.Error("Failed to get metrics from database", "err", err)
			continue
		}

		s.metrics.stalePerms.WithLabelValues("user").Set(float64(m.UsersWithStalePerms))
		s.metrics.permsGap.WithLabelValues("user").Set(m.UsersPermsGapSeconds)
		s.metrics.stalePerms.WithLabelValues("repo").Set(float64(m.ReposWithStalePerms))
		s.metrics.permsGap.WithLabelValues("repo").Set(m.ReposPermsGapSeconds)

		s.queue.mu.RLock()
		s.metrics.queueSize.Set(float64(s.queue.Len()))
		s.queue.mu.RUnlock()
	}
}

// Run kicks off the permissions syncing process, this method is blocking and
// should be called as a goroutine.
func (s *PermsSyncer) Run(ctx context.Context) {
	s.registerMetrics()
	go s.runSync(ctx)
	go s.runSchedule(ctx)
	go s.collectMetrics(ctx)

	<-ctx.Done()
}
