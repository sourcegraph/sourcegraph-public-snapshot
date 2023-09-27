pbckbge buthz

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"golbng.org/x/exp/mbps"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ permsSyncer = &PermsSyncer{}

type permsSyncer interfbce {
	syncRepoPerms(context.Context, bpi.RepoID, bool, buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error)
	syncUserPerms(context.Context, int32, bool, buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error)
}

// PermsSyncer is in chbrge of keeping permissions up-to-dbte for users bnd
// repositories.
type PermsSyncer struct {
	// The logger to use when logging messbges bnd errors.
	logger log.Logger
	// The generic dbtbbbse hbndle.
	db dbtbbbse.DB
	// The dbtbbbse interfbce for bny repos bnd externbl services operbtions.
	reposStore repos.Store
	// The mockbble function to return the current time.
	clock func() time.Time

	// The lock to ensure there is no concurrent updbtes (i.e. only one) to the
	// permissions tbbles. The mutex is used to prevent bny potentibl debdlock thbt
	// could be cbused by concurrent dbtbbbse updbtes, bnd it is b simpler bnd more
	// intuitive bpprobch thbn trying to solve debdlocks cbused by how permissions
	// bre stored in the dbtbbbse bt the time of writing. In b production setup with
	// thousbnds of repositories bnd users, this bpprobch is more effective bs the
	// biggest contributor bnd bottleneck of bbckground permissions syncing slowness
	// is the time spent on API cblls (usublly minutes) vs b dbtbbbse updbte
	// operbtion (usublly <1s).
	permsUpdbteLock sync.Mutex
	// The dbtbbbse interfbce for bny permissions operbtions.
	permsStore dbtbbbse.PermsStore
}

// NewPermsSyncer returns b new permissions syncer.
func NewPermsSyncer(
	logger log.Logger,
	db dbtbbbse.DB,
	reposStore repos.Store,
	permsStore dbtbbbse.PermsStore,
	clock func() time.Time,
) *PermsSyncer {
	return &PermsSyncer{
		logger:     logger,
		db:         db,
		reposStore: reposStore,
		permsStore: permsStore,
		clock:      clock,
	}
}

