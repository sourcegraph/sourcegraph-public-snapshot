package authz

import (
	"container/heap"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	gh "github.com/google/go-github/v41/github"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	eauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

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
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	s.scheduleUsers(ctx, users...)
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
			NextSyncAt: u.nextSyncAt,
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

	repos := make([]scheduledRepo, numberOfRepos)
	for i := range repoIDs {
		repos[i] = scheduledRepo{
			priority: priorityHigh,
			repoID:   repoIDs[i],
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority,
			// as the request is most likely triggered by a user action from OSS namespace.
		}
	}

	scheduleReposCounter.Add(float64(numberOfRepos))
	s.scheduleRepos(ctx, repos...)
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
			NextSyncAt: r.nextSyncAt,
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

func (s *PermsSyncer) getUserGitHubAppInstallations(ctx context.Context, acct *extsvc.Account) ([]gh.Installation, error) {
	if acct.ServiceType != extsvc.TypeGitHub {
		return nil, nil
	}

	_, tok, err := github.GetExternalAccountData(ctx, &acct.AccountData)

	if err != nil {
		return nil, err
	}

	if tok == nil {
		return nil, nil
	}

	// Not a GitHub App access token
	if !github.IsGitHubAppAccessToken(tok.AccessToken) {
		return nil, nil
	}

	apiURL, err := url.Parse(acct.ServiceID)
	if err != nil {
		return nil, err
	}
	apiURL, _ = github.APIRoot(apiURL)
	ghClient := github.NewV3Client(log.Scoped("perms_syncer.github.v3", "github v3 client for perms syncer"),
		extsvc.URNGitHubOAuth, apiURL, &auth.OAuthBearerToken{Token: tok.AccessToken}, nil)

	installations, err := ghClient.GetUserInstallations(ctx)
	if err != nil {
		return nil, err
	}

	return installations, nil
}

// fetchUserPermsViaExternalAccounts uses external accounts (aka. login
// connections) to list all accessible private repositories on code hosts for
// the given user.
//
// It returns a list of internal database repository IDs and is a noop when
// `envvar.SourcegraphDotComMode()` is true.
func (s *PermsSyncer) fetchUserPermsViaExternalAccounts(ctx context.Context, user *types.User, noPerms bool, fetchOpts authz.FetchPermsOptions) (repoIDs []uint32, subRepoPerms map[api.ExternalRepoSpec]*authz.SubRepoPermissions, err error) {
	// NOTE: OAuth scope on sourcegraph.com does not grant access to read private
	//  repositories, therefore it is no point wasting effort and code host API rate
	//  limit quota on trying.
	if envvar.SourcegraphDotComMode() {
		return []uint32{}, nil, nil
	}

	// Update tokens stored in external accounts
	accts, err := s.db.UserExternalAccounts().List(ctx,
		database.ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "list external accounts")
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
		return nil, nil, errors.Wrap(err, "list expired gitlab external accounts")
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
		return nil, nil, errors.Wrap(err, "list user verified emails")
	}

	emails := make([]string, len(userEmails))
	for i := range userEmails {
		emails[i] = userEmails[i].Email
	}

	byServiceID := s.providersByServiceID()
	accounts := s.db.UserExternalAccounts()
	logger := s.logger.Scoped("fetchUserPermsViaExternalServices", "sync permissions using external accounts (logging connections)").With(log.Int32("userID", user.ID))

	// Check if the user has an external account for every authz provider respectively,
	// and try to fetch the account when not.
	for _, provider := range byServiceID {
		providerLogger := logger.With(log.String("authzProvider", provider.ServiceID()))
		_, ok := serviceToAccounts[provider.ServiceType()+":"+provider.ServiceID()]
		if ok {
			continue
		}

		acct, err := provider.FetchAccount(ctx, user, accts, emails)
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
	subRepoPerms = make(map[api.ExternalRepoSpec]*authz.SubRepoPermissions)

	for _, acct := range accts {
		if acct.ServiceType == extsvc.TypeGitHubApp {
			continue
		}

		acctLogger := logger.With(log.Int32("acct.ID", acct.ID))

		provider := byServiceID[acct.ServiceID]
		if provider == nil {
			// We have no authz provider configured for this external account
			continue
		}

		if err := s.waitForRateLimit(ctx, provider.URN(), 1, "user"); err != nil {
			return nil, nil, errors.Wrap(err, "wait for rate limiter")
		}

		acctLogger.Debug("update GitHub App installation access", log.Int32("accountID", acct.ID))

		installations, err := s.getUserGitHubAppInstallations(ctx, acct)

		// These errors aren't fatal, so we continue with the normal flow
		// even if things go wrong.
		if err != nil && installations != nil {
			if err := s.db.UserExternalAccounts().UpdateGitHubAppInstallations(ctx, acct, installations); err != nil {
				acctLogger.Warn("failed to update GitHub App installation access", log.Error(err))
			}
		} else if err != nil {
			acctLogger.Warn("failed to fetch GitHub App installations", log.Error(err))
		}

		extPerms, err := provider.FetchUserPerms(ctx, acct, fetchOpts)
		if err != nil {
			acctLogger.Debug("fetching user permissions", log.Error(err))

			// FetchUserPerms makes API requests using a client that will deal with the token
			// expiration and try to refresh it when necessary. If the client fails to update
			// the token, or if the token is revoked, the "401 Unauthorized" error will be
			// handled here.
			unauthorized := errcode.IsUnauthorized(err)
			forbidden := errcode.IsForbidden(err)
			// Detect GitHub account suspension error
			accountSuspended := errcode.IsAccountSuspended(err)
			if unauthorized || accountSuspended || forbidden {
				// These are fatal errors that mean we should continue as if the account no
				// longer has any access.
				linkedAccts, err := s.db.UserExternalAccounts().List(ctx,
					database.ExternalAccountsListOptions{
						ServiceType:   extsvc.TypeGitHubApp,
						AccountIDLike: fmt.Sprintf("%%/%s", acct.AccountID),
					},
				)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "list linked accounts for %d", acct.ID)
				}

				acctIDs := make([]int32, 0, len(linkedAccts)+1)
				acctIDs = append(acctIDs, acct.ID)
				for _, linkedAcct := range linkedAccts {
					acctIDs = append(acctIDs, linkedAcct.ID)
				}
				if err = accounts.TouchExpired(ctx, acctIDs...); err != nil {
					return nil, nil, errors.Wrapf(err, "set expired for external account IDs %v", acctIDs)
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
				currentSubRepoPerms, err := s.db.SubRepoPerms().GetByUserAndService(ctx, user.ID, provider.ServiceType(), provider.ServiceID())
				if err != nil {
					return nil, nil, errors.Wrap(err, "fetching existing sub-repo permissions")
				}
				extPerms.SubRepoPermissions = make(map[extsvc.RepoID]*authz.SubRepoPermissions, len(currentSubRepoPerms))
				for k := range currentSubRepoPerms {
					v := currentSubRepoPerms[k]
					extPerms.SubRepoPermissions[extsvc.RepoID(k.ID)] = &v
				}

				// Load last synced repos for this user and provider
				currentRepos, err := s.permsStore.FetchReposByUserAndExternalService(ctx, user.ID, provider.ServiceType(), provider.ServiceID())
				if err != nil {
					return nil, nil, errors.Wrap(err, "fetching existing repo permissions")
				}
				for _, id := range currentRepos {
					repoIDs = append(repoIDs, uint32(id))
				}

				continue
			}

			// Process partial results if this is an initial fetch.
			if !noPerms {
				return nil, nil, errors.Wrapf(err, "fetch user permissions for external account %d", acct.ID)
			}
			acctLogger.Warn("proceedWithPartialResults", log.Error(err))
		} else {
			err = accounts.TouchLastValid(ctx, acct.ID)
			if err != nil {
				return nil, nil, errors.Wrapf(err, "set last valid for external account %d", acct.ID)
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
			subRepoPerms[spec] = extPerms.SubRepoPermissions[repoID]
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
		return nil, nil, errors.Wrap(err, "list external repositories by exact matching")
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
			return nil, nil, errors.Wrap(err, "list external repositories by contains matching")
		}
		repoNames = append(repoNames, rs...)
	}

	// repoIDs represents repos the user is allowed to read
	if len(repoIDs) == 0 {
		// We may already have some repos if we hit a temporary error above in which case
		// we don't want to clear it out
		repoIDs = make([]uint32, 0, len(repoNames))
	}
	for _, r := range repoNames {
		repoIDs = append(repoIDs, uint32(r.ID))
	}

	return repoIDs, subRepoPerms, nil
}

// fetchUserPermsViaExternalServices uses user code connections to list all
// accessible private repositories on code hosts for the given user.
func (s *PermsSyncer) fetchUserPermsViaExternalServices(ctx context.Context, userID int32, fetchOpts authz.FetchPermsOptions) (repoIDs []uint32, err error) {
	logger := s.logger.Scoped("fetchUserPermsViaExternalServices", "sync permissions using code host connections").With(log.Int32("userID", userID))

	has, err := s.permsStore.UserIsMemberOfOrgHasCodeHostConnection(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "check user organization membership with a code host connection")
	}

	// NOTE: User-centric permissions syncing needs parity from the repo-centric
	//  permissions syncing. Therefore, if the user is not a member of any
	//  organization that has a code host connection connected, there is no point to
	//  do the user-centric syncing.
	if !has {
		return []uint32{}, nil
	}

	svcs, err := s.db.ExternalServices().List(ctx,
		database.ExternalServicesListOptions{
			NamespaceUserID: userID,
			Kinds:           []string{extsvc.KindGitHub, extsvc.KindGitLab},
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "list user external services")
	}

	var repoSpecs, includeContainsSpecs, excludeContainsSpecs []api.ExternalRepoSpec
	for _, svc := range svcs {
		svcLogger := logger.With(log.Int32("svc.ID", int32(svc.ID)))

		provider, err := eauthz.ProviderFromExternalService(ctx, s.db.ExternalServices(), conf.Get().SiteConfiguration, svc, s.db)
		if err != nil {
			return nil, errors.Wrapf(err, "new provider from external service %d", svc.ID)
		}
		if provider == nil {
			// NOTE: User code host connection can only be added on sourcegraph.com, and
			//  authorization is enforced for everything, it does not make sense that we cannot
			//  derive an `authz.Provider` from it.
			svcLogger.Warn("noAuthzProvider")
			continue
		}

		token, err := extsvc.ExtractEncryptableToken(ctx, svc.Config, svc.Kind)
		if err != nil {
			return nil, errors.Wrapf(err, "extract token from external service %d", svc.ID)
		}
		if token == "" {
			return nil, errors.Errorf("empty token from external service %d", svc.ID)
		}

		if err := s.waitForRateLimit(ctx, provider.URN(), 1, "user"); err != nil {
			return nil, errors.Wrap(err, "wait for rate limiter")
		}

		extPerms, err := provider.FetchUserPermsByToken(ctx, token, fetchOpts)
		if err != nil {
			// The "401 Unauthorized" is returned by code hosts when the token is no longer valid
			unauthorized := errcode.IsUnauthorized(err)

			forbidden := errcode.IsForbidden(err)

			// Detect GitHub account suspension error
			accountSuspended := errcode.IsAccountSuspended(err)

			if unauthorized || accountSuspended || forbidden {
				svcLogger.Warn("expiredExternalService",
					log.Bool("unauthorized", unauthorized),
					log.Bool("accountSuspended", accountSuspended),
					log.Bool("forbidden", forbidden),
				)

				// We still want to continue processing other external services
				continue
			}

			// Skip this external account if unimplemented
			if errors.Is(err, &authz.ErrUnimplemented{}) {
				svcLogger.Debug("unimplemented", log.Error(err))
				continue
			}

			return nil, errors.Wrapf(err, "fetch user permissions for external service %d", svc.ID)
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
		return nil, errors.Wrap(err, "list external repositories by exact matching")
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
			return nil, errors.Wrap(err, "list external repositories by contains matching")
		}
		repoNames = append(repoNames, rs...)
	}

	repoIDs = make([]uint32, 0, len(repoNames))
	for _, r := range repoNames {
		repoIDs = append(repoIDs, uint32(r.ID))
	}
	return repoIDs, nil
}

