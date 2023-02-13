package authz

import (
	"container/heap"
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

// PermissionSyncingDisabled returns true if the background permissions syncing is not enabled.
// It is not enabled if:
//   - Permissions user mapping (aka explicit permissions API) is enabled
//   - Not purchased with the current license
//   - `disableAutoCodeHostSyncs` site setting is set to true
func PermissionSyncingDisabled() bool {
	return globals.PermissionsUserMapping().Enabled ||
		licensing.Check(licensing.FeatureACLs) != nil ||
		conf.Get().DisableAutoCodeHostSyncs
}

var scheduleReposCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "src_repoupdater_perms_syncer_schedule_repos_total",
	Help: "Counts number of repos for which permissions syncing request has been scheduled.",
})

// PermsSyncer is a permissions syncing manager that is in charge of keeping
// permissions up-to-date for users and repositories.
//
// It is meant to be running in the background.
type PermsSyncer struct {
	// The priority queue to maintain the permissions syncing requests.
	queue *requestQueue
	// The logger to use when logging messages and errors
	logger log.Logger
	// The generic database handle.
	db database.DB
	// The database interface for any repos and external services operations.
	reposStore repos.Store
	// The mockable function to return the current time.
	clock func() time.Time
	// The rate limit registry for code hosts.
	rateLimiterRegistry *ratelimit.Registry
	// The time duration of how often to re-compute schedule for users and repositories.
	scheduleInterval time.Duration

	// The lock to ensure there is no concurrent updates (i.e. only one) to the
	// permissions tables. The mutex is used to prevent any potential deadlock that
	// could be caused by concurrent database updates, and it is a simpler and more
	// intuitive approach than trying to solve deadlocks caused by how permissions
	// are stored in the database at the time of writing. In a production setup with
	// thousands of repositories and users, this approach is more effective as the
	// biggest contributor and bottleneck of background permissions syncing slowness
	// is the time spent on API calls (usually minutes) vs a database update
	// operation (usually <1s).
	permsUpdateLock sync.Mutex
	// The database interface for any permissions operations.
	permsStore edb.PermsStore

	// recordsStore tracks results of recent permissions sync jobs.
	recordsStore *syncjobs.RecordsStore
}

// NewPermsSyncer returns a new permissions syncing manager.
func NewPermsSyncer(
	logger log.Logger,
	db database.DB,
	reposStore repos.Store,
	permsStore edb.PermsStore,
	clock func() time.Time,
	rateLimiterRegistry *ratelimit.Registry,
) *PermsSyncer {
	return &PermsSyncer{
		queue:               newRequestQueue(),
		logger:              logger,
		db:                  db,
		reposStore:          reposStore,
		permsStore:          permsStore,
		clock:               clock,
		rateLimiterRegistry: rateLimiterRegistry,
		scheduleInterval:    scheduleInterval(),
		recordsStore:        syncjobs.NewRecordsStore(logger.Scoped("records", "sync jobs records store"), conf.DefaultClient()),
	}
}

