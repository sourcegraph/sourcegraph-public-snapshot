package authz

import (
	"container/heap"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/oauth2"

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
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
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
	// The generic database handle.
	db database.DB
	// The database interface for any repos and external services operations.
	reposStore repos.Store
	// The database interface for any permissions operations.
	permsStore edb.PermsStore
	// The mockable function to return the current time.
	clock func() time.Time
	// The rate limit registry for code hosts.
	rateLimiterRegistry *ratelimit.Registry
	// The time duration of how often to re-compute schedule for users and repositories.
	scheduleInterval time.Duration
}

// NewPermsSyncer returns a new permissions syncing manager.
func NewPermsSyncer(
	db database.DB,
	reposStore repos.Store,
	permsStore edb.PermsStore,
	clock func() time.Time,
	rateLimiterRegistry *ratelimit.Registry,
) *PermsSyncer {
	return &PermsSyncer{
		queue:               newRequestQueue(),
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
		log15.Warn("PermsSyncer.ScheduleUsers.disabled", "userIDs", userIDs)
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
			Options:    u.options,
			NextSyncAt: u.nextSyncAt,
			NoPerms:    u.noPerms,
		})
		log15.Debug("PermsSyncer.queue.enqueued", "userID", u.userID, "updated", updated)
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
		log15.Warn("PermsSyncer.ScheduleRepos.disabled", "repoIDs", repoIDs)
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

func oauth2ConfigFromGitLabProvider(p *schema.GitLabAuthProvider) *oauth2.Config {
	url := strings.TrimSuffix(p.Url, "/")
	return &oauth2.Config{
		ClientID:     p.ClientID,
		ClientSecret: p.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  url + "/oauth/authorize",
			TokenURL: url + "/oauth/token",
		},
		Scopes: gitlab.RequestedOAuthScopes(p.ApiScope, nil),
	}
}

func (s *PermsSyncer) maybeRefreshGitLabOAuthTokenFromAccount(ctx context.Context, acct *extsvc.Account) (err error) {
	if acct.ServiceType != extsvc.TypeGitLab {
		return nil
	}

	var oauthConfig *oauth2.Config
	for _, authProvider := range conf.SiteConfig().AuthProviders {
		if authProvider.Gitlab == nil ||
			strings.TrimSuffix(acct.ServiceID, "/") != strings.TrimSuffix(authProvider.Gitlab.Url, "/") {
			continue
		}
		oauthConfig = oauth2ConfigFromGitLabProvider(authProvider.Gitlab)
		break
	}
	if oauthConfig == nil {
		log15.Warn("PermsSyncer.maybeRefreshGitLabOAuthTokenFromAccount, external account has no auth.provider",
			"externalAccountID", acct.ID,
		)
		return nil
	}

	_, tok, err := gitlab.GetExternalAccountData(&acct.AccountData)
	if err != nil {
		return errors.Wrap(err, "get external account data")
	} else if tok == nil {
		return errors.New("no token found in the external account data")
	}

	refreshedToken, err := oauthConfig.TokenSource(ctx, tok).Token()
	if err != nil {
		return errors.Wrap(err, "refresh token")
	}

	if refreshedToken.AccessToken != tok.AccessToken {
		defer func() {
			success := err == nil
			gitlab.TokenRefreshCounter.WithLabelValues("external_account", strconv.FormatBool(success)).Inc()
		}()
		acct.AccountData.SetAuthData(refreshedToken)
		_, err := s.db.UserExternalAccounts().LookupUserAndSave(ctx, acct.AccountSpec, acct.AccountData)
		if err != nil {
			return errors.Wrap(err, "save refreshed token")
		}
	}
	return nil
}