// syncUserPerms processes permissions syncing request in user-centric way. When `noPerms` is true,
// the method will use partial results to update permissions tables even when error occurs.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32, noPerms bool, fetchOpts authz.FetchPermsOptions) (err error) {
	logger := s.logger.Scoped("syncUserPerms", "processes permissions sync request in user-centric way").With(log.Int32("userID", userID))
	ctx, save := s.observe(ctx, "PermsSyncer.syncUserPerms", "")
	defer save(requestTypeUser, userID, &err)

	user, err := s.db.Users().GetByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "get user")
	}

	// NOTE: If a <repo_id, user_id> pair is present in the external_service_repos
	//  table, the user has proven that they have read access to the repository.
	repoIDs, err := s.reposStore.ListExternalServicePrivateRepoIDsByUserID(ctx, user.ID)
	if err != nil {
		return errors.Wrap(err, "list external service repo IDs by user ID")
	}

	// We call this when there are errors communicating with external services so
	// that we don't have the same user stuck at the front of the queue.
	tryTouchUserPerms := func() {
		if err := s.permsStore.TouchUserPermissions(ctx, userID); err != nil {
			logger.Warn("touching user permissions", log.Int32("userID", userID), log.Error(err))
		}
	}

	externalAccountsRepoIDs, subRepoPerms, err := s.fetchUserPermsViaExternalAccounts(ctx, user, noPerms, fetchOpts)
	if err != nil {
		tryTouchUserPerms()
		return errors.Wrap(err, "fetch user permissions via external accounts")
	}

	externalServicesRepoIDs, err := s.fetchUserPermsViaExternalServices(ctx, user.ID, fetchOpts)
	if err != nil {
		tryTouchUserPerms()
		return errors.Wrap(err, "fetch user permissions via external services")
	}

	// Save permissions to database
	p := &authz.UserPermissions{
		UserID: user.ID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
		IDs:    map[int32]struct{}{},
	}
	for i := range repoIDs {
		p.IDs[int32(repoIDs[i])] = struct{}{}
	}

	// Looping over two slices individually in order to avoid unnecessary memory allocation.
	for i := range externalAccountsRepoIDs {
		p.IDs[int32(externalAccountsRepoIDs[i])] = struct{}{}
	}
	for i := range externalServicesRepoIDs {
		p.IDs[int32(externalServicesRepoIDs[i])] = struct{}{}
	}

	// Set sub-repository permissions
	srp := s.db.SubRepoPerms()
	for spec, perm := range subRepoPerms {
		if err := srp.UpsertWithSpec(ctx, user.ID, spec, *perm); err != nil {
			return errors.Wrapf(err, "upserting sub repo perms %v for user %d", spec, user.ID)
		}
	}

	if len(subRepoPerms) > 0 {
		logger.Debug("subRepoSynced",
			log.Int("count", len(subRepoPerms)),
		)
	}

	// NOTE: Please read the docstring of permsUpdateLock field for reasoning of the lock.
	s.permsUpdateLock.Lock()
	defer s.permsUpdateLock.Unlock()

	err = s.permsStore.SetUserPermissions(ctx, p)
	if err != nil {
		return errors.Wrap(err, "set user permissions")
	}

	logger.Debug("synced",
		log.Int("count", len(p.IDs)),
		log.Object("fetchOpts", log.Bool("InvalidateCache", fetchOpts.InvalidateCaches)),
	)

	return nil
}

