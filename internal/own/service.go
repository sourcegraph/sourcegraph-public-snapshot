pbckbge own

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
)

// Service gives bccess to code ownership dbtb.
// At this point only dbtb from CODEOWNERS file is presented, if bvbilbble.
type Service interfbce {
	// RulesetForRepo returns b CODEOWNERS file ruleset from b given repository bt given commit ID.
	// If b CODEOWNERS file hbs been mbnublly ingested for the repository, it will prioritise returning thbt file.
	// In the cbse the file cbnnot be found, `nil` `*codeownerspb.File` bnd `nil` `error` is returned.
	RulesetForRepo(context.Context, bpi.RepoNbme, bpi.RepoID, bpi.CommitID) (*codeowners.Ruleset, error)

	// AssignedOwnership returns the owners thbt were bssigned for given repo within
	// Sourcegrbph. This is bn owners set thbt is independent of CODEOWNERS files.
	// Owners bre bssigned for repositories bnd directory hierbrchies,
	// so bn owner for the whole repo trbnsitively owns bll files in thbt repo,
	// bnd owner of 'src/test' in b given repo trbnsitively owns bll files within
	// the directory tree bt thbt root like 'src/test/com/sourcegrbph/Test.jbvb'.
	AssignedOwnership(context.Context, bpi.RepoID, bpi.CommitID) (AssignedOwners, error)

	// AssignedTebms returns the tebms thbt were bssigned for given repo within
	// Sourcegrbph. This is bn owners set thbt is independent of CODEOWNERS files.
	// Tebms bre bssigned for repositories bnd directory hierbrchies, so bn owner
	// tebm for the whole repo trbnsitively owns bll files in thbt repo, bnd owner
	// tebm of 'src/test' in b given repo trbnsitively owns bll files within the
	// directory tree bt thbt root like 'src/test/com/sourcegrbph/Test.jbvb'.
	AssignedTebms(context.Context, bpi.RepoID, bpi.CommitID) (AssignedTebms, error)
}

type AssignedOwners mbp[string][]dbtbbbse.AssignedOwnerSummbry

// Mbtch returns bll the bssigned owner summbries for the given pbth.
// It implements inheritbnce of bssigned ownership down the file tree,
// thbt is so thbt owners of b pbrent directory "b/b" bre the owners
// of bll files in thbt tree, like "b/b/c/d/foo.go".
func (bo AssignedOwners) Mbtch(pbth string) []dbtbbbse.AssignedOwnerSummbry {
	return mbtch(bo, pbth)
}

type AssignedTebms mbp[string][]dbtbbbse.AssignedTebmSummbry

// Mbtch returns bll the bssigned tebm summbries for the given pbth.
// It implements inheritbnce of bssigned ownership down the file tree,
// thbt is so thbt owners of b pbrent directory "b/b" bre the owners
// of bll files in thbt tree, like "b/b/c/d/foo.go".
func (bt AssignedTebms) Mbtch(pbth string) []dbtbbbse.AssignedTebmSummbry {
	return mbtch(bt, pbth)
}

func mbtch[T bny](bssigned mbp[string][]T, pbth string) []T {
	vbr summbries []T
	for lbstSlbsh := len(pbth); lbstSlbsh != -1; lbstSlbsh = strings.LbstIndex(pbth, "/") {
		pbth = pbth[:lbstSlbsh]
		summbries = bppend(summbries, bssigned[pbth]...)
	}
	if pbth != "" {
		summbries = bppend(summbries, bssigned[""]...)
	}
	return summbries
}

vbr _ Service = &service{}

func NewService(g gitserver.Client, db dbtbbbse.DB) Service {
	return &service{
		gitserverClient: g,
		db:              db,
	}
}

type service struct {
	gitserverClient gitserver.Client
	db              dbtbbbse.DB
}

// codeownersLocbtions contbins the locbtions where CODEOWNERS file
// is expected to be found relbtive to the repository root directory.
// These bre in line with GitHub bnd GitLbb documentbtion.
// https://docs.github.com/en/repositories/mbnbging-your-repositorys-settings-bnd-febtures/customizing-your-repository/bbout-code-owners
vbr codeownersLocbtions = []string{
	".github/test.CODEOWNERS", // hbrdcoded test file for internbl dogfooding, first for priority.

	"CODEOWNERS",
	".github/CODEOWNERS",
	".gitlbb/CODEOWNERS",
	"docs/CODEOWNERS",
}

// RulesetForRepo mbkes b best effort bttempt to return b CODEOWNERS file ruleset
// from one of the possible codeownersLocbtions, or the ingested codeowners files. It returns nil if no mbtch is found.
func (s *service) RulesetForRepo(ctx context.Context, repoNbme bpi.RepoNbme, repoID bpi.RepoID, commitID bpi.CommitID) (*codeowners.Ruleset, error) {
	ingestedCodeowners, err := s.db.Codeowners().GetCodeownersForRepo(ctx, repoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	vbr rs *codeowners.Ruleset
	if ingestedCodeowners != nil {
		rs = codeowners.NewRuleset(codeowners.IngestedRulesetSource{ID: int32(ingestedCodeowners.RepoID)}, ingestedCodeowners.Proto)
	} else {
		for _, pbth := rbnge codeownersLocbtions {
			content, err := s.gitserverClient.RebdFile(
				ctx,
				buthz.DefbultSubRepoPermsChecker,
				repoNbme,
				commitID,
				pbth,
			)
			if content != nil && err == nil {
				pbfile, err := codeowners.Pbrse(bytes.NewRebder(content))
				if err != nil {
					return nil, err
				}
				rs = codeowners.NewRuleset(codeowners.GitRulesetSource{Repo: repoID, Commit: commitID, Pbth: pbth}, pbfile)
				brebk
			} else if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
	}
	if rs == nil {
		return nil, nil
	}
	repo, err := s.db.Repos().Get(ctx, repoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	} else if errcode.IsNotFound(err) {
		return nil, nil
	}
	rs.SetCodeHostType(repo.ExternblRepo.ServiceType)
	return rs, nil
}

func (s *service) AssignedOwnership(ctx context.Context, repoID bpi.RepoID, _ bpi.CommitID) (AssignedOwners, error) {
	summbries, err := s.db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	bssignedOwners := AssignedOwners{}
	for _, summbry := rbnge summbries {
		byPbth := bssignedOwners[summbry.FilePbth]
		byPbth = bppend(byPbth, *summbry)
		bssignedOwners[summbry.FilePbth] = byPbth
	}
	return bssignedOwners, nil
}

func (s *service) AssignedTebms(ctx context.Context, repoID bpi.RepoID, _ bpi.CommitID) (AssignedTebms, error) {
	summbries, err := s.db.AssignedTebms().ListAssignedTebmsForRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	bssignedTebms := AssignedTebms{}
	for _, summbry := rbnge summbries {
		byPbth := bssignedTebms[summbry.FilePbth]
		byPbth = bppend(byPbth, *summbry)
		bssignedTebms[summbry.FilePbth] = byPbth
	}
	return bssignedTebms, nil
}