// ScheduleUsers schedules new permissions syncing requests for given users.
// By design, all schedules triggered by user actions are in high priority.
//
// This method implements the repoupdater.Server.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleUsers(ctx context.Context, opts authz.FetchPermsOptions, userIDs ...int32) {
	if len(userIDs) == 0 {
		return
	} else if s.isDisabled() {
		s.logger.Debug("PermsSyncer.ScheduleUsers.disabled", log.Int("userIDs", len(userIDs)))
		return
	}

	users := make([]scheduledUser, len(userIDs))
	for i := range userIDs {
		users[i] = scheduledUser{
			priority: priorityHigh,
			userID:   userIDs[i],
			options:  opts,
			// NOTE: Have processAfter with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleUsers(ctx, users...)
	metricsItemsSyncScheduled.WithLabelValues("manualUsersTrigger", "high").Set(float64(len(userIDs)))
	s.collectQueueSize()
}

func (s *PermsSyncer) scheduleUsers(ctx context.Context, users ...scheduledUser) {
	logger := s.logger.Scoped("scheduledUsers", "routine for adding users to a queue for sync")
	for _, u := range users {
		select {
		case <-ctx.Done():
			s.logger.Debug("canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   u.priority,
			Type:       requestTypeUser,
			ID:         u.userID,
			Options:    u.options,
			NextSyncAt: u.processAfter,
			NoPerms:    u.noPerms,
		})
		logger.Debug("queue.enqueued", log.Int32("userID", u.userID), log.Bool("updated", updated))
	}
}

// ScheduleRepos schedules new permissions syncing requests for given repositories.
// By design, all schedules triggered by user actions are in high priority.
//
// This method implements the repoupdater.Server.PermsSyncer in the OSS namespace.
func (s *PermsSyncer) ScheduleRepos(ctx context.Context, repoIDs ...api.RepoID) {
	numberOfRepos := len(repoIDs)
	if numberOfRepos == 0 {
		return
	} else if s.isDisabled() {
		s.logger.Debug("ScheduleRepos.disabled", log.Int("len(repoIDs)", len(repoIDs)))
		return
	}

	repositories := make([]scheduledRepo, numberOfRepos)
	for i := range repoIDs {
		repositories[i] = scheduledRepo{
			priority: priorityHigh,
			repoID:   repoIDs[i],
			// NOTE: Have processAfter with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	scheduleReposCounter.Add(float64(numberOfRepos))
	s.scheduleRepos(ctx, repositories...)
	metricsItemsSyncScheduled.WithLabelValues("manualReposTrigger", "high").Set(float64(numberOfRepos))
	s.collectQueueSize()
}

func (s *PermsSyncer) scheduleRepos(ctx context.Context, repos ...scheduledRepo) {
	logger := s.logger.Scoped("scheduleRepos", "")
	for _, r := range repos {
		select {
		case <-ctx.Done():
			logger.Debug("canceled")
			return
		default:
		}

		updated := s.queue.enqueue(&requestMeta{
			Priority:   r.priority,
			Type:       requestTypeRepo,
			ID:         int32(r.repoID),
			NextSyncAt: r.processAfter,
			NoPerms:    r.noPerms,
		})
		logger.Debug("queue.enqueued", log.Int32("repoID", int32(r.repoID)), log.Bool("updated", updated))
	}
}

// providersByServiceID returns a list of authz.Provider configured in the external services.
// Keys are ServiceID, e.g. "https://github.com/".
func (s *PermsSyncer) providersByServiceID() map[string]authz.Provider {
	_, ps := authz.GetProviders()
	providers := make(map[string]authz.Provider, len(ps))
	for _, p := range ps {
		providers[p.ServiceID()] = p
	}
	return providers
}

// providersByURNs returns a list of authz.Provider configured in the external services.
// Keys are URN, e.g. "extsvc:github:1".
func (s *PermsSyncer) providersByURNs() map[string]authz.Provider {
	_, ps := authz.GetProviders()
	providers := make(map[string]authz.Provider, len(ps))
	for _, p := range ps {
		providers[p.URN()] = p
	}
	return providers
}

// listPrivateRepoNamesBySpecs slices over the `repoSpecs` at pace of 10000
// elements at a time to workaround Postgres' limit of 65535 bind parameters
// using exact name matching. This method only includes private repository names
// and does not do deduplication on the returned list.
func (s *PermsSyncer) listPrivateRepoNamesBySpecs(ctx context.Context, repoSpecs []api.ExternalRepoSpec) ([]types.MinimalRepo, error) {
	if len(repoSpecs) == 0 {
		return []types.MinimalRepo{}, nil
	}

	remaining := repoSpecs
	nextCut := 10000
	if len(remaining) < nextCut {
		nextCut = len(remaining)
	}

	repoNames := make([]types.MinimalRepo, 0, len(repoSpecs))
	for nextCut > 0 {
		rs, err := s.reposStore.RepoStore().ListMinimalRepos(ctx,
			database.ReposListOptions{
				ExternalRepos: remaining[:nextCut],
				OnlyPrivate:   true,
			},
		)
		if err != nil {
			return nil, err
		}

		repoNames = append(repoNames, rs...)

		remaining = remaining[nextCut:]
		if len(remaining) < nextCut {
			nextCut = len(remaining)
		}
	}
	return repoNames, nil
}

type fetchUserPermsViaExternalAccountsResults struct {
	repoIDs      []uint32
	subRepoPerms map[api.ExternalRepoSpec]*authz.SubRepoPermissions

	providerStates providerStatesSet
}

// fetchUserPermsViaExternalAccounts uses external accounts (aka. login
// connections) to list all accessible private repositories on code hosts for
// the given user.
//
// It returns a list of internal database repository IDs and is a noop when
// `envvar.SourcegraphDotComMode()` is true.
func (s *PermsSyncer) fetchUserPermsViaExternalAccounts(ctx context.Context, user *types.User, noPerms bool, fetchOpts authz.FetchPermsOptions) (results fetchUserPermsViaExternalAccountsResults, err error) {
	// NOTE: OAuth scope on sourcegraph.com does not grant access to read private
	//  repositories, therefore it is no point wasting effort and code host API rate
	//  limit quota on trying.
	if envvar.SourcegraphDotComMode() {
		return results, nil
	}

	// Update tokens stored in external accounts
	accts, err := s.db.UserExternalAccounts().List(ctx,
		database.ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return results, errors.Wrap(err, "list external accounts")
	}

	// We also want to include any expired accounts for GitLab as they can be
	// refreshed
	expireGitLabAccounts, err := s.db.UserExternalAccounts().List(ctx,
		database.ExternalAccountsListOptions{
			UserID:      user.ID,
			ServiceType: extsvc.TypeGitLab,
			OnlyExpired: true,
		},
	)
	if err != nil {
		return results, errors.Wrap(err, "list expired gitlab external accounts")
	}
	accts = append(accts, expireGitLabAccounts...)

	serviceToAccounts := make(map[string]*extsvc.Account)
	for _, acct := range accts {
		serviceToAccounts[acct.ServiceType+":"+acct.ServiceID] = acct
	}

	userEmails, err := s.db.UserEmails().ListByUser(ctx,
		database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		},
	)
	if err != nil {
		return results, errors.Wrap(err, "list user verified emails")
	}

	emails := make([]string, len(userEmails))
	for i := range userEmails {
		emails[i] = userEmails[i].Email
	}

	byServiceID := s.providersByServiceID()
	accounts := s.db.UserExternalAccounts()
	logger := s.logger.Scoped("fetchUserPermsViaExternalAccounts", "sync permissions using external accounts (logging connections)").With(log.Int32("userID", user.ID))

	// Check if the user has an external account for every authz provider respectively,
	// and try to fetch the account when not.
	for _, provider := range byServiceID {
		providerLogger := logger.With(log.String("authzProvider", provider.ServiceID()))
		_, ok := serviceToAccounts[provider.ServiceType()+":"+provider.ServiceID()]
		if ok {
			continue
		}

		acct, err := provider.FetchAccount(ctx, user, accts, emails)
		results.providerStates = append(results.providerStates, newProviderState(provider, err, "FetchAccount"))
		if err != nil {
			providerLogger.Error("could not fetch account from authz provider", log.Error(err))
			continue
		}

		// Not an operation failure but the authz provider is unable to determine
		// the external account for the current user.
		if acct == nil {
			providerLogger.Debug("no user account found for provider", log.String("provider_urn", provider.URN()), log.Int32("user_id", user.ID))
			continue
		}
		providerLogger.Debug("account found for provider", log.String("provider_urn", provider.URN()), log.Int32("user_id", user.ID), log.Int32("account_id", acct.ID))

		err = accounts.AssociateUserAndSave(ctx, user.ID, acct.AccountSpec, acct.AccountData)
		if err != nil {
			providerLogger.Error("could not associate external account to user", log.Error(err))
			continue
		}

		accts = append(accts, acct)
	}

	var repoSpecs, includeContainsSpecs, excludeContainsSpecs []api.ExternalRepoSpec
	results.subRepoPerms = make(map[api.ExternalRepoSpec]*authz.SubRepoPermissions)

	for _, acct := range accts {
		acctLogger := logger.With(log.Int32("acct.ID", acct.ID))

		provider := byServiceID[acct.ServiceID]
		if provider == nil {
			// We have no authz provider configured for this external account
			continue
		}

		acctLogger.Debug("update GitHub App installation access", log.Int32("accountID", acct.ID))

		// FetchUserPerms makes API requests using a client that will deal with the token
		// expiration and try to refresh it when necessary. If the client fails to update
		// the token, or if the token is revoked, the "401 Unauthorized" error will be
		// handled here.
		extPerms, err := provider.FetchUserPerms(ctx, acct, fetchOpts)
		results.providerStates = append(results.providerStates, newProviderState(provider, err, "FetchUserPerms"))
		if err != nil {
			acctLogger.Debug("error fetching user permissions", log.Error(err))

			unauthorized := errcode.IsUnauthorized(err)
			forbidden := errcode.IsForbidden(err)
			// Detect GitHub account suspension error
			accountSuspended := errcode.IsAccountSuspended(err)
			if unauthorized || accountSuspended || forbidden {
				// These are fatal errors that mean we should continue as if the account no
				// longer has any access.
				if err = accounts.TouchExpired(ctx, acct.ID); err != nil {
					return results, errors.Wrapf(err, "set expired for external account ID %v", acct.ID)
				}

				if unauthorized {
					acctLogger.Warn("setExternalAccountExpired, token is revoked",
						log.Bool("unauthorized", unauthorized),
					)
					continue
				}
				acctLogger.Debug("setExternalAccountExpired",
					log.Bool("unauthorized", unauthorized),
					log.Bool("accountSuspended", accountSuspended),
					log.Bool("forbidden", forbidden),
				)

				// We still want to continue processing other external accounts
				continue
			}

			// Skip this external account if unimplemented
			if errors.Is(err, &authz.ErrUnimplemented{}) {
				continue
			}

			if errcode.IsTemporary(err) {
				// If we have a temporary issue, we should instead return any permissions we
				// already know about to ensure that we don't temporarily remove access for the
				// user because of intermittent errors.
				acctLogger.Warn("temporary error, returning previously synced permissions", log.Error(err))

				extPerms = new(authz.ExternalUserPermissions)

				// Load last synced sub-repo perms for this user and provider
				currentSubRepoPerms, err := edb.NewEnterpriseDB(s.db).SubRepoPerms().GetByUserAndService(ctx, user.ID, provider.ServiceType(), provider.ServiceID())
				if err != nil {
					return results, errors.Wrap(err, "fetching existing sub-repo permissions")
				}
				extPerms.SubRepoPermissions = make(map[extsvc.RepoID]*authz.SubRepoPermissions, len(currentSubRepoPerms))
				for k := range currentSubRepoPerms {
					v := currentSubRepoPerms[k]
					extPerms.SubRepoPermissions[extsvc.RepoID(k.ID)] = &v
				}

				// Load last synced repos for this user and provider
				currentRepos, err := s.permsStore.FetchReposByUserAndExternalService(ctx, user.ID, provider.ServiceType(), provider.ServiceID())
				if err != nil {
					return results, errors.Wrap(err, "fetching existing repo permissions")
				}
				for _, id := range currentRepos {
					results.repoIDs = append(results.repoIDs, uint32(id))
				}

				continue
			}

			// Process partial results if this is an initial fetch.
			if !noPerms {
				return results, errors.Wrapf(err, "fetch user permissions for external account %d", acct.ID)
			}
			acctLogger.Warn("proceedWithPartialResults", log.Error(err))
		} else {
			err = accounts.TouchLastValid(ctx, acct.ID)
			if err != nil {
				return results, errors.Wrapf(err, "set last valid for external account %d", acct.ID)
			}
		}

		if extPerms == nil {
			continue
		}

		for _, exact := range extPerms.Exacts {
			repoSpecs = append(repoSpecs,
				api.ExternalRepoSpec{
					ID:          string(exact),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}

		// Record any sub-repository permissions
		for repoID := range extPerms.SubRepoPermissions {
			spec := api.ExternalRepoSpec{
				// This is safe since repoID is an extsvc.RepoID which represents the external id
				// of the repo.
				ID:          string(repoID),
				ServiceType: provider.ServiceType(),
				ServiceID:   provider.ServiceID(),
			}
			results.subRepoPerms[spec] = extPerms.SubRepoPermissions[repoID]
		}

		for _, includePrefix := range extPerms.IncludeContains {
			includeContainsSpecs = append(includeContainsSpecs,
				api.ExternalRepoSpec{
					ID:          string(includePrefix),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}

		for _, excludePrefix := range extPerms.ExcludeContains {
			excludeContainsSpecs = append(excludeContainsSpecs,
				api.ExternalRepoSpec{
					ID:          string(excludePrefix),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}
	}

	// Get corresponding internal database IDs
	repoNames, err := s.listPrivateRepoNamesBySpecs(ctx, repoSpecs)
	if err != nil {
		return results, errors.Wrap(err, "list external repositories by exact matching")
	}

	// Exclusions are relative to inclusions, so if there is no inclusion, exclusion
	// are meaningless and no need to trigger a DB query.
	if len(includeContainsSpecs) > 0 {
		rs, err := s.reposStore.RepoStore().ListMinimalRepos(ctx,
			database.ReposListOptions{
				ExternalRepoIncludeContains: includeContainsSpecs,
				ExternalRepoExcludeContains: excludeContainsSpecs,
				OnlyPrivate:                 true,
			},
		)
		if err != nil {
			return results, errors.Wrap(err, "list external repositories by contains matching")
		}
		repoNames = append(repoNames, rs...)
	}

	// repoIDs represents repos the user is allowed to read
	if len(results.repoIDs) == 0 {
		// We may already have some repos if we hit a temporary error above in which case
		// we don't want to clear it out
		results.repoIDs = make([]uint32, 0, len(repoNames))
	}
	for _, r := range repoNames {
		results.repoIDs = append(results.repoIDs, uint32(r.ID))
	}

	return results, nil
}

// syncUserPerms processes permissions syncing request in user-centric way. When `noPerms` is true,
// the method will use partial results to update permissions tables even when error occurs.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32, noPerms bool, fetchOpts authz.FetchPermsOptions) (result *database.SetPermissionsResult, providerStates []syncjobs.ProviderStatus, err error) {
	ctx, save := s.observe(ctx, "PermsSyncer.syncUserPerms", "")
	defer save(requestTypeUser, userID, &err)

	user, err := s.db.Users().GetByID(ctx, userID)
	if err != nil {
		return result, providerStates, errors.Wrap(err, "get user")
	}

	logger := s.logger.Scoped("syncUserPerms", "processes permissions sync request in user-centric way").With(
		log.Object("user",
			log.Int32("ID", userID),
			log.String("name", user.Username)),
	)

	// We call this when there are errors communicating with external services so
	// that we don't have the same user stuck at the front of the queue.
	tryTouchUserPerms := func() {
		if err := s.permsStore.TouchUserPermissions(ctx, userID); err != nil {
			logger.Warn("touching user permissions", log.Error(err))
		}
	}

	results, err := s.fetchUserPermsViaExternalAccounts(ctx, user, noPerms, fetchOpts)
	providerStates = results.providerStates
	if err != nil {
		tryTouchUserPerms()
		return result, providerStates, errors.Wrapf(err, "fetch permissions via external accounts for user %q (id: %d)", user.Username, user.ID)
	}

	// fetch current permissions from database
	oldPerms := &authz.UserPermissions{
		UserID: user.ID,
		Perm:   authz.Read,
		Type:   authz.PermRepos,
		IDs:    map[int32]struct{}{},
	}
	_ = s.permsStore.LoadUserPermissions(ctx, oldPerms)

	// Save new permissions to database
	p := &authz.UserPermissions{
		UserID: user.ID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
		IDs:    map[int32]struct{}{},
	}

	for i := range results.repoIDs {
		p.IDs[int32(results.repoIDs[i])] = struct{}{}
	}

	// Set sub-repository permissions
	srp := edb.NewEnterpriseDB(s.db).SubRepoPerms()
	for spec, perm := range results.subRepoPerms {
		if err := srp.UpsertWithSpec(ctx, user.ID, spec, *perm); err != nil {
			return result, providerStates, errors.Wrapf(err, "upserting sub repo perms %v for user %q (id: %d)", spec, user.Username, user.ID)
		}
	}

	if len(results.subRepoPerms) > 0 {
		logger.Debug("subRepoSynced",
			log.Int("count", len(results.subRepoPerms)),
		)
	}

	// NOTE: Please read the docstring of permsUpdateLock field for reasoning of the lock.
	s.permsUpdateLock.Lock()
	defer s.permsUpdateLock.Unlock()

	result, err = s.permsStore.SetUserPermissions(ctx, p)
	if err != nil {
		return result, providerStates, errors.Wrapf(err, "set user permissions for user %q (id: %d)", user.Username, user.ID)
	}

	logger.Debug("synced",
		log.Int("count", len(p.IDs)),
		log.Object("fetchOpts", log.Bool("InvalidateCache", fetchOpts.InvalidateCaches)),
	)

	metricsSuccessPermsSyncs.WithLabelValues("user").Inc()

	if !oldPerms.SyncedAt.IsZero() {
		metricsPermsConsecutiveSyncDelay.WithLabelValues("user").Set(p.SyncedAt.Sub(oldPerms.SyncedAt).Seconds())
	} else {
		metricsFirstPermsSyncs.WithLabelValues("user").Inc()
		metricsPermsFirstSyncDelay.WithLabelValues("user").Set(p.SyncedAt.Sub(user.CreatedAt).Seconds())
	}

	return result, providerStates, nil
}

// syncRepoPerms processes permissions syncing request in repository-centric way.
// When `noPerms` is true, the method will use partial results to update permissions
// tables even when error occurs.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID api.RepoID, noPerms bool, fetchOpts authz.FetchPermsOptions) (result *database.SetPermissionsResult, providerStates []syncjobs.ProviderStatus, err error) {
	ctx, save := s.observe(ctx, "PermsSyncer.syncRepoPerms", "")
	defer save(requestTypeRepo, int32(repoID), &err)

	repo, err := s.reposStore.RepoStore().Get(ctx, repoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return result, providerStates, nil
		}
		return result, providerStates, errors.Wrap(err, "get repository")
	}
	var provider authz.Provider

	// Only check authz provider for private repositories because we only need to
	// fetch permissions for private repositories.
	if repo.Private {
		// Loop over repository's sources and see if matching any authz provider's URN.
		providers := s.providersByURNs()
		for urn := range repo.Sources {
			p, ok := providers[urn]
			if ok {
				provider = p
				break
			}
		}
	}

	logger := s.logger.Scoped("syncRepoPerms", "processes permissions syncing request in a repo-centric way").With(
		log.Object("repo",
			log.Int32("ID", int32(repo.ID)),
			log.String("name", string(repo.Name)),
			log.Bool("private", repo.Private)),
	)

	// For non-private repositories, we rely on the fact that the `provider` is
	// always nil and we do not restrict access
	if provider == nil {
		logger.Debug("skipFetchPerms")

		// We have no authz provider configured for the repository.
		// However, we need to upsert the dummy record in order to
		// prevent scheduler keep scheduling this repository.
		return result, providerStates, errors.Wrap(s.permsStore.TouchRepoPermissions(ctx, int32(repoID)), "touch repository permissions")
	}

	pendingAccountIDsSet := make(map[string]struct{})
	accountIDsToUserIDs := make(map[string]int32) // Account ID -> User ID

	extAccountIDs, err := provider.FetchRepoPerms(ctx, &extsvc.Repository{
		URI:              repo.URI,
		ExternalRepoSpec: repo.ExternalRepo,
	}, fetchOpts)
	providerStates = append(providerStates, newProviderState(provider, err, "FetchRepoPerms"))

	// Detect 404 error (i.e. not authorized to call given APIs) that often happens with GitHub.com
	// when the owner of the token only has READ access. However, we don't want to fail
	// so the scheduler won't keep trying to fetch permissions of this same repository, so we
	// return a nil error and log a warning message.
	var apiErr *github.APIError
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusNotFound {
		logger.Warn("ignoreUnauthorizedAPIError",
			log.Error(err),
			log.String("suggestion", "GitHub access token user may only have read access to the repository, but needs write for permissions"),
		)
		return result, providerStates, errors.Wrap(s.permsStore.TouchRepoPermissions(ctx, int32(repoID)), "touch repository permissions")
	}

	// Skip repo if unimplemented
	if errors.Is(err, &authz.ErrUnimplemented{}) {
		logger.Debug("unimplemented", log.Error(err))

		// We should still touch the repo perms so that we don't keep scheduling the repo
		// for permissions syncs on a tight interval.
		if err = s.permsStore.TouchRepoPermissions(ctx, int32(repoID)); err != nil {
			logger.Warn("error touching permissions for unimplemented authz provider", log.Error(err))
		}

		return result, providerStates, nil
	}

	if err != nil {
		// Process partial results if this is an initial fetch.
		if !noPerms {
			return result, providerStates, errors.Wrapf(err, "fetch repository permissions for repository %q (id: %d)", repo.Name, repo.ID)
		}
		logger.Warn("proceedWithPartialResults", log.Error(err))
	}

	if len(extAccountIDs) > 0 {
		accountIDs := make([]string, len(extAccountIDs))
		for i := range extAccountIDs {
			accountIDs[i] = string(extAccountIDs[i])
		}

		// Get corresponding internal database IDs
		accountIDsToUserIDs, err = s.permsStore.GetUserIDsByExternalAccounts(ctx, &extsvc.Accounts{
			ServiceType: provider.ServiceType(),
			ServiceID:   provider.ServiceID(),
			AccountIDs:  accountIDs,
		})

		if err != nil {
			return result, providerStates, errors.Wrapf(err, "get user IDs by external accounts for repository %q (id: %d)", repo.Name, repo.ID)
		}

		// Set up the set of all account IDs that need to be bound to permissions
		pendingAccountIDsSet = make(map[string]struct{}, len(accountIDs))
		for i := range accountIDs {
			pendingAccountIDsSet[accountIDs[i]] = struct{}{}
		}
	}

	// Load current permissions from database
	oldPerms := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read,
		UserIDs: map[int32]struct{}{},
	}
	_ = s.permsStore.LoadRepoPermissions(ctx, oldPerms)

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: map[int32]struct{}{},
	}

	for aid, uid := range accountIDsToUserIDs {
		// Add existing user to permissions
		p.UserIDs[uid] = struct{}{}

		// Remove existing user from the set of pending users
		delete(pendingAccountIDsSet, aid)
	}

	pendingAccountIDs := make([]string, 0, len(pendingAccountIDsSet))
	for aid := range pendingAccountIDsSet {
		pendingAccountIDs = append(pendingAccountIDs, aid)
	}

	// NOTE: Please read the docstring of permsUpdateLock field for reasoning of the lock.
	s.permsUpdateLock.Lock()
	defer s.permsUpdateLock.Unlock()

	txs, err := s.permsStore.Transact(ctx)
	if err != nil {
		return result, providerStates, errors.Wrapf(err, "start transaction for repository %q (id: %d)", repo.Name, repo.ID)
	}
	defer func() { err = txs.Done(err) }()

	result, err = txs.SetRepoPermissions(ctx, p)
	if err != nil {
		return result, providerStates, errors.Wrapf(err, "set repository permissions for repository %q (id: %d)", repo.Name, repo.ID)
	}
	regularCount := len(p.UserIDs)

	accounts := &extsvc.Accounts{
		ServiceType: provider.ServiceType(),
		ServiceID:   provider.ServiceID(),
		AccountIDs:  pendingAccountIDs,
	}
	if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
		return result, providerStates, errors.Wrapf(err, "set repository pending permissions for repository %q (id: %d)", repo.Name, repo.ID)
	}
	pendingCount := len(p.UserIDs)

	metricsSuccessPermsSyncs.WithLabelValues("repo").Inc()

	var delayMetricField log.Field
	if !oldPerms.SyncedAt.IsZero() {
		delay := p.SyncedAt.Sub(oldPerms.SyncedAt)
		metricsPermsConsecutiveSyncDelay.WithLabelValues("repo").Set(delay.Seconds())
		delayMetricField = log.Duration("consecutiveSyncDelay", delay)
	} else {
		metricsFirstPermsSyncs.WithLabelValues("repo").Inc()
		delay := p.SyncedAt.Sub(repo.CreatedAt)
		metricsPermsFirstSyncDelay.WithLabelValues("repo").Set(delay.Seconds())
		delayMetricField = log.Duration("consecutiveSyncDelay", delay)
	}

	logger.Debug("synced",
		log.Int("regularCount", regularCount),
		log.Int("pendingCount", pendingCount),
		log.Object("fetchOpts", log.Bool("invalidateCaches", fetchOpts.InvalidateCaches)),
		delayMetricField,
	)

	return result, providerStates, nil
}

// syncPerms processes the permissions syncing request and removes the request
// from the queue once the process is done (regardless of success or failure).
// The given sync groups are used to control the max concurrency, this method
// only returns when the sync process is spawned, and blocks when it reaches max
// concurrency defined by the sync group.
func (s *PermsSyncer) syncPerms(ctx context.Context, syncGroups map[requestType]group.ContextGroup, request *syncRequest) {
	logger := s.logger.Scoped("syncPerms", "process perms sync request").With(
		log.Object("request",
			log.String("type", request.Type.String()),
			log.Int32(request.IDFieldName(), request.ID),
		))

	defer s.queue.remove(request.Type, request.ID, true)

	var runSync func() (*database.SetPermissionsResult, providerStatesSet, error)
	switch request.Type {
	case requestTypeUser:
		runSync = func() (*database.SetPermissionsResult, providerStatesSet, error) {
			// Ensure the job field is recorded when monitoring external API calls
			ctx = metrics.ContextWithTask(ctx, "SyncUserPerms")
			return s.syncUserPerms(ctx, request.ID, request.NoPerms, request.Options)
		}
	case requestTypeRepo:
		runSync = func() (*database.SetPermissionsResult, providerStatesSet, error) {
			// Ensure the job field is recorded when monitoring external API calls
			ctx = metrics.ContextWithTask(ctx, "SyncRepoPerms")
			return s.syncRepoPerms(ctx, api.RepoID(request.ID), request.NoPerms, request.Options)
		}
	default:
		logger.Error("unexpected request type", log.Int("type", int(request.Type)))
		return
	}

	// The call is blocked if reached max concurrency
	syncGroups[request.Type].Go(
		func(ctx context.Context) error {
			metricsConcurrentSyncs.WithLabelValues(request.Type.String()).Inc()
			defer metricsConcurrentSyncs.WithLabelValues(request.Type.String()).Dec()

			_, providerStates, err := runSync()
			if err != nil {
				logger.Error("failed to sync permissions",
					providerStates.SummaryField(),
					log.Error(err),
				)

				if request.Type == requestTypeUser {
					metricsFailedPermsSyncs.WithLabelValues("user").Inc()
				} else {
					metricsFailedPermsSyncs.WithLabelValues("repo").Inc()
				}
			} else {
				logger.Debug("succeeded in syncing permissions",
					providerStates.SummaryField())
			}

			s.collectQueueSize()
			s.recordsStore.Record(request.Type.String(), request.ID, providerStates, err)

			return nil
		},
	)

	// NOTE: We do not need to call Wait method of the sync group because all we need
	// here is to control the max concurrency rather than sending out a batch of jobs
	// and collecting their results at the same point in time. Especially true when
	// the background permissions syncing is a never-ending process, and it is
	// desired to abort and give up any running jobs ASAP upon quitting the program.
}

func (s *PermsSyncer) runSync(ctx context.Context) {
	logger := s.logger.Scoped("runSync", "routine to start processing the sync request queue")
	defer logger.Info("stopped")

	userMaxConcurrency := syncUsersMaxConcurrency()
	logger.Debug("started", log.Int("syncUsersMaxConcurrency", userMaxConcurrency))

	syncGroups := map[requestType]group.ContextGroup{
		requestTypeUser: group.New().WithContext(ctx).WithMaxConcurrency(userMaxConcurrency),

		// NOTE: This is not strictly needed as part of effort for
		// https://github.com/sourcegraph/sourcegraph/issues/37918, but doing it this way
		// has much simpler code logic and is much easier to reason about the behavior.
		//
		// It is also worth noting that naively increasing the max concurrency of
		// repo-centric syncing for GitHub may not work as intended because all sync jobs
		// derived from the same code host connection is sharing the same personal access
		// token and its concurrency throttled to 1 by the github-proxy in the current
		// architecture.
		requestTypeRepo: group.New().WithContext(ctx).WithMaxConcurrency(1),
	}

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

			logger.Debug("waitForNextSync", log.Duration("duration", wait))
			continue
		}

		notify(notifyDequeued)

		s.syncPerms(ctx, syncGroups, request)
	}
}