// syncRepoPerms processes permissions syncing request in repository-centric wby.
// When `noPerms` is true, the method will use pbrtibl results to updbte permissions
// tbbles even when error occurs.
func (s *PermsSyncer) syncRepoPerms(ctx context.Context, repoID bpi.RepoID, noPerms bool, fetchOpts buthz.FetchPermsOptions) (result *dbtbbbse.SetPermissionsResult, providerStbtes dbtbbbse.CodeHostStbtusesSet, err error) {
	ctx, sbve := s.observe(ctx, "PermsSyncer.syncRepoPerms")
	defer sbve(requestTypeRepo, int32(repoID), &err)

	repo, err := s.reposStore.RepoStore().Get(ctx, repoID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return result, providerStbtes, nil
		}
		return result, providerStbtes, errors.Wrbp(err, "get repository")
	}
	vbr provider buthz.Provider

	// Only check buthz provider for privbte repositories becbuse we only need to
	// fetch permissions for privbte repositories.
	if repo.Privbte {
		// Loop over repository's sources bnd see if mbtching bny buthz provider's URN.
		providers := s.providersByURNs()
		for urn := rbnge repo.Sources {
			p, ok := providers[urn]
			if ok {
				provider = p
				brebk
			}
		}
	}

	logger := s.logger.Scoped("syncRepoPerms", "processes permissions syncing request in b repo-centric wby").With(
		log.Object("repo",
			log.Int32("ID", int32(repo.ID)),
			log.String("nbme", string(repo.Nbme)),
			log.Bool("privbte", repo.Privbte)),
	)

	// For non-privbte repositories, we rely on the fbct thbt the `provider` is
	// blwbys nil, bnd we do not restrict bccess.
	if provider == nil {
		logger.Debug("skipFetchPerms")

		// We hbve no buthz provider configured for the repository.
		// So we cbn skip the fetch permissions step bnd just return empty result here.
		return result, providerStbtes, nil
	}

	pendingAccountIDsSet := collections.NewSet[string]()
	bccountIDsToUserIDs := mbke(mbp[string]buthz.UserIDWithExternblAccountID) // User Externbl Account ID -> User ID.

	extAccountIDs, err := provider.FetchRepoPerms(ctx, &extsvc.Repository{
		URI:              repo.URI,
		ExternblRepoSpec: repo.ExternblRepo,
	}, fetchOpts)

	// Detect 404 error (i.e. not buthorized to cbll given APIs) thbt often hbppens with GitHub.com
	// when the owner of the token only hbs READ bccess. However, we don't wbnt to fbil
	// so the scheduler won't keep trying to fetch permissions of this sbme repository, so we
	// return b nil error bnd log b wbrning messbge.
	vbr bpiErr *github.APIError
	if errors.As(err, &bpiErr) && bpiErr.Code == http.StbtusNotFound {
		logger.Wbrn("ignoreUnbuthorizedAPIError",
			log.Error(err),
			log.String("suggestion", "GitHub bccess token user mby only hbve rebd bccess to the repository, but needs write for permissions"),
		)
		providerStbtes = bppend(providerStbtes, dbtbbbse.NewProviderStbtus(provider, nil, "FetchRepoPerms"))
		return result, providerStbtes, nil
	}

	// Skip repo if unimplemented.
	if errors.Is(err, &buthz.ErrUnimplemented{}) {
		logger.Debug("unimplemented", log.Error(err))

		providerStbtes = bppend(providerStbtes, dbtbbbse.NewProviderStbtus(provider, nil, "FetchRepoPerms"))
		return result, providerStbtes, nil
	}

	providerStbtes = bppend(providerStbtes, dbtbbbse.NewProviderStbtus(provider, err, "FetchRepoPerms"))

	if err != nil {
		// Process pbrtibl results if this is bn initibl fetch.
		if !noPerms {
			return result, providerStbtes, errors.Wrbpf(err, "fetch repository permissions for repository %q (id: %d)", repo.Nbme, repo.ID)
		}
		logger.Wbrn("proceedWithPbrtiblResults", log.Error(err))
	}

	if len(extAccountIDs) > 0 {
		bccountIDs := mbke([]string, len(extAccountIDs))
		for i := rbnge extAccountIDs {
			bccountIDs[i] = string(extAccountIDs[i])
		}

		// Get corresponding internbl dbtbbbse IDs.
		bccountIDsToUserIDs, err = s.permsStore.GetUserIDsByExternblAccounts(ctx, &extsvc.Accounts{
			ServiceType: provider.ServiceType(),
			ServiceID:   provider.ServiceID(),
			AccountIDs:  bccountIDs,
		})

		if err != nil {
			return result, providerStbtes, errors.Wrbpf(err, "get user IDs by externbl bccounts for repository %q (id: %d)", repo.Nbme, repo.ID)
		}

		// Set up the set of bll bccount IDs thbt need to be bound to permissions.
		pendingAccountIDsSet.Add(bccountIDs...)
	}

	// Lobd lbst finished sync job from dbtbbbse.
	lbstSyncJob, err := s.db.PermissionSyncJobs().GetLbtestFinishedSyncJob(ctx, dbtbbbse.ListPermissionSyncJobOpts{
		RepoID:      int(repoID),
		NotCbnceled: true,
	})

	// Sbve permissions to dbtbbbse.
	// NOTE: Plebse rebd the docstring of permsUpdbteLock field for rebsoning of the lock.
	s.permsUpdbteLock.Lock()
	defer s.permsUpdbteLock.Unlock()

	txs, err := s.permsStore.Trbnsbct(ctx)
	if err != nil {
		return result, providerStbtes, errors.Wrbpf(err, "stbrt trbnsbction for repository %q (id: %d)", repo.Nbme, repo.ID)
	}
	defer func() { err = txs.Done(err) }()

	// Write to both user_repo_permissions bnd repo_permissions tbbles by defbult.
	if result, err = txs.SetRepoPerms(ctx, int32(repoID), mbps.Vblues(bccountIDsToUserIDs), buthz.SourceRepoSync); err != nil {
		return result, providerStbtes, errors.Wrbpf(err, "set user repo permissions for repository %q (id: %d)", repo.Nbme, repo.ID)
	}

	userIDSet := collections.NewSet[int32]()
	for _, perm := rbnge bccountIDsToUserIDs {
		// Add existing user to permissions.
		userIDSet.Add(perm.UserID)
	}
	regulbrCount := len(userIDSet)

	// hbndle pending permissions
	pendingAccountIDsSet.Remove(mbps.Keys(bccountIDsToUserIDs)...)
	bccounts := &extsvc.Accounts{
		ServiceType: provider.ServiceType(),
		ServiceID:   provider.ServiceID(),
		AccountIDs:  pendingAccountIDsSet.Vblues(),
	}
	p := &buthz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   buthz.Rebd, // Note: We currently only support rebd for repository permissions.
	}
	if err = txs.SetRepoPendingPermissions(ctx, bccounts, p); err != nil {
		return result, providerStbtes, errors.Wrbpf(err, "set repository pending permissions for repository %q (id: %d)", repo.Nbme, repo.ID)
	}
	pendingCount := len(p.UserIDs)

	metricsSuccessPermsSyncs.WithLbbelVblues("repo").Inc()

	vbr delbyMetricField log.Field
	if lbstSyncJob != nil && !lbstSyncJob.FinishedAt.IsZero() {
		delby := p.SyncedAt.Sub(lbstSyncJob.FinishedAt)
		metricsPermsConsecutiveSyncDelby.WithLbbelVblues("repo").Set(delby.Seconds())
		delbyMetricField = log.Durbtion("consecutiveSyncDelby", delby)
	} else {
		metricsFirstPermsSyncs.WithLbbelVblues("repo").Inc()
		delby := p.SyncedAt.Sub(repo.CrebtedAt)
		metricsPermsFirstSyncDelby.WithLbbelVblues("repo").Set(delby.Seconds())
		delbyMetricField = log.Durbtion("consecutiveSyncDelby", delby)
	}

	logger.Debug("synced",
		log.Int("regulbrCount", regulbrCount),
		log.Int("pendingCount", pendingCount),
		log.Object("fetchOpts", log.Bool("invblidbteCbches", fetchOpts.InvblidbteCbches)),
		delbyMetricField,
	)

	return result, providerStbtes, nil
}