// syncRepoPerms processes permissions syncing request in repository-centric way.
// When `noPerms` is true, the method will use partial results to update permissions
// tables even when error occurs.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID api.RepoID, noPerms bool, fetchOpts authz.FetchPermsOptions) (err error) {
	ctx, save := s.observe(ctx, "PermsSyncer.syncRepoPerms", "")
	defer save(requestTypeRepo, int32(repoID), &err)

	rs, err := s.reposStore.RepoStore().List(ctx, database.ReposListOptions{
		IDs: []api.RepoID{repoID},
	})
	if err != nil {
		return errors.Wrap(err, "list repositories")
	} else if len(rs) == 0 {
		return nil
	}
	repo := rs[0]

	var userIDs []int32
	var provider authz.Provider

	// Only check authz provider for private repositories because we only need to
	// fetch permissions for private repositories.
	if repo.Private {
		// NOTE: If a <repo_id, user_id> pair is present in the external_service_repos
		//  table, the user has proven that they have read access to the repository.
		userIDs, err = s.reposStore.ListExternalServiceUserIDsByRepoID(ctx, repoID)
		if err != nil {
			return errors.Wrap(err, "list external service user IDs by repo ID")
		}

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
	// always nil and no user IDs here because we don't restrict access to
	// non-private repositories.
	if provider == nil && len(userIDs) == 0 {
		logger.Debug("skipFetchPerms")

		// We have no authz provider configured for the repository.
		// However, we need to upsert the dummy record in order to
		// prevent scheduler keep scheduling this repository.
		return errors.Wrap(s.permsStore.TouchRepoPermissions(ctx, int32(repoID)), "touch repository permissions")
	}

	pendingAccountIDsSet := make(map[string]struct{})
	accountIDsToUserIDs := make(map[string]int32) // Account ID -> User ID
	if provider != nil {
		if err := s.waitForRateLimit(ctx, provider.URN(), 1, "repo"); err != nil {
			return errors.Wrap(err, "wait for rate limiter")
		}

		extAccountIDs, err := provider.FetchRepoPerms(ctx, &extsvc.Repository{
			URI:              repo.URI,
			ExternalRepoSpec: repo.ExternalRepo,
		}, fetchOpts)

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
			return errors.Wrap(s.permsStore.TouchRepoPermissions(ctx, int32(repoID)), "touch repository permissions")
		}

		// Skip repo if unimplemented
		if errors.Is(err, &authz.ErrUnimplemented{}) {
			logger.Debug("unimplemented", log.Error(err))

			// We should still touch the repo perms so that we don't keep scheduling the repo
			// for permissions syncs on a tight interval.
			if err = s.permsStore.TouchRepoPermissions(ctx, int32(repoID)); err != nil {
				logger.Warn("error touching permissions for unimplemented authz provider", log.Error(err))
			}

			return nil
		}

		if err != nil {
			// Process partial results if this is an initial fetch.
			if !noPerms {
				return errors.Wrap(err, "fetch repository permissions")
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

			if provider.ServiceType() == extsvc.TypeGitHub {
				linkedAccountIDsToUserIDs, err := s.permsStore.GetUserIDsByExternalAccounts(ctx, &extsvc.Accounts{
					ServiceType: extsvc.TypeGitHubApp,
					ServiceID:   provider.ServiceID(),
					AccountIDs:  accountIDs,
				})
				if err == nil {
					for k, v := range linkedAccountIDsToUserIDs {
						accountIDsToUserIDs[k] = v
					}
				} else {
					// Only log in case of error, as there can still be valid permissions syncing.
					logger.Warn("error fetching linked accounts", log.Error(err))
				}
			}

			if err != nil {
				return errors.Wrap(err, "get user IDs by external accounts")
			}

			// Set up the set of all account IDs that need to be bound to permissions
			pendingAccountIDsSet = make(map[string]struct{}, len(accountIDs))
			for i := range accountIDs {
				pendingAccountIDsSet[accountIDs[i]] = struct{}{}
			}
		}
	}

	// Save permissions to database
	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: map[int32]struct{}{},
	}

	for i := range userIDs {
		p.UserIDs[userIDs[i]] = struct{}{}
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
		return errors.Wrap(err, "start transaction")
	}
	defer func() { err = txs.Done(err) }()

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	}
	regularCount := len(p.UserIDs)

	// If there is no provider, there would be no pending permissions that need to be generated.
	if provider != nil {
		accounts := &extsvc.Accounts{
			ServiceType: provider.ServiceType(),
			ServiceID:   provider.ServiceID(),
			AccountIDs:  pendingAccountIDs,
		}
		if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
			return errors.Wrap(err, "set repository pending permissions")
		}
	}
	pendingCount := len(p.UserIDs)

	logger.Debug("synced",
		log.Int("regularCount", regularCount),
		log.Int("pendingCount", pendingCount),
		log.Object("fetchOpts", log.Bool("invalidateCaches", fetchOpts.InvalidateCaches)),
	)
	return nil
}