// scheduleUsersWithNoPerms returns computed schedules for users who have no
// permissions found in database.
func (s *PermsSyncer) scheduleUsersWithNoPerms(ctx context.Context) ([]scheduledUser, error) {
	ids, err := s.permsStore.UserIDsWithNoPerms(ctx)
	if err != nil {
		return nil, err
	}
	metricsNoPerms.WithLabelValues("user").Set(float64(len(ids)))

	users := make([]scheduledUser, len(ids))
	for i, id := range ids {
		users[i] = scheduledUser{
			priority: priorityLow,
			userID:   id,
			// NOTE: Have processAfter with zero value (i.e. not set) gives it higher priority.
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
	metricsNoPerms.WithLabelValues("repo").Set(float64(len(ids)))

	repositories := make([]scheduledRepo, len(ids))
	for i, id := range ids {
		repositories[i] = scheduledRepo{
			priority: priorityLow,
			repoID:   id,
			// NOTE: Have processAfter with zero value (i.e. not set) gives it higher priority.
			noPerms: true,
		}
	}
	return repositories, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have the
// oldest permissions in database and capped results by the limit.
func (s *PermsSyncer) scheduleUsersWithOldestPerms(ctx context.Context, limit int, age time.Duration) ([]scheduledUser, error) {
	results, err := s.permsStore.UserIDsWithOldestPerms(ctx, limit, age)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, 0, len(results))
	for id, t := range results {
		users = append(users, scheduledUser{
			priority:     priorityLow,
			userID:       id,
			processAfter: t,
		})
	}
	return users, nil
}

// scheduleReposWithOldestPerms returns computed schedules for private repositories that
// have oldest permissions in database.
func (s *PermsSyncer) scheduleReposWithOldestPerms(ctx context.Context, limit int, age time.Duration) ([]scheduledRepo, error) {
	results, err := s.permsStore.ReposIDsWithOldestPerms(ctx, limit, age)
	if err != nil {
		return nil, err
	}

	repositories := make([]scheduledRepo, 0, len(results))
	for id, t := range results {
		repositories = append(repositories, scheduledRepo{
			priority:     priorityLow,
			repoID:       id,
			processAfter: t,
		})
	}
	return repositories, nil
}

// schedule contains information for scheduling users and repositories.
type schedule struct {
	Users []scheduledUser
	Repos []scheduledRepo
}

// scheduledUser contains information for scheduling a user.
type scheduledUser struct {
	priority     priority
	userID       int32
	options      authz.FetchPermsOptions
	processAfter time.Time

	// Whether the user has no permissions when scheduled. Currently used to
	// accept partial results from authz provider in case of error.
	noPerms bool
}

// scheduledRepo contains for scheduling a repository.
type scheduledRepo struct {
	priority     priority
	repoID       api.RepoID
	processAfter time.Time

	// Whether the repository has no permissions when scheduled. Currently used
	// to accept partial results from authz provider in case of error.
	noPerms bool
}

// schedule computes schedule four lists in the following order:
//  1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//  2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//  3. Rolling updating user permissions over time from oldest ones.
//  4. Rolling updating repository permissions over time from oldest ones.
func (s *PermsSyncer) schedule(ctx context.Context) (*schedule, error) {
	schedule := new(schedule)

	usersWithNoPerms, err := s.scheduleUsersWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule users with no permissions")
	}
	schedule.Users = append(schedule.Users, usersWithNoPerms...)

	reposWithNoPerms, err := s.scheduleReposWithNoPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule repositories with no permissions")
	}
	schedule.Repos = append(schedule.Repos, reposWithNoPerms...)

	// TODO(jchen): Predict a limit taking account into:
	//   1. Based on total repos and users that make sense to finish syncing before
	//      next schedule call, so we don't waste database bandwidth.
	//   2. How we're doing in terms of rate limiting.
	// Formula (in worse case scenario, at the pace of 1 req/s):
	//   initial limit  = <predicted from the previous step>
	//	 consumed by users = <initial limit> / (<total repos> / <page size>)
	//   consumed by repos = (<initial limit> - <consumed by users>) / (<total users> / <page size>)
	// Hard coded both to 10 for now.
	userLimit, repoLimit := oldestUserPermissionsBatchSize(), oldestRepoPermissionsBatchSize()

	usersWithOldestPerms, err := s.scheduleUsersWithOldestPerms(ctx, userLimit, syncUserBackoff())
	if err != nil {
		return nil, errors.Wrap(err, "load users with oldest permissions")
	}
	schedule.Users = append(schedule.Users, usersWithOldestPerms...)

	reposWithOldestPerms, err := s.scheduleReposWithOldestPerms(ctx, repoLimit, syncRepoBackoff())
	if err != nil {
		return nil, errors.Wrap(err, "scan repositories with oldest permissions")
	}
	schedule.Repos = append(schedule.Repos, reposWithOldestPerms...)

	metricsItemsSyncScheduled.WithLabelValues("usersWithNoPerms", "low").Set(float64(len(usersWithNoPerms)))
	metricsItemsSyncScheduled.WithLabelValues("usersWithOldestPerms", "low").Set(float64(len(usersWithOldestPerms)))
	metricsItemsSyncScheduled.WithLabelValues("reposWithNoPerms", "low").Set(float64(len(reposWithNoPerms)))
	metricsItemsSyncScheduled.WithLabelValues("reposWithOldestPerms", "low").Set(float64(len(reposWithOldestPerms)))

	return schedule, nil
}

