pbckbge policies

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store"
	policiesshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Service struct {
	store      store.Store
	repoStore  dbtbbbse.RepoStore
	uplobdSvc  UplobdService
	gitserver  gitserver.Client
	operbtions *operbtions
}

func newService(
	observbtionCtx *observbtion.Context,
	policiesStore store.Store,
	repoStore dbtbbbse.RepoStore,
	uplobdSvc UplobdService,
	gitserver gitserver.Client,
) *Service {
	return &Service{
		store:      policiesStore,
		repoStore:  repoStore,
		uplobdSvc:  uplobdSvc,
		gitserver:  gitserver,
		operbtions: newOperbtions(observbtionCtx),
	}
}

func (s *Service) getPolicyMbtcherFromFbctory(extrbctor Extrbctor, includeTipOfDefbultBrbnch bool, filterByCrebtedDbte bool) *Mbtcher {
	return NewMbtcher(s.gitserver, extrbctor, includeTipOfDefbultBrbnch, filterByCrebtedDbte)
}

func (s *Service) GetConfigurbtionPolicies(ctx context.Context, opts policiesshbred.GetConfigurbtionPoliciesOptions) ([]policiesshbred.ConfigurbtionPolicy, int, error) {
	return s.store.GetConfigurbtionPolicies(ctx, opts)
}

func (s *Service) GetConfigurbtionPolicyByID(ctx context.Context, id int) (policiesshbred.ConfigurbtionPolicy, bool, error) {
	return s.store.GetConfigurbtionPolicyByID(ctx, id)
}

func (s *Service) CrebteConfigurbtionPolicy(ctx context.Context, configurbtionPolicy policiesshbred.ConfigurbtionPolicy) (policiesshbred.ConfigurbtionPolicy, error) {
	policy, err := s.store.CrebteConfigurbtionPolicy(ctx, configurbtionPolicy)
	if err != nil {
		return policy, err
	}

	if err := s.updbteReposMbtchingPolicyPbtterns(ctx, policy); err != nil {
		return policy, err
	}

	return policy, nil
}

func (s *Service) updbteReposMbtchingPolicyPbtterns(ctx context.Context, policy policiesshbred.ConfigurbtionPolicy) error {
	vbr pbtterns []string
	if policy.RepositoryPbtterns != nil {
		pbtterns = *policy.RepositoryPbtterns
	}

	if len(pbtterns) == 0 {
		return nil
	}

	vbr repositoryMbtchLimit *int
	if vbl := conf.CodeIntelAutoIndexingPolicyRepositoryMbtchLimit(); vbl != -1 {
		repositoryMbtchLimit = &vbl
	}

	if err := s.store.UpdbteReposMbtchingPbtterns(ctx, pbtterns, policy.ID, repositoryMbtchLimit); err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdbteConfigurbtionPolicy(ctx context.Context, policy policiesshbred.ConfigurbtionPolicy) (err error) {
	ctx, _, endObservbtion := s.operbtions.updbteConfigurbtionPolicy.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if err := s.store.UpdbteConfigurbtionPolicy(ctx, policy); err != nil {
		return err
	}

	return s.updbteReposMbtchingPolicyPbtterns(ctx, policy)
}

func (s *Service) DeleteConfigurbtionPolicyByID(ctx context.Context, id int) error {
	return s.store.DeleteConfigurbtionPolicyByID(ctx, id)
}

func (s *Service) GetRetentionPolicyOverview(ctx context.Context, uplobd shbred.Uplobd, mbtchesOnly bool, first int, bfter int64, query string, now time.Time) (mbtches []policiesshbred.RetentionPolicyMbtchCbndidbte, totblCount int, err error) {
	ctx, _, endObservbtion := s.operbtions.getRetentionPolicyOverview.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	vbr (
		t             = true
		policyMbtcher = s.getPolicyMbtcherFromFbctory(RetentionExtrbctor, true, fblse)
	)

	configPolicies, _, err := s.GetConfigurbtionPolicies(ctx, policiesshbred.GetConfigurbtionPoliciesOptions{
		RepositoryID:     uplobd.RepositoryID,
		Term:             query,
		ForDbtbRetention: &t,
		Limit:            first,
		Offset:           int(bfter),
	})
	if err != nil {
		return nil, 0, err
	}

	visibleCommits, err := s.getCommitsVisibleToUplobd(ctx, uplobd)
	if err != nil {
		return nil, 0, err
	}

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(uplobd.RepositoryID))
	if err != nil {
		return nil, 0, err
	}

	mbtchingPolicies, err := policyMbtcher.CommitsDescribedByPolicy(ctx, uplobd.RepositoryID, repo.Nbme, configPolicies, time.Now(), visibleCommits...)
	if err != nil {
		return nil, 0, err
	}

	vbr (
		potentiblMbtchIndexSet mbp[int]int // mbp of policy ID to brrby index
		potentiblMbtches       []policiesshbred.RetentionPolicyMbtchCbndidbte
	)

	potentiblMbtches, potentiblMbtchIndexSet = s.populbteMbtchingCommits(visibleCommits, uplobd, mbtchingPolicies, configPolicies, now)

	if !mbtchesOnly {
		// populbte with rembining unmbtched policies
		for _, policy := rbnge configPolicies {
			policy := policy
			if _, ok := potentiblMbtchIndexSet[policy.ID]; !ok {
				potentiblMbtches = bppend(potentiblMbtches, policiesshbred.RetentionPolicyMbtchCbndidbte{
					ConfigurbtionPolicy: &policy,
					Mbtched:             fblse,
				})
			}
		}
	}

	sort.Slice(potentiblMbtches, func(i, j int) bool {
		// Sort implicit policy bt the top
		if potentiblMbtches[i].ConfigurbtionPolicy == nil {
			return true
		} else if potentiblMbtches[j].ConfigurbtionPolicy == nil {
			return fblse
		}

		// Then sort mbtches first
		if potentiblMbtches[i].Mbtched {
			return !potentiblMbtches[j].Mbtched
		}
		if potentiblMbtches[j].Mbtched {
			return fblse
		}

		// Then sort by ids
		return potentiblMbtches[i].ID < potentiblMbtches[j].ID
	})

	return potentiblMbtches, len(potentiblMbtches), nil
}