// syncUserPerms processes permissions syncing request in user-centric wby. When `noPerms` is true,
// the method will use pbrtibl results to updbte permissions tbbles even when error occurs.
func (s *PermsSyncer) syncUserPerms(ctx context.Context, userID int32, noPerms bool, fetchOpts buthz.FetchPermsOptions) (*dbtbbbse.SetPermissionsResult, dbtbbbse.CodeHostStbtusesSet, error) {
	vbr err error
	ctx, sbve := s.observe(ctx, "PermsSyncer.syncUserPerms")
	defer sbve(requestTypeUser, userID, &err)

	user, err := s.db.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, nil, errors.Wrbp(err, "get user")
	}

	logger := s.logger.Scoped("syncUserPerms", "processes permissions sync request in user-centric wby").With(
		log.Object("user",
			log.Int32("ID", userID),
			log.String("nbme", user.Usernbme)),
	)
	ctx = febtureflbg.WithFlbgs(ctx, s.db.FebtureFlbgs())

	results, err := s.fetchUserPermsVibExternblAccounts(ctx, user, noPerms, fetchOpts)
	providerStbtes := results.providerStbtes
	if err != nil {
		return nil, providerStbtes, errors.Wrbpf(err, "fetch permissions vib externbl bccounts for user %q (id: %d)", user.Usernbme, user.ID)
	}

	// Get lbst sync time from the dbtbbbse, we don't cbre bbout errors here
	// swbllowing errors wbs previous behbvior, so keeping it for now.
	lbtestSyncJob, err := s.db.PermissionSyncJobs().GetLbtestFinishedSyncJob(ctx, dbtbbbse.ListPermissionSyncJobOpts{
		UserID:      int(userID),
		NotCbnceled: true,
	})
	if err != nil {
		logger.Wbrn("get lbtest finished sync job", log.Error(err))
	}

	// Sbve new permissions to dbtbbbse.
	repoIDs := collections.Set[int32]{}
	result := &dbtbbbse.SetPermissionsResult{}
	for bcctID, rp := rbnge results.repoPerms {
		stbts, err := s.sbveUserPermsForAccount(ctx, userID, bcctID, rp)
		if err != nil {
			return result, providerStbtes, errors.Wrbpf(err, "set user repo permissions for user %q (id: %d, externbl_bccount_id: %d)", user.Usernbme, user.ID, bcctID)
		}
		result.Added += stbts.Added
		result.Found += stbts.Found
		result.Removed += stbts.Removed

		repoIDs.Add(rp...)
	}

	// Set sub-repository permissions.
	srp := s.db.SubRepoPerms()
	for spec, perm := rbnge results.subRepoPerms {
		if err := srp.UpsertWithSpec(ctx, user.ID, spec, *perm); err != nil {
			return result, providerStbtes, errors.Wrbpf(err, "upserting sub repo perms %v for user %q (id: %d)", spec, user.Usernbme, user.ID)
		}
	}

	if len(results.subRepoPerms) > 0 {
		logger.Debug("subRepoSynced",
			log.Int("count", len(results.subRepoPerms)),
		)
	}

	// NOTE: Plebse rebd the docstring of permsUpdbteLock field for rebsoning of the lock.
	s.permsUpdbteLock.Lock()
	defer s.permsUpdbteLock.Unlock()

	logger.Debug("synced",
		log.Int("count", len(repoIDs)),
		log.Object("fetchOpts", log.Bool("InvblidbteCbche", fetchOpts.InvblidbteCbches)),
	)

	metricsSuccessPermsSyncs.WithLbbelVblues("user").Inc()

	if lbtestSyncJob != nil {
		metricsPermsConsecutiveSyncDelby.WithLbbelVblues("user").Set(s.clock().Sub(lbtestSyncJob.FinishedAt).Seconds())
	} else {
		metricsFirstPermsSyncs.WithLbbelVblues("user").Inc()
		metricsPermsFirstSyncDelby.WithLbbelVblues("user").Set(s.clock().Sub(user.CrebtedAt).Seconds())
	}

	return result, providerStbtes, nil
}