func (s *PermsSyncer) isDisabled() bool { return PermissionSyncingDisabled() }

// runSchedule periodically looks for least updated records and schedule syncs
// for them.
func (s *PermsSyncer) runSchedule(ctx context.Context) {
	logger := s.logger.Scoped("runSchedule", "periodically queue old records for sync")

	logger.Info("started")
	defer logger.Info("stopped")

	ticker := time.NewTicker(s.scheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		if s.isDisabled() || permssync.PermissionSyncWorkerEnabled(ctx, s.db, logger) {
			logger.Info("disabled")
			continue
		}

		schedule, err := s.schedule(ctx)
		if err != nil {
			logger.Error("failed to compute schedule", log.Error(err))
			continue
		}

		logger.Info("scheduling permission syncs", log.Int("users", len(schedule.Users)), log.Int("repos", len(schedule.Repos)))

		s.scheduleUsers(ctx, schedule.Users...)
		s.scheduleRepos(ctx, schedule.Repos...)

		s.collectMetrics(ctx)
	}
}

// DebugDump returns the state of the permissions syncer for debugging.
func (s *PermsSyncer) DebugDump(_ context.Context) any {
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
				Options:    request.Options,
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
		tr.SetAttributes(attribute.Int64("id", int64(id)))

		var typLabel string
		switch typ {
		case requestTypeRepo:
			typLabel = "repo"
		case requestTypeUser:
			typLabel = "user"
		default:
			tr.SetError(errors.Errorf("unexpected request type: %v", typ))
			return
		}

		success := err == nil || *err == nil
		metricsSyncDuration.WithLabelValues(typLabel, strconv.FormatBool(success)).Observe(time.Since(began).Seconds())

		if !success {
			tr.SetError(*err)
			metricsSyncErrors.WithLabelValues(typLabel).Inc()
		}
	}
}