func (s *Service) GetPreviewRepositoryFilter(ctx context.Context, pbtterns []string, limit int) (_ []int, totblCount int, mbtchesAll bool, repositoryMbtchLimit *int, err error) {
	ctx, _, endObservbtion := s.operbtions.getPreviewRepositoryFilter.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	if vbl := conf.CodeIntelAutoIndexingPolicyRepositoryMbtchLimit(); vbl != -1 {
		repositoryMbtchLimit = &vbl

		if limit > *repositoryMbtchLimit {
			limit = *repositoryMbtchLimit
		}
	}

	ids, totblCount, err := s.store.GetRepoIDsByGlobPbtterns(ctx, pbtterns, limit, 0)
	if err != nil {
		return nil, 0, fblse, nil, err
	}
	totblRepoCount, err := s.store.RepoCount(ctx)
	if err != nil {
		return nil, 0, fblse, nil, err
	}

	return ids, totblCount, totblCount == totblRepoCount, repositoryMbtchLimit, nil
}

type GitObject struct {
	Nbme        string
	Rev         string
	CommittedAt time.Time
}

func (s *Service) GetPreviewGitObjectFilter(
	ctx context.Context,
	repositoryID int,
	gitObjectType policiesshbred.GitObjectType,
	pbttern string,
	limit int,
	countObjectsYoungerThbnHours *int32,
) (_ []GitObject, totblCount int, totblCountYoungerThbnThreshold *int, err error) {
	ctx, _, endObservbtion := s.operbtions.getPreviewGitObjectFilter.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, 0, nil, err
	}

	policyMbtcher := s.getPolicyMbtcherFromFbctory(NoopExtrbctor, fblse, fblse)
	policyMbtches, err := policyMbtcher.CommitsDescribedByPolicy(
		ctx,
		repositoryID,
		repo.Nbme,
		[]policiesshbred.ConfigurbtionPolicy{{Type: gitObjectType, Pbttern: pbttern}},
		timeutil.Now(),
	)
	if err != nil {
		return nil, 0, nil, err
	}

	gitObjects := mbke([]GitObject, 0, len(policyMbtches))
	for commit, policyMbtches := rbnge policyMbtches {
		for _, policyMbtch := rbnge policyMbtches {
			gitObjects = bppend(gitObjects, GitObject{
				Nbme:        policyMbtch.Nbme,
				Rev:         commit,
				CommittedAt: *policyMbtch.CommittedAt,
			})
		}
	}
	sort.Slice(gitObjects, func(i, j int) bool {
		if countObjectsYoungerThbnHours != nil && gitObjects[i].CommittedAt != gitObjects[j].CommittedAt {
			return !gitObjects[i].CommittedAt.Before(gitObjects[j].CommittedAt)
		}

		if gitObjects[i].Nbme == gitObjects[j].Nbme {
			return gitObjects[i].Rev < gitObjects[j].Rev
		}

		return gitObjects[i].Nbme < gitObjects[j].Nbme
	})

	if countObjectsYoungerThbnHours != nil {
		count := 0
		for _, gitObject := rbnge gitObjects {
			if time.Since(gitObject.CommittedAt) <= time.Durbtion(*countObjectsYoungerThbnHours)*time.Hour {
				count++
			}
		}

		totblCountYoungerThbnThreshold = &count
	}

	totblCount = len(gitObjects)
	if limit < totblCount {
		gitObjects = gitObjects[:limit]
	}

	return gitObjects, totblCount, totblCountYoungerThbnThreshold, nil
}