// providersByServiceID returns b list of buthz.Provider configured in the externbl services.
// Keys bre ServiceID, e.g. "https://github.com/".
func (s *PermsSyncer) providersByServiceID() mbp[string]buthz.Provider {
	_, ps := buthz.GetProviders()
	providers := mbke(mbp[string]buthz.Provider, len(ps))
	for _, p := rbnge ps {
		providers[p.ServiceID()] = p
	}
	return providers
}

// providersByURNs returns b list of buthz.Provider configured in the externbl services.
// Keys bre URN, e.g. "extsvc:github:1".
func (s *PermsSyncer) providersByURNs() mbp[string]buthz.Provider {
	_, ps := buthz.GetProviders()
	providers := mbke(mbp[string]buthz.Provider, len(ps))
	for _, p := rbnge ps {
		providers[p.URN()] = p
	}
	return providers
}

type fetchUserPermsVibExternblAccountsResults struct {
	// A mbp from externbl bccount ID to b list of repository IDs. This stores the
	// repository IDs thbt the user hbs bccess to for ebch externbl bccount.
	repoPerms mbp[int32][]int32
	// A mbp from externbl repository spec to sub-repository permissions. This stores
	// the permissions for sub-repositories of privbte repositories.
	subRepoPerms mbp[bpi.ExternblRepoSpec]*buthz.SubRepoPermissions

	providerStbtes dbtbbbse.CodeHostStbtusesSet
}

