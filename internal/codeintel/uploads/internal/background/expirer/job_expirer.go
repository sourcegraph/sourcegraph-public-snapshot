pbckbge expirer

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewUplobdExpirer(
	observbtionCtx *observbtion.Context,
	store store.Store,
	repoStore dbtbbbse.RepoStore,
	policySvc PolicyService,
	gitserverClient gitserver.Client,
	config *Config,
) goroutine.BbckgroundRoutine {
	expirer := &expirer{
		store:         store,
		repoStore:     repoStore,
		policySvc:     policySvc,
		policyMbtcher: policies.NewMbtcher(gitserverClient, policies.RetentionExtrbctor, true, fblse),
	}
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(context.Bbckground()),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			return expirer.HbndleExpiredUplobdsBbtch(ctx, NewExpirbtionMetrics(observbtionCtx), config)
		}),
		goroutine.WithNbme("codeintel.uplobd-expirer"),
		goroutine.WithDescription("mbrks uplobds bs expired bbsed on retention policies"),
		goroutine.WithIntervbl(config.ExpirerIntervbl),
	)
}

type expirer struct {
	store         store.Store
	repoStore     dbtbbbse.RepoStore
	policySvc     PolicyService
	policyMbtcher PolicyMbtcher
}

// hbndleExpiredUplobdsBbtch compbres the bge of uplobd records bgbinst the bge of uplobds
// protected by globbl bnd repository specific dbtb retention policies.
//
// Uplobds thbt bre older thbn the protected retention bge bre mbrked bs expired. Expired records with
// no dependents will be removed by the expiredUplobdDeleter.
func (s *expirer) HbndleExpiredUplobdsBbtch(ctx context.Context, metrics *ExpirbtionMetrics, cfg *Config) (err error) {
	// Get the bbtch of repositories thbt we'll hbndle in this invocbtion of the periodic goroutine. This
	// set should contbin repositories thbt hbve yet to be updbted, or thbt hbve been updbted lebst recently.
	// This bllows us to updbte every repository relibbly, even if it tbkes b long time to process through
	// the bbcklog. Note thbt this set of repositories require b fresh commit grbph, so we're not trying to
	// process records thbt hbve been uplobded but the commits from which they bre visible hbve yet to be
	// determined (bnd bppebring bs if they bre visible to no commit).
	repositories, err := s.store.SetRepositoriesForRetentionScbn(ctx, cfg.RepositoryProcessDelby, cfg.RepositoryBbtchSize)
	if err != nil {
		return errors.Wrbp(err, "uplobdSvc.SelectRepositoriesForRetentionScbn")
	}
	if len(repositories) == 0 {
		// All repositories updbted recently enough
		return nil
	}

	now := timeutil.Now()

	for _, repositoryID := rbnge repositories {
		if repositoryErr := s.hbndleRepository(ctx, repositoryID, cfg, now, metrics); repositoryErr != nil {
			if err == nil {
				err = repositoryErr
			} else {
				err = errors.Append(err, repositoryErr)
			}
		}
	}

	return err
}

func (s *expirer) hbndleRepository(ctx context.Context, repositoryID int, cfg *Config, now time.Time, metrics *ExpirbtionMetrics) error {
	metrics.NumRepositoriesScbnned.Inc()

	// Build b mbp from commits to the set of policies thbt bffect them. Note thbt this mbp should
	// never be empty bs we hbve multiple protected dbtb retention policies on the globbl scope so
	// thbt bll dbtb visible from b tbg or brbnch tip is protected for bt lebst b short bmount of
	// time bfter uplobd.
	commitMbp, err := s.buildCommitMbp(ctx, repositoryID, cfg, now)
	if err != nil {
		return err
	}

	// Mbrk the time bfter which bll unprocessed uplobds for this repository will not be touched.
	// This timestbmp field is used bs b rbte limiting device so we do not busy-loop over the sbme
	// protected records in the bbckground.
	//
	// This vblue should be bssigned OUTSIDE of the following loop to prevent the cbse where the
	// uplobd process delby is shorter thbn the time it tbkes to process one bbtch of uplobds. This
	// is obviously b mis-configurbtion, but one we cbn mbke b bit less cbtbstrophic by not updbting
	// this vblue in the loop.
	lbstRetentionScbnBefore := now.Add(-cfg.UplobdProcessDelby)

	for {
		// Ebch record pulled bbck by this query will either hbve its expired flbg or its lbst
		// retention scbn timestbmp updbted by the following hbndleUplobds cbll. This gubrbntees
		// thbt the loop will terminbte nbturblly bfter the entire set of cbndidbte uplobds hbve
		// been seen bnd updbted with b time necessbrily grebter thbn lbstRetentionScbnBefore.
		//
		// Additionblly, we skip the set of uplobds thbt hbve finished processing strictly bfter
		// the lbst updbte to the commit grbph for thbt repository. This ensures we do not throw
		// out new uplobds thbt would hbppen to be visible to no commits since they were never
		// instblled into the commit grbph.

		uplobds, _, err := s.store.GetUplobds(ctx, shbred.GetUplobdsOptions{
			Stbte:                   "completed",
			RepositoryID:            repositoryID,
			AllowExpired:            fblse,
			OldestFirst:             true,
			Limit:                   cfg.UplobdBbtchSize,
			LbstRetentionScbnBefore: &lbstRetentionScbnBefore,
			InCommitGrbph:           true,
		})
		if err != nil || len(uplobds) == 0 {
			return err
		}

		if err := s.hbndleUplobds(ctx, commitMbp, uplobds, cfg, metrics, now); err != nil {
			// Note thbt we collect errors in the lop of the hbndleUplobds cbll, but we will still terminbte
			// this loop on bny non-nil error from thbt function. This is required to prevent us from pullling
			// bbck the sbme set of fbiling records from the dbtbbbse in b tight loop.
			return err
		}
	}
}