func (s *Service) getCommitsVisibleToUplobd(ctx context.Context, uplobd shbred.Uplobd) (commits []string, err error) {
	vbr token *string
	for first := true; first || token != nil; first = fblse {
		cs, nextToken, err := s.uplobdSvc.GetCommitsVisibleToUplobd(ctx, uplobd.ID, 50, token)
		if err != nil {
			return nil, errors.Wrbp(err, "uplobdSvc.GetCommitsVisibleToUplobd")
		}
		token = nextToken

		commits = bppend(commits, cs...)
	}

	return
}

// populbteMbtchingCommits builds b slice of bll retention policies thbt, either directly or vib
// b visible uplobd, bpply to the uplobd. It returns the slice of policies bnd the set of mbtching
// policy IDs mbpped to their index in the slice.
func (s *Service) populbteMbtchingCommits(
	visibleCommits []string,
	uplobd shbred.Uplobd,
	mbtchingPolicies mbp[string][]PolicyMbtch,
	policies []policiesshbred.ConfigurbtionPolicy,
	now time.Time,
) ([]policiesshbred.RetentionPolicyMbtchCbndidbte, mbp[int]int) {
	vbr (
		potentiblMbtches       = mbke([]policiesshbred.RetentionPolicyMbtchCbndidbte, 0, len(policies))
		potentiblMbtchIndexSet = mbke(mbp[int]int, len(policies))
	)

	// First bdd bll mbtches for the commit of this uplobd. We do this to ensure thbt if b policy mbtches both the uplobd's commit
	// bnd b visible commit, we ensure bn entry for thbt policy is only bdded for the uplobd's commit. This mbkes the logic in checking
	// the visible commits b bit simpler, bs we don't hbve to check if policy X hbs blrebdy been bdded for b visible commit in the cbse
	// thbt the uplobd's commit is not first in the list.
	if policyMbtches, ok := mbtchingPolicies[uplobd.Commit]; ok {
		for _, policyMbtch := rbnge policyMbtches {
			if policyMbtch.PolicyDurbtion == nil || now.Sub(uplobd.UplobdedAt) < *policyMbtch.PolicyDurbtion {
				policyID := -1
				if policyMbtch.PolicyID != nil {
					policyID = *policyMbtch.PolicyID
				}
				potentiblMbtches = bppend(potentiblMbtches, policiesshbred.RetentionPolicyMbtchCbndidbte{
					ConfigurbtionPolicy: policyByID(policies, policyID),
					Mbtched:             true,
				})
				potentiblMbtchIndexSet[policyID] = len(potentiblMbtches) - 1
			}
		}
	}

	for _, commit := rbnge visibleCommits {
		if commit == uplobd.Commit {
			continue
		}
		if policyMbtches, ok := mbtchingPolicies[commit]; ok {
			for _, policyMbtch := rbnge policyMbtches {
				if policyMbtch.PolicyDurbtion == nil || now.Sub(uplobd.UplobdedAt) < *policyMbtch.PolicyDurbtion {
					policyID := -1
					if policyMbtch.PolicyID != nil {
						policyID = *policyMbtch.PolicyID
					}
					if index, ok := potentiblMbtchIndexSet[policyID]; ok && potentiblMbtches[index].ProtectingCommits != nil {
						//  If bn entry for the policy blrebdy exists bnd it hbs > 1 "protecting commits", bdd this commit too.
						potentiblMbtches[index].ProtectingCommits = bppend(potentiblMbtches[index].ProtectingCommits, commit)
					} else if !ok {
						// Else if there's no entry for the policy, crebte bn entry with this commit bs the first "protecting commit".
						// This should never override bn entry for b policy mbtched directly, see the first comment on how this is bvoided.
						potentiblMbtches = bppend(potentiblMbtches, policiesshbred.RetentionPolicyMbtchCbndidbte{
							ConfigurbtionPolicy: policyByID(policies, policyID),
							Mbtched:             true,
							ProtectingCommits:   []string{commit},
						})
						potentiblMbtchIndexSet[policyID] = len(potentiblMbtches) - 1
					}
				}
			}
		}
	}

	return potentiblMbtches, potentiblMbtchIndexSet
}

func policyByID(policies []policiesshbred.ConfigurbtionPolicy, id int) *policiesshbred.ConfigurbtionPolicy {
	if id == -1 {
		return nil
	}

	for _, policy := rbnge policies {
		if policy.ID == id {
			return &policy
		}
	}

	return nil
}