// waitForRateLimit blocks until rate limit permits n events to happen. It returns
// an error if n exceeds the limiter's burst size, the context is canceled, or the
// expected wait time exceeds the context's deadline. The burst limit is ignored if
// the rate limit is Inf.
func (s *PermsSyncer) waitForRateLimit(ctx context.Context, urn string, n int, syncType string) error {
	if s.rateLimiterRegistry == nil {
		return nil
	}

	rl := s.rateLimiterRegistry.Get(urn)
	began := time.Now()
	if err := rl.WaitN(ctx, n); err != nil {
		metricsRateLimiterWaitDuration.WithLabelValues(syncType, strconv.FormatBool(false)).Observe(time.Since(began).Seconds())
		return err
	}
	metricsRateLimiterWaitDuration.WithLabelValues(syncType, strconv.FormatBool(true)).Observe(time.Since(began).Seconds())
	return nil
}

// syncPerms processes the permissions syncing request and removes the request
// from the queue once the process is done (regardless of success or failure).
// The given sync groups are used to control the max concurrency, this method
// only returns when the sync process is spawned, and blocks when it reaches max
// concurrency defined by the sync group.
func (s *PermsSyncer) syncPerms(ctx context.Context, logger log.Logger, syncGroups map[requestType]group.ContextGroup, request *syncRequest) {
	defer s.queue.remove(request.Type, request.ID, true)

	var runSync func() error
	switch request.Type {
	case requestTypeUser:
		runSync = func() error {
			// Ensure the job field is recorded when monitoring external API calls
			ctx = metrics.ContextWithTask(ctx, "SyncUserPerms")
			return s.syncUserPerms(ctx, request.ID, request.NoPerms, request.Options)
		}
	case requestTypeRepo:
		runSync = func() error {
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

			err := runSync()
			if err != nil {
				logger.Error("failed to sync permissions",
					log.Object("request",
						log.Int("type", int(request.Type)),
						log.Int32("id", request.ID),
					),
					log.Error(err),
				)
			}
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

		s.syncPerms(ctx, logger, syncGroups, request)
	}
}

// scheduleUsersWithOutdatedPerms returns computed schedules for users who have
// outdated permissions in database.
func (s *PermsSyncer) scheduleUsersWithOutdatedPerms(ctx context.Context) ([]scheduledUser, error) {
	results, err := s.permsStore.UserIDsWithOutdatedPerms(ctx)
	if err != nil {
		return nil, err
	}
	metricsOutdatedPerms.WithLabelValues("user").Set(float64(len(results)))

	users := make([]scheduledUser, 0, len(results))
	for id, t := range results {
		users = append(users,
			scheduledUser{
				priority:   priorityLow,
				userID:     id,
				nextSyncAt: t,
			},
		)
	}
	return users, nil
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
	metricsNoPerms.WithLabelValues("repo").Set(float64(len(ids)))

	repos := make([]scheduledRepo, len(ids))
	for i, id := range ids {
		repos[i] = scheduledRepo{
			priority: priorityLow,
			repoID:   id,
			// NOTE: Have nextSyncAt with zero value (i.e. not set) gives it higher priority.
			noPerms: true,
		}
	}
	return repos, nil
}

// scheduleUsersWithOldestPerms returns computed schedules for users who have oldest
// permissions in database and capped results by the limit.
func (s *PermsSyncer) scheduleUsersWithOldestPerms(ctx context.Context, limit int, age time.Duration) ([]scheduledUser, error) {
	results, err := s.permsStore.UserIDsWithOldestPerms(ctx, limit, age)
	if err != nil {
		return nil, err
	}

	users := make([]scheduledUser, 0, len(results))
	for id, t := range results {
		users = append(users, scheduledUser{
			priority:   priorityLow,
			userID:     id,
			nextSyncAt: t,
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

	repos := make([]scheduledRepo, 0, len(results))
	for id, t := range results {
		repos = append(repos, scheduledRepo{
			priority:   priorityLow,
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
	priority   priority
	userID     int32
	options    authz.FetchPermsOptions
	nextSyncAt time.Time

	// Whether the user has no permissions when scheduled. Currently used to
	// accept partial results from authz provider in case of error.
	noPerms bool
}

// scheduledRepo contains for scheduling a repository.
type scheduledRepo struct {
	priority   priority
	repoID     api.RepoID
	nextSyncAt time.Time

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

	users, err := s.scheduleUsersWithOutdatedPerms(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "schedule users with outdated permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	users, err = s.scheduleUsersWithNoPerms(ctx)
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
	userLimit, repoLimit := oldestUserPermissionsBatchSize(), oldestRepoPermissionsBatchSize()

	// TODO(jchen): Use better heuristics for setting NextSyncAt, the initial version
	// just uses the value of LastUpdatedAt get from the perms tables.

	users, err = s.scheduleUsersWithOldestPerms(ctx, userLimit, syncUserBackoff())
	if err != nil {
		return nil, errors.Wrap(err, "load users with oldest permissions")
	}
	schedule.Users = append(schedule.Users, users...)

	repos, err = s.scheduleReposWithOldestPerms(ctx, repoLimit, syncRepoBackoff())
	if err != nil {
		return nil, errors.Wrap(err, "scan repositories with oldest permissions")
	}
	schedule.Repos = append(schedule.Repos, repos...)

	return schedule, nil
}

// isDisabled returns true if the background permissions syncing is not enabled.
// It is not enabled if:
//   - Permissions user mapping is enabled
//   - Not purchased with the current license
//   - `disableAutoCodeHostSyncs` site setting is set to true
func (s *PermsSyncer) isDisabled() bool {
	return globals.PermissionsUserMapping().Enabled ||
		(licensing.EnforceTiers && licensing.Check(licensing.FeatureACLs) != nil) ||
		conf.Get().DisableAutoCodeHostSyncs
}

// runSchedule periodically looks for least updated records and schedule syncs
// for them.
func (s *PermsSyncer) runSchedule(ctx context.Context) {
	logger := s.logger.Scoped("runSchedule", "periodically queue old records for sync")

	logger.Debug("started")
	defer logger.Info("stopped")

	ticker := time.NewTicker(s.scheduleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}

		if s.isDisabled() {
			logger.Debug("disabled")
			continue
		}

		schedule, err := s.schedule(ctx)
		if err != nil {
			logger.Error("failed to compute schedule", log.Error(err))
			continue
		}
		s.scheduleUsers(ctx, schedule.Users...)
		s.scheduleRepos(ctx, schedule.Repos...)
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
		tr.LogFields(otlog.Int32("id", id))

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

// collectMetrics periodically collecting metrics values from both database and memory objects.
func (s *PermsSyncer) collectMetrics(ctx context.Context) {
	logger := s.logger.Scoped("collectMetrics", "")
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
			logger.Error("failed to get metrics from database", log.Error(err))
			continue
		}
		mstrict, err := s.permsStore.Metrics(ctx, 1*time.Hour)
		if err != nil {
			logger.Error("failed to get metrics from database", log.Error(err))
			continue
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

		s.queue.mu.RLock()
		metricsQueueSize.Set(float64(s.queue.Len()))
		s.queue.mu.RUnlock()
	}
}

// Run kicks off the permissions syncing process, this method is blocking and
// should be called as a goroutine.
func (s *PermsSyncer) Run(ctx context.Context) {
	go s.runSync(ctx)
	go s.runSchedule(ctx)
	go s.collectMetrics(ctx)

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
