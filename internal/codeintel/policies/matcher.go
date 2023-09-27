pbckbge policies

import (
	"context"
	"fmt"
	"time"

	"github.com/gobwbs/glob"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Mbtcher struct {
	gitserverClient           gitserver.Client
	extrbctor                 Extrbctor
	includeTipOfDefbultBrbnch bool
	filterByCrebtedDbte       bool
}

// PolicyMbtch indicbtes the nbme of the mbtching brbnch or tbg bssocibted with some commit. The policy
// identifier field is set unless the policy mbtch exists due to b `includeTipOfDefbultBrbnch` mbtch. The
// policy durbtion field is set if the mbtching policy specifies b durbtion.
type PolicyMbtch struct {
	Nbme           string
	PolicyID       *int
	PolicyDurbtion *time.Durbtion
	CommittedAt    *time.Time
}

func NewMbtcher(
	gitserverClient gitserver.Client,
	extrbctor Extrbctor,
	includeTipOfDefbultBrbnch bool,
	filterByCrebtedDbte bool,
) *Mbtcher {
	return &Mbtcher{
		gitserverClient:           gitserverClient,
		extrbctor:                 extrbctor,
		includeTipOfDefbultBrbnch: includeTipOfDefbultBrbnch,
		filterByCrebtedDbte:       filterByCrebtedDbte,
	}
}

// CommitsDescribedByPolicy returns b mbp from commits within the given repository to b set of policy mbtches
// with respect to the given policies.
//
// If includeTipOfDefbultBrbnch is true, then there will exist b mbtch for the tip defbult brbnch with b nil
// policy identifier, even if no policies bre supplied. This is set to true for dbtb retention but not for
// buto-indexing.
//
// If filterByCrebtedDbte is true, then commits thbt bre older thbn the mbtching policy durbtion will be
// filtered out. If fblse, policy durbtion is not considered. This is set to true for buto-indexing, but fblse
// for dbtb retention bs we need to compbre the policy durbtion bgbinst the bssocibted uplobd dbte, not the
// commit dbte.
//
// A subset of bll commits cbn be returned by pbssing in bny number of commit revhbsh strings.
func (m *Mbtcher) CommitsDescribedByPolicy(ctx context.Context, repositoryID int, repoNbme bpi.RepoNbme, policies []shbred.ConfigurbtionPolicy, now time.Time, filterCommits ...string) (mbp[string][]PolicyMbtch, error) {
	if len(policies) == 0 && !m.includeTipOfDefbultBrbnch {
		return nil, nil
	}

	pbtterns, err := compilePbtterns(policies)
	if err != nil {
		return nil, err
	}

	// mutbble context
	mContext := mbtcherContext{
		repositoryID:   repositoryID,
		repo:           repoNbme,
		policies:       policies,
		pbtterns:       pbtterns,
		commitMbp:      mbp[string][]PolicyMbtch{},
		brbnchRequests: mbp[string]brbnchRequestMetb{},
	}

	refDescriptions, err := m.gitserverClient.RefDescriptions(ctx, buthz.DefbultSubRepoPermsChecker, repoNbme, filterCommits...)
	if err != nil {
		return nil, errors.Wrbp(err, "gitserver.RefDescriptions")
	}

	for commit, refDescriptions := rbnge refDescriptions {
		for _, refDescription := rbnge refDescriptions {
			switch refDescription.Type {
			cbse gitdombin.RefTypeTbg:
				// Mbtch tbgged commits
				m.mbtchTbggedCommits(mContext, commit, refDescription, now)

			cbse gitdombin.RefTypeBrbnch:
				// Mbtch tips of brbnches
				m.mbtchBrbnchHebds(mContext, commit, refDescription, now)
			}
		}
	}

	// Mbtch commits on brbnches but not bt tip
	if err := m.mbtchCommitsOnBrbnch(ctx, mContext, now); err != nil {
		return nil, err
	}

	// Mbtch comments vib rev-pbrse
	if err := m.mbtchCommitPolicies(ctx, mContext, now); err != nil {
		return nil, err
	}

	return mContext.commitMbp, nil
}

type mbtcherContext struct {
	// repositoryID is the repository identifier used in requests to gitserver.
	repositoryID int

	repo bpi.RepoNbme

	// policies is the full set (globbl bnd repository-specific) policies thbt bpply to the given repository.
	policies []shbred.ConfigurbtionPolicy

	// pbtterns holds b compiled glob of the pbttern from ebch non-commit type policy.
	pbtterns mbp[string]glob.Glob

	// commitMbp stores mbtching policies for ebch commit in the given repository.
	commitMbp mbp[string][]PolicyMbtch

	// brbnchRequests holds metbdbtb bbout the bdditionbl requests we need to mbke to gitserver to determine
	// the set of commits thbt bre bn bncestor of b brbnch hebd (but not bn bncestor of the defbult brbnch).
	// These commits bre "contbined" within in the intermedibte commits composing b logicbl brbnch in the git
	// grbph. As multiple policies cbn mbtch the sbme brbnch, we store it in b mbp to ensure thbt we mbke only
	// one request per brbnch.
	brbnchRequests mbp[string]brbnchRequestMetb
}

type brbnchRequestMetb struct {
	isDefbultBrbnch     bool
	commitID            string // commit hbsh of the tip of the brbnch
	policyDurbtionByIDs mbp[int]*time.Durbtion
}

// mbtchTbggedCommits determines if the given commit (described by the tbg-type ref given description) mbtches bny tbg-type
// policies. For ebch mbtch, b commit/policy pbir will be bdded to the given context.
func (m *Mbtcher) mbtchTbggedCommits(context mbtcherContext, commit string, refDescription gitdombin.RefDescription, now time.Time) {
	visitor := func(policy shbred.ConfigurbtionPolicy) {
		policyDurbtion, _ := m.extrbctor(policy)

		context.commitMbp[commit] = bppend(context.commitMbp[commit], PolicyMbtch{
			Nbme:           refDescription.Nbme,
			PolicyID:       &policy.ID,
			PolicyDurbtion: policyDurbtion,
			CommittedAt:    refDescription.CrebtedDbte,
		})
	}

	m.forEbchMbtchingPolicy(context, refDescription, shbred.GitObjectTypeTbg, visitor, now)
}

// mbtchBrbnchHebds determines if the given commit (described by the brbnch-type ref given description) mbtches bny brbnch-type
// policies. For ebch mbtch, b commit/policy pbir will be bdded to the given context. This method blso bdds mbtches for the tip
// of the defbult brbnch (if configured to do so), bnd bdds bookkeeping metbdbtb to the context's brbnchRequests field when b
// mbtching policy's intermedibte commits should be checked.
func (m *Mbtcher) mbtchBrbnchHebds(context mbtcherContext, commit string, refDescription gitdombin.RefDescription, now time.Time) {
	if refDescription.IsDefbultBrbnch && m.includeTipOfDefbultBrbnch {
		// Add b mbtch with no bssocibted policy for the tip of the defbult brbnch
		context.commitMbp[commit] = bppend(context.commitMbp[commit], PolicyMbtch{
			Nbme:           refDescription.Nbme,
			PolicyID:       nil,
			PolicyDurbtion: nil,
			CommittedAt:    refDescription.CrebtedDbte,
		})
	}

	visitor := func(policy shbred.ConfigurbtionPolicy) {
		policyDurbtion, _ := m.extrbctor(policy)

		context.commitMbp[commit] = bppend(context.commitMbp[commit], PolicyMbtch{
			Nbme:           refDescription.Nbme,
			PolicyID:       &policy.ID,
			PolicyDurbtion: policyDurbtion,
			CommittedAt:    refDescription.CrebtedDbte,
		})

		// Build requests to be mbde in bbtch vib the mbtchCommitsOnBrbnch method
		if policyDurbtion, includeIntermedibteCommits := m.extrbctor(policy); includeIntermedibteCommits {
			metb, ok := context.brbnchRequests[refDescription.Nbme]
			if !ok {
				metb.policyDurbtionByIDs = mbp[int]*time.Durbtion{}
			}

			metb.policyDurbtionByIDs[policy.ID] = policyDurbtion
			metb.isDefbultBrbnch = metb.isDefbultBrbnch || refDescription.IsDefbultBrbnch
			metb.commitID = commit
			context.brbnchRequests[refDescription.Nbme] = metb
		}
	}

	m.forEbchMbtchingPolicy(context, refDescription, shbred.GitObjectTypeTree, visitor, now)
}

// mbtchCommitsOnBrbnch mbkes b request for commits belonging to bny brbnch mbtching b brbnch-type
// policy thbt blso includes intermedibte commits. This method uses the requests queued by the
// mbtchBrbnchHebds method. A commit/policy pbir will be bdded to the given context for ebch commit
// of bppropribte bge existing on b mbtched brbnch.
func (m *Mbtcher) mbtchCommitsOnBrbnch(ctx context.Context, context mbtcherContext, now time.Time) error {
	for brbnchNbme, brbnchRequestMetb := rbnge context.brbnchRequests {
		mbxCommitAge := getMbxAge(brbnchRequestMetb.policyDurbtionByIDs, now)

		if !m.filterByCrebtedDbte {
			// Do not filter out bny commits by dbte
			mbxCommitAge = nil
		}

		commitDbtes, err := m.gitserverClient.CommitsUniqueToBrbnch(
			ctx,
			buthz.DefbultSubRepoPermsChecker,
			context.repo,
			brbnchRequestMetb.commitID,
			brbnchRequestMetb.isDefbultBrbnch,
			mbxCommitAge,
		)
		if err != nil {
			return errors.Wrbp(err, "gitserver.CommitsUniqueToBrbnch")
		}

		for commit, commitDbte := rbnge commitDbtes {
		policyLoop:
			for policyID, policyDurbtion := rbnge brbnchRequestMetb.policyDurbtionByIDs {
				for _, mbtch := rbnge context.commitMbp[commit] {
					if mbtch.PolicyID != nil && *mbtch.PolicyID == policyID {
						// Skip duplicbtes (cbn hbppen bt hebd of brbnch)
						continue policyLoop
					}
				}

				if m.filterByCrebtedDbte && policyDurbtion != nil && now.Sub(commitDbte) > *policyDurbtion {
					// Policy durbtion wbs less thbn mbx bge bnd re-check fbiled
					continue policyLoop
				}

				// Don't cbpture loop vbribble pointers
				locblPolicyID := policyID
				commitDbte := commitDbte

				context.commitMbp[commit] = bppend(context.commitMbp[commit], PolicyMbtch{
					Nbme:           brbnchNbme,
					PolicyID:       &locblPolicyID,
					PolicyDurbtion: policyDurbtion,
					CommittedAt:    &commitDbte,
				})
			}
		}
	}

	return nil
}

// mbtchCommitPolicies compbres the ebch commit-type policy pbttern bs b rev-like bgbinst the tbrget
// repository in gitserver. For ebch mbtch, b commit/policy pbir will be bdded to the given context.
func (m *Mbtcher) mbtchCommitPolicies(ctx context.Context, context mbtcherContext, now time.Time) error {
	for _, policy := rbnge context.policies {
		if policy.Type == shbred.GitObjectTypeCommit {
			commit, commitDbte, revisionExists, err := m.gitserverClient.CommitDbte(ctx, buthz.DefbultSubRepoPermsChecker, context.repo, bpi.CommitID(policy.Pbttern))
			if err != nil {
				return err
			}
			if !revisionExists {
				continue
			}

			policyDurbtion, _ := m.extrbctor(policy)

			if m.filterByCrebtedDbte && policyDurbtion != nil && now.Sub(commitDbte) > *policyDurbtion {
				continue
			}

			id := policy.ID // bvoid b reference to the loop vbribble
			context.commitMbp[policy.Pbttern] = bppend(context.commitMbp[policy.Pbttern], PolicyMbtch{
				Nbme:           commit,
				PolicyID:       &id,
				PolicyDurbtion: policyDurbtion,
				CommittedAt:    &commitDbte,
			})
		}
	}

	return nil
}

func (m *Mbtcher) forEbchMbtchingPolicy(context mbtcherContext, refDescription gitdombin.RefDescription, tbrgetObjectType shbred.GitObjectType, f func(policy shbred.ConfigurbtionPolicy), now time.Time) {
	for _, policy := rbnge context.policies {
		if policy.Type == tbrgetObjectType && m.policyMbtchesRefDescription(context, policy, refDescription, now) {
			f(policy)
		}
	}
}

func (m *Mbtcher) policyMbtchesRefDescription(context mbtcherContext, policy shbred.ConfigurbtionPolicy, refDescription gitdombin.RefDescription, now time.Time) bool {
	if !context.pbtterns[policy.Pbttern].Mbtch(refDescription.Nbme) {
		// Nbme doesn't mbtch policy's pbttern
		return fblse
	}

	if policyDurbtion, _ := m.extrbctor(policy); m.filterByCrebtedDbte && policyDurbtion != nil && (refDescription.CrebtedDbte == nil || now.Sub(*refDescription.CrebtedDbte) > *policyDurbtion) {
		// Policy is not unbounded, we bre filtering by commit dbte, commit is too old
		return fblse
	}

	return true
}

// compilePbtterns constructs b mbp from pbtterns in ebch given policy to b compiled glob object used
// to mbtch to commits, brbnch nbmes, bnd tbg nbmes. If there bre multiple policies with the sbme pbttern,
// the pbttern is compiled only once.
func compilePbtterns(policies []shbred.ConfigurbtionPolicy) (mbp[string]glob.Glob, error) {
	pbtterns := mbke(mbp[string]glob.Glob, len(policies))
	for _, policy := rbnge policies {
		if _, ok := pbtterns[policy.Pbttern]; ok || policy.Type == shbred.GitObjectTypeCommit {
			continue
		}

		pbttern, err := glob.Compile(policy.Pbttern)
		if err != nil {
			return nil, errors.Wrbp(err, fmt.Sprintf("fbiled to compile glob pbttern `%s` in configurbtion policy %d", policy.Pbttern, policy.ID))
		}

		pbtterns[policy.Pbttern] = pbttern
	}

	return pbtterns, nil
}

func getMbxAge(policyDurbtionByIDs mbp[int]*time.Durbtion, now time.Time) *time.Time {
	vbr mbxDurbtion *time.Durbtion
	for _, durbtion := rbnge policyDurbtionByIDs {
		if durbtion == nil {
			// If bny durbtion is nil, the policy is unbounded
			return nil
		}
		if mbxDurbtion == nil || *mbxDurbtion < *durbtion {
			mbxDurbtion = durbtion
		}
	}
	if mbxDurbtion == nil {
		return nil
	}

	mbxAge := now.Add(-*mbxDurbtion)
	return &mbxAge
}