var collectMetricsDisabled = false

func (s *PermsSyncer) collectQueueSize() {
	if collectMetricsDisabled {
		return
	}

	s.queue.mu.RLock()
	metricsQueueSize.Set(float64(s.queue.Len()))
	s.queue.mu.RUnlock()
}

// collectMetrics collects metrics values from both database and memory objects.
func (s *PermsSyncer) collectMetrics(ctx context.Context) {
	if collectMetricsDisabled {
		return
	}

	logger := s.logger.Scoped("collectMetrics", "")

	m, err := s.permsStore.Metrics(ctx, 3*24*time.Hour)
	if err != nil {
		logger.Error("failed to get metrics from database", log.Error(err))
		return
	}
	mstrict, err := s.permsStore.Metrics(ctx, 1*time.Hour)
	if err != nil {
		logger.Error("failed to get metrics from database", log.Error(err))
		return
	}

	metricsStalePerms.WithLabelValues("user").Set(float64(m.UsersWithStalePerms))
	metricsStrictStalePerms.WithLabelValues("user").Set(float64(mstrict.UsersWithStalePerms))
	metricsPermsGap.WithLabelValues("user").Set(m.UsersPermsGapSeconds)
	metricsStalePerms.WithLabelValues("repo").Set(float64(m.ReposWithStalePerms))
	metricsStrictStalePerms.WithLabelValues("repo").Set(float64(mstrict.ReposWithStalePerms))
	metricsPermsGap.WithLabelValues("repo").Set(m.ReposPermsGapSeconds)
	metricsStalePerms.WithLabelValues("sub-repo").Set(float64(m.SubReposWithStalePerms))
	metricsStrictStalePerms.WithLabelValues("sub-repo").Set(float64(mstrict.SubReposWithStalePerms))
	metricsPermsGap.WithLabelValues("sub-repo").Set(m.SubReposPermsGapSeconds)

	s.collectQueueSize()
}