// fetchUserPermsViaExternalAccounts uses external accounts (aka. login
// connections) to list all accessible private repositories on code hosts for
// the given user.
//
// It returns a list of internal database repository IDs and is a noop when
// `envvar.SourcegraphDotComMode()` is true.
func (s *PermsSyncer) fetchUserPermsViaExternalAccounts(ctx context.Context, user *types.User, accts []*extsvc.Account, noPerms bool, fetchOpts authz.FetchPermsOptions) (repoIDs []uint32, subRepoPerms map[api.ExternalRepoSpec]*authz.SubRepoPermissions, err error) {
	// NOTE: OAuth scope on sourcegraph.com does not grant access to read private
	//  repositories, therefore it is no point wasting effort and code host API rate
	//  limit quota on trying.
	if envvar.SourcegraphDotComMode() {
		return []uint32{}, nil, nil
	}

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

	// Check if the user has an external account for every authz provider respectively,
	// and try to fetch the account when not.
	for _, provider := range byServiceID {
		_, ok := serviceToAccounts[provider.ServiceType()+":"+provider.ServiceID()]
		if ok {
			continue
		}

		acct, err := provider.FetchAccount(ctx, user, accts, emails)
		if err != nil {
			log15.Error("Could not fetch account from authz provider",
				"userID", user.ID,
				"authzProvider", provider.ServiceID(),
				"error", err,
			)
			continue
		}

		// Not an operation failure but the authz provider is unable to determine
		// the external account for the current user.
		if acct == nil {
			continue
		}

		err = accounts.AssociateUserAndSave(ctx, user.ID, acct.AccountSpec, acct.AccountData)
		if err != nil {
			log15.Error("Could not associate external account to user",
				"userID", user.ID,
				"authzProvider", provider.ServiceID(),
				"error", err)
			continue
		}

		accts = append(accts, acct)
	}

	var repoSpecs, includeContainsSpecs, excludeContainsSpecs []api.ExternalRepoSpec
	subRepoPerms = make(map[api.ExternalRepoSpec]*authz.SubRepoPermissions)

	for _, acct := range accts {
		provider := byServiceID[acct.ServiceID]
		if provider == nil {
			// We have no authz provider configured for this external account
			continue
		}

		if err := s.waitForRateLimit(ctx, provider.URN(), 1, "user"); err != nil {
			return nil, nil, errors.Wrap(err, "wait for rate limiter")
		}

		extPerms, err := provider.FetchUserPerms(ctx, acct, fetchOpts)
		if err != nil {
			// The "401 Unauthorized" is returned by code hosts when the token is revoked
			unauthorized := errcode.IsUnauthorized(err)

			forbidden := errcode.IsForbidden(err)

			// Detect GitHub account suspension error
			accountSuspended := errcode.IsAccountSuspended(err)

			if unauthorized || accountSuspended || forbidden {
				err = accounts.TouchExpired(ctx, acct.ID)
				if err != nil {
					return nil, nil, errors.Wrapf(err, "set expired for external account %d", acct.ID)
				}
				if unauthorized {
					log15.Warn("PermsSyncer.fetchUserPermsViaExternalAccounts.setExternalAccountExpired, token is revoked",
						"userID", user.ID,
						"id", acct.ID,
						"unauthorized", unauthorized,
					)
					continue
				}
				log15.Debug("PermsSyncer.fetchUserPermsViaExternalAccounts.setExternalAccountExpired",
					"userID", user.ID,
					"id", acct.ID,
					"unauthorized", unauthorized,
					"accountSuspended", accountSuspended,
					"forbidden", forbidden,
				)

				// We still want to continue processing other external accounts
				continue
			}

			// Skip this external account if unimplemented
			if errors.Is(err, &authz.ErrUnimplemented{}) {
				log15.Debug("PermsSyncer.fetchUserPermsViaExternalAccounts.unimplemented", "userID", user.ID, "id", acct.ID, "error", err)
				continue
			}

			// Process partial results if this is an initial fetch.
			if !noPerms {
				return nil, nil, errors.Wrapf(err, "fetch user permissions for external account %d", acct.ID)
			}
			log15.Warn("PermsSyncer.syncUserPerms.proceedWithPartialResults", "userID", user.ID, "error", err)
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
	repoIDs = make([]uint32, 0, len(repoNames))
	for _, r := range repoNames {
		repoIDs = append(repoIDs, uint32(r.ID))
	}

	return repoIDs, subRepoPerms, nil
}

// fetchUserPermsViaExternalServices uses user code connections to list all
// accessible private repositories on code hosts for the given user.
func (s *PermsSyncer) fetchUserPermsViaExternalServices(ctx context.Context, userID int32, fetchOpts authz.FetchPermsOptions) (repoIDs []uint32, err error) {
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
		provider, err := eauthz.ProviderFromExternalService(s.db.ExternalServices(), conf.Get().SiteConfiguration, svc, s.db)
		if err != nil {
			return nil, errors.Wrapf(err, "new provider from external service %d", svc.ID)
		}
		if provider == nil {
			// NOTE: User code host connection can only be added on sourcegraph.com, and
			//  authorization is enforced for everything, it does not make sense that we cannot
			//  derive an `authz.Provider` from it.
			log15.Warn("PermsSyncer.fetchUserPermsViaExternalServices.noAuthzProvider",
				"userID", userID,
				"id", svc.ID,
			)
			continue
		}

		token, err := extsvc.ExtractToken(svc.Config, svc.Kind)
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
				log15.Warn("PermsSyncer.fetchUserPermsViaExternalServices.expiredExternalService",
					"userID", userID,
					"id", svc.ID,
					"unauthorized", unauthorized,
					"accountSuspended", accountSuspended,
					"forbidden", forbidden,
				)

				// We still want to continue processing other external services
				continue
			}

			// Skip this external account if unimplemented
			if errors.Is(err, &authz.ErrUnimplemented{}) {
				log15.Debug("PermsSyncer.fetchUserPermsViaExternalServices.unimplemented", "userID", userID, "id", svc.ID, "error", err)
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
	ctx, save := s.observe(ctx, "PermsSyncer.syncUserPerms", "")
	defer save(requestTypeUser, userID, &err)

	user, err := s.db.Users().GetByID(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "get user")
	}

	// Update tokens stored in external accounts
	accts, err := s.db.UserExternalAccounts().List(ctx,
		database.ExternalAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return errors.Wrap(err, "list external accounts")
	}
	for _, acct := range accts {
		log15.Info("maybe refresh account", "userID", acct.UserID)
		if err := s.maybeRefreshGitLabOAuthTokenFromAccount(ctx, acct); err != nil {
			return errors.Wrap(err, "refreshing GitLab OAuth token for account")
		}
	}

	// NOTE: If a <repo_id, user_id> pair is present in the external_service_repos
	//  table, the user has proven that they have read access to the repository.
	repoIDs, err := s.reposStore.ListExternalServicePrivateRepoIDsByUserID(ctx, user.ID)
	if err != nil {
		return errors.Wrap(err, "list external service repo IDs by user ID")
	}

	externalAccountsRepoIDs, subRepoPerms, err := s.fetchUserPermsViaExternalAccounts(ctx, user, accts, noPerms, fetchOpts)
	if err != nil {
		return errors.Wrap(err, "fetch user permissions via external accounts")
	}

	externalServicesRepoIDs, err := s.fetchUserPermsViaExternalServices(ctx, user.ID, fetchOpts)
	if err != nil {
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
	err = s.permsStore.SetUserPermissions(ctx, p)
	if err != nil {
		return errors.Wrap(err, "set user permissions")
	}

	log15.Debug("PermsSyncer.syncUserPerms.synced",
		"userID", user.ID,
		"count", len(p.IDs),
		"fetchOpts.invalidateCaches", fetchOpts.InvalidateCaches,
	)

	// Set sub-repository permissions
	srp := s.db.SubRepoPerms()
	for spec, perm := range subRepoPerms {
		if err := srp.UpsertWithSpec(ctx, user.ID, spec, *perm); err != nil {
			return errors.Wrapf(err, "upserting sub repo perms %v for user %d", spec, user.ID)
		}
	}

	if len(subRepoPerms) > 0 {
		log15.Debug("PermsSyncer.syncUserPerms.subRepoSynced",
			"userID", user.ID,
			"count", len(subRepoPerms),
		)
	}

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

	// For non-private repositories, we rely on the fact that the `provider` is
	// always nil and no user IDs here because we don't restrict access to
	// non-private repositories.
	if provider == nil && len(userIDs) == 0 {
		log15.Debug("PermsSyncer.syncRepoPerms.skipFetchPerms",
			"repoID", repo.ID,
			"private", repo.Private,
		)

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
			log15.Warn("PermsSyncer.syncRepoPerms.ignoreUnauthorizedAPIError",
				"repoID", repo.ID,
				"err", err,
				"suggestion", "GitHub access token user may only have read access to the repository, but needs write for permissions",
			)
			return errors.Wrap(s.permsStore.TouchRepoPermissions(ctx, int32(repoID)), "touch repository permissions")
		}

		// Skip repo if unimplemented
		if errors.Is(err, &authz.ErrUnimplemented{}) {
			log15.Debug("PermsSyncer.syncRepoPerms.unimplemented", "repoID", repo.ID, "err", err)

			// We should still touch the repo perms so that we don't keep scheduling the repo
			// for permissions syncs on a tight interval.
			if err = s.permsStore.TouchRepoPermissions(ctx, int32(repoID)); err != nil {
				log15.Warn("Error touching permissions for unimplemented authz provider", "err", err)
			}

			return nil
		}

		if err != nil {
			// Process partial results if this is an initial fetch.
			if !noPerms {
				return errors.Wrap(err, "fetch repository permissions")
			}
			log15.Warn("PermsSyncer.syncRepoPerms.proceedWithPartialResults", "repoID", repo.ID, "err", err)
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

	txs, err := s.permsStore.Transact(ctx)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer func() { err = txs.Done(err) }()

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return errors.Wrap(err, "set repository permissions")
	}

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

	log15.Debug("PermsSyncer.syncRepoPerms.synced",
		"repoID", repo.ID,
		"name", repo.Name,
		"count", len(p.UserIDs),
		"fetchOpts.invalidateCaches", fetchOpts.InvalidateCaches,
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

// syncPerms processes the permissions syncing request and remove the request from
// the queue once it is done (independent of success or failure).
func (s *PermsSyncer) syncPerms(ctx context.Context, request *syncRequest) error {
	defer s.queue.remove(request.Type, request.ID, true)

	var err error
	switch request.Type {
	case requestTypeUser:
		// Ensure the job field is recorded when monitoring external API calls
		ctx = metrics.ContextWithTask(ctx, "SyncUserPerms")
		err = s.syncUserPerms(ctx, request.ID, request.NoPerms, request.Options)
	case requestTypeRepo:
		ctx = metrics.ContextWithTask(ctx, "SyncRepoPerms")
		err = s.syncRepoPerms(ctx, api.RepoID(request.ID), request.NoPerms, request.Options)
	default:
		err = errors.Errorf("unexpected request type: %v", request.Type)
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
			log15.Error("Failed to sync permissions", "type", request.Type, "id", request.ID, "err", err)
			continue
		}
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
//   1. Users with no permissions, because they can't do anything meaningful (e.g. not able to search).
//   2. Private repositories with no permissions, because those can't be viewed by anyone except site admins.
//   3. Rolling updating user permissions over time from oldest ones.
//   4. Rolling updating repository permissions over time from oldest ones.
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

		if s.isDisabled() {
			log15.Debug("PermsSyncer.runSchedule.disabled")
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
func (s *PermsSyncer) DebugDump() any {
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
		mstrict, err := s.permsStore.Metrics(ctx, 1*time.Hour)
		if err != nil {
			log15.Error("Failed to get metrics from database", "err", err)
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