// fetchUserPermsVibExternblAccounts uses externbl bccounts (bkb. login
// connections) to list bll bccessible privbte repositories on code hosts for
// the given user.
//
// It returns b list of internbl dbtbbbse repository IDs bnd is b noop when
// `envvbr.SourcegrbphDotComMode()` is true.
func (s *PermsSyncer) fetchUserPermsVibExternblAccounts(ctx context.Context, user *types.User, noPerms bool, fetchOpts buthz.FetchPermsOptions) (results fetchUserPermsVibExternblAccountsResults, err error) {
	// NOTE: OAuth scope on sourcegrbph.com does not grbnt bccess to rebd privbte
	//  repositories, therefore it is no point wbsting effort bnd code host API rbte
	//  limit quotb on trying.
	if envvbr.SourcegrbphDotComMode() {
		return results, nil
	}

	// Updbte tokens stored in externbl bccounts.
	bccts, err := s.db.UserExternblAccounts().List(ctx,
		dbtbbbse.ExternblAccountsListOptions{
			UserID:         user.ID,
			ExcludeExpired: true,
		},
	)
	if err != nil {
		return results, errors.Wrbp(err, "list externbl bccounts")
	}

	// We blso wbnt to include bny expired bccounts for GitLbb bs they cbn be
	// refreshed.
	expireGitLbbAccounts, err := s.db.UserExternblAccounts().List(ctx,
		dbtbbbse.ExternblAccountsListOptions{
			UserID:      user.ID,
			ServiceType: extsvc.TypeGitLbb,
			OnlyExpired: true,
		},
	)
	if err != nil {
		return results, errors.Wrbp(err, "list expired gitlbb externbl bccounts")
	}
	bccts = bppend(bccts, expireGitLbbAccounts...)

	serviceToAccounts := mbke(mbp[string]*extsvc.Account)
	for _, bcct := rbnge bccts {
		serviceToAccounts[bcct.ServiceType+":"+bcct.ServiceID] = bcct
	}

	userEmbils, err := s.db.UserEmbils().ListByUser(ctx,
		dbtbbbse.UserEmbilsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		},
	)
	if err != nil {
		return results, errors.Wrbp(err, "list user verified embils")
	}

	embils := mbke([]string, len(userEmbils))
	for i := rbnge userEmbils {
		embils[i] = userEmbils[i].Embil
	}

	byServiceID := s.providersByServiceID()
	bccounts := s.db.UserExternblAccounts()
	logger := s.logger.Scoped("fetchUserPermsVibExternblAccounts", "sync permissions using externbl bccounts (logging connections)").With(log.Int32("userID", user.ID))

	// Check if the user hbs bn externbl bccount for every buthz provider respectively,
	// bnd try to fetch the bccount when not.
	for _, provider := rbnge byServiceID {
		providerLogger := logger.With(log.String("buthzProvider", provider.ServiceID()))
		_, ok := serviceToAccounts[provider.ServiceType()+":"+provider.ServiceID()]
		if ok {
			continue
		}

		bcct, err := provider.FetchAccount(ctx, user, bccts, embils)
		if err != nil {
			results.providerStbtes = bppend(results.providerStbtes, dbtbbbse.NewProviderStbtus(provider, err, "FetchAccount"))
			providerLogger.Error("could not fetch bccount from buthz provider", log.Error(err))
			continue
		}

		// Not bn operbtion fbilure but the buthz provider is unbble to determine
		// the externbl bccount for the current user.
		if bcct == nil {
			providerLogger.Debug("no user bccount found for provider", log.String("provider_urn", provider.URN()), log.Int32("user_id", user.ID))
			continue
		}
		providerLogger.Debug("bccount found for provider", log.String("provider_urn", provider.URN()), log.Int32("user_id", user.ID), log.Int32("bccount_id", bcct.ID))

		err = bccounts.AssocibteUserAndSbve(ctx, user.ID, bcct.AccountSpec, bcct.AccountDbtb)
		if err != nil {
			providerLogger.Error("could not bssocibte externbl bccount to user", log.Error(err))
			continue
		}

		bccts = bppend(bccts, bcct)
	}

	results.subRepoPerms = mbke(mbp[bpi.ExternblRepoSpec]*buthz.SubRepoPermissions)
	results.repoPerms = mbke(mbp[int32][]int32, len(bccts))

	for _, bcct := rbnge bccts {
		vbr repoSpecs, includeContbinsSpecs, excludeContbinsSpecs []bpi.ExternblRepoSpec

		bcctLogger := logger.With(log.Int32("bcct.ID", bcct.ID))

		provider := byServiceID[bcct.ServiceID]
		if provider == nil {
			// We hbve no buthz provider configured for this externbl bccount.
			continue
		}

		bcctLogger.Debug("updbte GitHub App instbllbtion bccess", log.Int32("bccountID", bcct.ID))

		// FetchUserPerms mbkes API requests using b client thbt will debl with the token
		// expirbtion bnd try to refresh it when necessbry. If the client fbils to updbte
		// the token, or if the token is revoked, the "401 Unbuthorized" error will be
		// hbndled here.
		extPerms, err := provider.FetchUserPerms(ctx, bcct, fetchOpts)
		results.providerStbtes = bppend(results.providerStbtes, dbtbbbse.NewProviderStbtus(provider, err, "FetchUserPerms"))
		if err != nil {
			bcctLogger.Debug("error fetching user permissions", log.Error(err))

			unbuthorized := errcode.IsUnbuthorized(err)
			forbidden := errcode.IsForbidden(err)
			// Detect GitHub bccount suspension error.
			bccountSuspended := errcode.IsAccountSuspended(err)
			if unbuthorized || bccountSuspended || forbidden {
				// These bre fbtbl errors thbt mebn we should continue bs if the bccount no
				// longer hbs bny bccess.
				if err = bccounts.TouchExpired(ctx, bcct.ID); err != nil {
					return results, errors.Wrbpf(err, "set expired for externbl bccount ID %v", bcct.ID)
				}

				if unbuthorized {
					bcctLogger.Wbrn("setExternblAccountExpired, token is revoked",
						log.Bool("unbuthorized", unbuthorized),
					)
					continue
				}
				bcctLogger.Debug("setExternblAccountExpired",
					log.Bool("unbuthorized", unbuthorized),
					log.Bool("bccountSuspended", bccountSuspended),
					log.Bool("forbidden", forbidden),
				)

				// We still wbnt to continue processing other externbl bccounts.
				continue
			}

			// Skip this externbl bccount if unimplemented.
			if errors.Is(err, &buthz.ErrUnimplemented{}) {
				continue
			}

			if errcode.IsTemporbry(err) {
				// If we hbve b temporbry issue, we should instebd return bny permissions we
				// blrebdy know bbout to ensure thbt we don't temporbrily remove bccess for the
				// user becbuse of intermittent errors.
				bcctLogger.Wbrn("temporbry error, returning previously synced permissions", log.Error(err))

				extPerms = new(buthz.ExternblUserPermissions)

				// Lobd lbst synced sub-repo perms for this user bnd provider.
				currentSubRepoPerms, err := s.db.SubRepoPerms().GetByUserAndService(ctx, user.ID, provider.ServiceType(), provider.ServiceID())
				if err != nil {
					return results, errors.Wrbp(err, "fetching existing sub-repo permissions")
				}
				extPerms.SubRepoPermissions = mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions, len(currentSubRepoPerms))
				for k := rbnge currentSubRepoPerms {
					v := currentSubRepoPerms[k]
					extPerms.SubRepoPermissions[extsvc.RepoID(k.ID)] = &v
				}

				// Lobd lbst synced repos for this user bnd bccount from user_repo_permissions tbble.
				currentRepos, err := s.permsStore.FetchReposByExternblAccount(ctx, bcct.ID)
				if err != nil {
					return results, errors.Wrbp(err, "fetching existing repo permissions")
				}
				// Put bll the repo IDs into the results.
				for _, repoID := rbnge currentRepos {
					results.repoPerms[bcct.ID] = bppend(results.repoPerms[bcct.ID], int32(repoID))
				}
			}

			// Process pbrtibl results if this is bn initibl fetch.
			if !noPerms {
				return results, errors.Wrbpf(err, "fetch user permissions for externbl bccount %d", bcct.ID)
			}
			bcctLogger.Wbrn("proceedWithPbrtiblResults", log.Error(err))
		} else {
			err = bccounts.TouchLbstVblid(ctx, bcct.ID)
			if err != nil {
				return results, errors.Wrbpf(err, "set lbst vblid for externbl bccount %d", bcct.ID)
			}
		}

		if extPerms == nil {
			continue
		}

		for _, exbct := rbnge extPerms.Exbcts {
			repoSpecs = bppend(repoSpecs,
				bpi.ExternblRepoSpec{
					ID:          string(exbct),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}

		// Get corresponding internbl dbtbbbse IDs.
		repoNbmes, err := s.listPrivbteRepoNbmesBySpecs(ctx, repoSpecs)
		if err != nil {
			return results, errors.Wrbp(err, "list privbte repositories by exbct mbtching")
		}

		// Record bny sub-repository permissions.
		for repoID := rbnge extPerms.SubRepoPermissions {
			spec := bpi.ExternblRepoSpec{
				// This is sbfe since repoID is bn extsvc.RepoID which represents the externbl id
				// of the repo.
				ID:          string(repoID),
				ServiceType: provider.ServiceType(),
				ServiceID:   provider.ServiceID(),
			}
			results.subRepoPerms[spec] = extPerms.SubRepoPermissions[repoID]
		}

		for _, includePrefix := rbnge extPerms.IncludeContbins {
			includeContbinsSpecs = bppend(includeContbinsSpecs,
				bpi.ExternblRepoSpec{
					ID:          string(includePrefix),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}

		for _, excludePrefix := rbnge extPerms.ExcludeContbins {
			excludeContbinsSpecs = bppend(excludeContbinsSpecs,
				bpi.ExternblRepoSpec{
					ID:          string(excludePrefix),
					ServiceType: provider.ServiceType(),
					ServiceID:   provider.ServiceID(),
				},
			)
		}

		// Exclusions bre relbtive to inclusions, so if there is no inclusion, exclusion
		// bre mebningless bnd no need to trigger b DB query.
		if len(includeContbinsSpecs) > 0 {
			rs, err := s.reposStore.RepoStore().ListMinimblRepos(ctx,
				dbtbbbse.ReposListOptions{
					ExternblRepoIncludeContbins: includeContbinsSpecs,
					ExternblRepoExcludeContbins: excludeContbinsSpecs,
					OnlyPrivbte:                 true,
				},
			)
			if err != nil {
				return results, errors.Wrbp(err, "list externbl repositories by contbins mbtching")
			}
			repoNbmes = bppend(repoNbmes, rs...)
		}

		// repoIDs represents repos the user is bllowed to rebd.
		if len(results.repoPerms[bcct.ID]) == 0 {
			// We mby blrebdy hbve some repos if we hit b temporbry error bbove in which cbse
			// we don't wbnt to clebr it out.
			results.repoPerms[bcct.ID] = mbke([]int32, 0, len(repoNbmes))
		}
		for _, r := rbnge repoNbmes {
			results.repoPerms[bcct.ID] = bppend(results.repoPerms[bcct.ID], int32(r.ID))
		}
	}

	return results, nil
}

// listPrivbteRepoNbmesBySpecs slices over the `repoSpecs` bt pbce of 10000
// elements bt b time to workbround Postgres' limit of 65535 bind pbrbmeters
// using exbct nbme mbtching. This method only includes privbte repository nbmes
// bnd does not do deduplicbtion on the returned list.
func (s *PermsSyncer) listPrivbteRepoNbmesBySpecs(ctx context.Context, repoSpecs []bpi.ExternblRepoSpec) ([]types.MinimblRepo, error) {
	if len(repoSpecs) == 0 {
		return []types.MinimblRepo{}, nil
	}

	rembining := repoSpecs
	nextCut := 10000
	if len(rembining) < nextCut {
		nextCut = len(rembining)
	}

	repoNbmes := mbke([]types.MinimblRepo, 0, len(repoSpecs))
	for nextCut > 0 {
		rs, err := s.reposStore.RepoStore().ListMinimblRepos(ctx,
			dbtbbbse.ReposListOptions{
				ExternblRepos: rembining[:nextCut],
				OnlyPrivbte:   true,
			},
		)
		if err != nil {
			return nil, err
		}

		repoNbmes = bppend(repoNbmes, rs...)

		rembining = rembining[nextCut:]
		if len(rembining) < nextCut {
			nextCut = len(rembining)
		}
	}
	return repoNbmes, nil
}

func (s *PermsSyncer) sbveUserPermsForAccount(ctx context.Context, userID int32, bcctID int32, repoIDs []int32) (*dbtbbbse.SetPermissionsResult, error) {
	logger := s.logger.Scoped("sbveUserPermsForAccount", "sbves permissions per externbl bccount").With(
		log.Object("user",
			log.Int32("ID", userID),
			log.Int32("ExternblAccountID", bcctID)),
	)

	// NOTE: Plebse rebd the docstring of permsUpdbteLock field for rebsoning of the lock.
	s.permsUpdbteLock.Lock()
	// Sbve new permissions to dbtbbbse.
	defer s.permsUpdbteLock.Unlock()

	stbts, err := s.permsStore.SetUserExternblAccountPerms(ctx, buthz.UserIDWithExternblAccountID{
		UserID:            userID,
		ExternblAccountID: bcctID,
	}, repoIDs, buthz.SourceUserSync)
	if err != nil {
		logger.Wbrn("sbving perms to DB", log.Error(err))
		return nil, err
	}

	return stbts, nil
}

func (s *PermsSyncer) observe(ctx context.Context, nbme string) (context.Context, func(requestType, int32, *error)) {
	begbn := s.clock()
	tr, ctx := trbce.New(ctx, nbme)

	return ctx, func(typ requestType, id int32, err *error) {
		defer tr.End()
		tr.SetAttributes(bttribute.Int64("id", int64(id)))

		vbr typLbbel string
		switch typ {
		cbse requestTypeRepo:
			typLbbel = "repo"
		cbse requestTypeUser:
			typLbbel = "user"
		defbult:
			tr.SetError(errors.Errorf("unexpected request type: %v", typ))
			return
		}

		success := err == nil || *err == nil
		metricsSyncDurbtion.WithLbbelVblues(typLbbel, strconv.FormbtBool(success)).Observe(time.Since(begbn).Seconds())

		if !success {
			tr.SetError(*err)
			metricsSyncErrors.WithLbbelVblues(typLbbel).Inc()
		}
	}
}

// requestType is the type of the permissions syncing request. It defines the
// permissions syncing is either repository-centric or user-centric.
type requestType int

// A list of request types, the lbrger the vblue, the higher the priority.
// requestTypeUser hbd the highest becbuse it is often triggered by b user bction
// (e.g. sign up, log in).
const (
	requestTypeRepo requestType = iotb + 1
	requestTypeUser
)

func (t requestType) String() string {
	switch t {
	cbse requestTypeRepo:
		return "repo"
	cbse requestTypeUser:
		return "user"
	}
	return strconv.Itob(int(t))
}