// Run kicks off the permissions syncing process, this method is blocking and
// should be called as a goroutine.
func (s *PermsSyncer) Run(ctx context.Context) {
	go s.runSync(ctx)
	go s.runSchedule(ctx)

	<-ctx.Done()
}

func scheduleInterval() time.Duration {
	seconds := conf.Get().PermissionsSyncScheduleInterval
	if seconds <= 0 {
		return 15 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func oldestUserPermissionsBatchSize() int {
	batchSize := conf.Get().PermissionsSyncOldestUsers
	if batchSize <= 0 {
		return 10
	}
	return batchSize
}

func oldestRepoPermissionsBatchSize() int {
	batchSize := conf.Get().PermissionsSyncOldestRepos
	if batchSize <= 0 {
		return 10
	}
	return batchSize
}

func syncUserBackoff() time.Duration {
	seconds := conf.Get().PermissionsSyncUsersBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func syncRepoBackoff() time.Duration {
	seconds := conf.Get().PermissionsSyncReposBackoffSeconds
	if seconds <= 0 {
		return 60 * time.Second
	}
	return time.Duration(seconds) * time.Second
}

func syncUsersMaxConcurrency() int {
	n := conf.Get().PermissionsSyncUsersMaxConcurrency
	if n <= 0 {
		return 1
	}
	return n
}