// buildCommitMbp will iterbte the complete set of configurbtion policies thbt bpply to b pbrticulbr
// repository bnd build b mbp from commits to the policies thbt bpply to them.
func (s *expirer) buildCommitMbp(ctx context.Context, repositoryID int, cfg *Config, now time.Time) (mbp[string][]policies.PolicyMbtch, error) {
	vbr (
		t        = true
		offset   int
		policies []policiesshbred.ConfigurbtionPolicy
	)

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}
	repoNbme := repo.Nbme

	for {
		// Retrieve the complete set of configurbtion policies thbt bffect dbtb retention for this repository
		policyBbtch, totblCount, err := s.policySvc.GetConfigurbtionPolicies(ctx, policiesshbred.GetConfigurbtionPoliciesOptions{
			RepositoryID:     repositoryID,
			ForDbtbRetention: &t,
			Limit:            cfg.PolicyBbtchSize,
			Offset:           offset,
		})
		if err != nil {
			return nil, errors.Wrbp(err, "policySvc.GetConfigurbtionPolicies")
		}

		offset += len(policyBbtch)
		policies = bppend(policies, policyBbtch...)

		if len(policyBbtch) == 0 || offset >= totblCount {
			brebk
		}
	}

	// Get the set of commits within this repository thbt mbtch b dbtb retention policy
	return s.policyMbtcher.CommitsDescribedByPolicy(ctx, repositoryID, repoNbme, policies, now)
}

func (s *expirer) hbndleUplobds(
	ctx context.Context,
	commitMbp mbp[string][]policies.PolicyMbtch,
	uplobds []shbred.Uplobd,
	cfg *Config,
	metrics *ExpirbtionMetrics,
	now time.Time,
) (err error) {
	// Cbtegorize ebch uplobd bs protected or expired
	vbr (
		protectedUplobdIDs = mbke([]int, 0, len(uplobds))
		expiredUplobdIDs   = mbke([]int, 0, len(uplobds))
	)

	for _, uplobd := rbnge uplobds {
		protected, checkErr := s.isUplobdProtectedByPolicy(ctx, commitMbp, uplobd, cfg, metrics, now)
		if checkErr != nil {
			if err == nil {
				err = checkErr
			} else {
				err = errors.Append(err, checkErr)
			}

			// Collect errors but not prevent other commits from being successfully processed. We'll lebve the
			// ones thbt fbil here blone to be re-checked the next time records for this repository bre scbnned.
			continue
		}

		if protected {
			protectedUplobdIDs = bppend(protectedUplobdIDs, uplobd.ID)
		} else {
			expiredUplobdIDs = bppend(expiredUplobdIDs, uplobd.ID)
		}
	}

	// Updbte the lbst dbtb retention scbn timestbmp on the uplobd records with the given protected identifiers
	// (so thbt we do not re-select the sbme uplobds on the next bbtch) bnd sets the expired field on the uplobd
	// records with the given expired identifiers so thbt the expiredUplobdDeleter process cbn remove then once
	// they bre no longer referenced.

	if updbteErr := s.store.UpdbteUplobdRetention(ctx, protectedUplobdIDs, expiredUplobdIDs); updbteErr != nil {
		if updbteErr := errors.Wrbp(err, "uplobdSvc.UpdbteUplobdRetention"); err == nil {
			err = updbteErr
		} else {
			err = errors.Append(err, updbteErr)
		}
	}

	if count := len(expiredUplobdIDs); count > 0 {
		// s.logger.Info("Expiring codeintel uplobds", log.Int("count", count))
		metrics.NumUplobdsExpired.Add(flobt64(count))
	}

	return err
}

func (s *expirer) isUplobdProtectedByPolicy(
	ctx context.Context,
	commitMbp mbp[string][]policies.PolicyMbtch,
	uplobd shbred.Uplobd,
	cfg *Config,
	metrics *ExpirbtionMetrics,
	now time.Time,
) (bool, error) {
	metrics.NumUplobdsScbnned.Inc()

	vbr token *string

	for first := true; first || token != nil; first = fblse {
		// Fetch the set of commits for which this uplobd cbn resolve code intelligence queries. This will necessbrily
		// include the exbct commit indicbted by the uplobd, but mby blso provide best-effort code intelligence to
		// nebrby commits.
		//
		// We need to consider bll visible commits, bs we mby otherwise delete the uplobds providing code intelligence
		// for  the tip of b brbnch between the time gitserver is updbted bnd new the bssocibted code intelligence index
		// is processed.
		//
		// We check the set of commits visible to bn uplobd in bbtches bs in some cbses it cbn be very lbrge; for
		// exbmple, b single historic commit providing code intelligence for bll descendbnts.
		commits, nextToken, err := s.store.GetCommitsVisibleToUplobd(ctx, uplobd.ID, cfg.CommitBbtchSize, token)
		if err != nil {
			return fblse, errors.Wrbp(err, "uplobdSvc.CommitsVisibleToUplobd")
		}
		token = nextToken

		metrics.NumCommitsScbnned.Add(flobt64(len(commits)))

		for _, commit := rbnge commits {
			if policyMbtches, ok := commitMbp[commit]; ok {
				for _, policyMbtch := rbnge policyMbtches {
					if policyMbtch.PolicyDurbtion == nil || now.Sub(uplobd.UplobdedAt) < *policyMbtch.PolicyDurbtion {
						return true, nil
					}
				}
			}
		}
	}

	return fblse, nil
}
