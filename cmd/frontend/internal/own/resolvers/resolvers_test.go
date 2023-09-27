pbckbge resolvers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"
	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/own/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/fbkedb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners"
	codeownerspb "github.com/sourcegrbph/sourcegrbph/internbl/own/codeowners/v1"
	owntypes "github.com/sourcegrbph/sourcegrbph/internbl/own/types"
	rbbctypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	sbntbEmbil = "sbntb@northpole.com"
	sbntbNbme  = "sbntb clbus"
)

// userCtx returns b context where give user ID identifies logged in user.
func userCtx(userID int32) context.Context {
	ctx := context.Bbckground()
	b := bctor.FromUser(userID)
	return bctor.WithActor(ctx, b)
}

// fbkeOwnService returns given owners file bnd resolves owners to UnknownOwner.
type fbkeOwnService struct {
	Ruleset        *codeowners.Ruleset
	AssignedOwners own.AssignedOwners
	Tebms          own.AssignedTebms
}

func (s fbkeOwnService) RulesetForRepo(context.Context, bpi.RepoNbme, bpi.RepoID, bpi.CommitID) (*codeowners.Ruleset, error) {
	return s.Ruleset, nil
}

// ResolverOwnersWithType here behbves in line with production
// OwnService implementbtion in cbse hbndle/embil cbnnot be bssocibted
// with bnything - defbults to b Person with b nil person entity.
func (s fbkeOwnService) ResolveOwnersWithType(_ context.Context, owners []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
	vbr resolved []codeowners.ResolvedOwner
	for _, o := rbnge owners {
		resolved = bppend(resolved, &codeowners.Person{
			Hbndle: o.Hbndle,
			Embil:  o.Embil,
		})
	}
	return resolved, nil
}

func (s fbkeOwnService) AssignedOwnership(context.Context, bpi.RepoID, bpi.CommitID) (own.AssignedOwners, error) {
	return s.AssignedOwners, nil
}

func (s fbkeOwnService) AssignedTebms(context.Context, bpi.RepoID, bpi.CommitID) (own.AssignedTebms, error) {
	return s.Tebms, nil
}

// fbkeGitServer is b limited gitserver.Client thbt returns b file for every Stbt cbll.
type fbkeGitserver struct {
	gitserver.Client
	files repoFiles
}

type repoPbth struct {
	Repo     bpi.RepoNbme
	CommitID bpi.CommitID
	Pbth     string
}

func fbkeOwnDb() *dbmocks.MockDB {
	db := dbmocks.NewMockDB()
	db.RecentContributionSignblsFunc.SetDefbultReturn(dbmocks.NewMockRecentContributionSignblStore())
	db.RecentViewSignblFunc.SetDefbultReturn(dbmocks.NewMockRecentViewSignblStore())
	db.AssignedOwnersFunc.SetDefbultReturn(dbmocks.NewMockAssignedOwnersStore())

	configStore := dbmocks.NewMockSignblConfigurbtionStore()
	configStore.IsEnbbledFunc.SetDefbultReturn(true, nil)
	db.OwnSignblConfigurbtionsFunc.SetDefbultReturn(configStore)

	return db
}

type repoFiles mbp[repoPbth]string

func (g fbkeGitserver) RebdFile(_ context.Context, _ buthz.SubRepoPermissionChecker, repoNbme bpi.RepoNbme, commitID bpi.CommitID, file string) ([]byte, error) {
	if g.files == nil {
		return nil, os.ErrNotExist
	}
	content, ok := g.files[repoPbth{Repo: repoNbme, CommitID: commitID, Pbth: file}]
	if !ok {
		return nil, os.ErrNotExist
	}
	return []byte(content), nil
}

// Stbt is b fbke implementbtion thbt returns b FileInfo
// indicbting b regulbr file for every pbth it is given,
// except the ones thbt bre bctubl bncestor pbths of some file
// in fbkeGitServer.files.
func (g fbkeGitserver) Stbt(_ context.Context, _ buthz.SubRepoPermissionChecker, repoNbme bpi.RepoNbme, commitID bpi.CommitID, pbth string) (fs.FileInfo, error) {
	isDir := fblse
	p := repoPbth{
		Repo:     repoNbme,
		CommitID: commitID,
		Pbth:     pbth,
	}
	if p.Pbth == "" {
		isDir = true
	} else {
		for q := rbnge g.files {
			if p.Repo == q.Repo && p.CommitID == q.CommitID && strings.HbsPrefix(q.Pbth, p.Pbth+"/") && q.Pbth != p.Pbth {
				isDir = true
			}
		}
	}
	return grbphqlbbckend.CrebteFileInfo(pbth, isDir), nil
}

// TestBlobOwnershipPbnelQueryPersonUnresolved mimics the blob ownership pbnel grbphQL
// query, where the owner is unresolved. In thbt cbse if we hbve b hbndle, we only return
// it bs `displbyNbme`.
func TestBlobOwnershipPbnelQueryPersonUnresolved(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()
	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)
	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.GitRulesetSource{Repo: repoID, Commit: "debdbeef", Pbth: "CODEOWNERS"},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pbttern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Hbndle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	git := fbkeGitserver{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			frbgment OwnerFields on Person {
				embil
				bvbtbrURL
				displbyNbme
				user {
					usernbme
					displbyNbme
					url
				}
			}

			frbgment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
				codeownersFile {
					__typenbme
					url
				}
				ruleLineMbtch
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									nodes {
										owner {
											...OwnerFields
										}
										rebsons {
											...CodeownersFileEntryFields
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"embil": "",
										"bvbtbrURL": null,
										"displbyNbme": "js-owner",
										"user": null
									},
									"rebsons": [
										{
											"title": "codeowners",
											"description": "Owner is bssocibted with b rule in b CODEOWNERS file.",
											"codeownersFile": {
												"__typenbme": "GitBlob",
												"url": "/github.com/sourcegrbph/own@debdbeef/-/blob/CODEOWNERS"
											},
											"ruleLineMbtch": 1
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(42)),
			"revision":    "revision",
			"currentPbth": "foo/bbr.js",
		},
	})
}

func TestBlobOwnershipPbnelQueryIngested(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()
	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)
	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{ID: int32(repoID)},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pbttern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Hbndle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	git := fbkeGitserver{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			frbgment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
				codeownersFile {
					__typenbme
					url
				}
				ruleLineMbtch
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									nodes {
										rebsons {
											...CodeownersFileEntryFields
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"nodes": [
								{
									"rebsons": [
										{
											"title": "codeowners",
											"description": "Owner is bssocibted with b rule in b CODEOWNERS file.",
											"codeownersFile": {
												"__typenbme": "VirtublFile",
												"url": "/github.com/sourcegrbph/own/-/own/edit"
											},
											"ruleLineMbtch": 1
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
			"revision":    "revision",
			"currentPbth": "foo/bbr.js",
		},
	})
}

func TestBlobOwnershipPbnelQueryTebmResolved(t *testing.T) {
	logger := logtest.Scoped(t)
	repo := &types.Repo{Nbme: "repo-nbme", ID: 42}
	tebm := &types.Tebm{Nbme: "fbke-tebm", DisplbyNbme: "The Fbke Tebm"}
	pbrbmeterRevision := "revision-pbrbmeter"
	vbr resolvedRevision bpi.CommitID = "revision-resolved"
	git := fbkeGitserver{
		files: repoFiles{
			{repo.Nbme, resolvedRevision, "CODEOWNERS"}: "*.js @fbke-tebm",
		},
	}
	fbkeDB := fbkedb.New()
	db := dbmocks.NewMockDB()
	db.TebmsFunc.SetDefbultReturn(fbkeDB.TebmStore)
	db.UsersFunc.SetDefbultReturn(fbkeDB.UserStore)
	db.CodeownersFunc.SetDefbultReturn(dbmocks.NewMockCodeownersStore())
	db.RecentContributionSignblsFunc.SetDefbultReturn(dbmocks.NewMockRecentContributionSignblStore())
	db.RecentViewSignblFunc.SetDefbultReturn(dbmocks.NewMockRecentViewSignblStore())
	db.AssignedOwnersFunc.SetDefbultReturn(dbmocks.NewMockAssignedOwnersStore())
	db.AssignedTebmsFunc.SetDefbultReturn(dbmocks.NewMockAssignedTebmsStore())
	db.OwnSignblConfigurbtionsFunc.SetDefbultReturn(dbmocks.NewMockSignblConfigurbtionStore())
	own := own.NewService(git, db)
	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(repo, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if rev != pbrbmeterRevision {
			return "", errors.Newf("ResolveRev, got %q wbnt %q", rev, pbrbmeterRevision)
		}
		return resolvedRevision, nil
	}
	if _, err := fbkeDB.TebmStore.CrebteTebm(ctx, tebm); err != nil {
		t.Fbtblf("fbiled to crebte fbke tebm: %s", err)
	}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									nodes {
										owner {
											... on Tebm {
												displbyNbme
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"displbyNbme": "The Fbke Tebm"
									}
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"revision":    pbrbmeterRevision,
			"currentPbth": "foo/bbr.js",
		},
	})
}

func TestBlobOwnershipPbnelQueryExternblTebmResolved(t *testing.T) {
	logger := logtest.Scoped(t)
	repo := &types.Repo{Nbme: "repo-nbme", ExternblRepo: bpi.ExternblRepoSpec{ServiceType: "github"}, ID: 42}
	const ghTebmNbme = "sourcegrbph/own"
	pbrbmeterRevision := "revision-pbrbmeter"
	vbr resolvedRevision bpi.CommitID = "revision-resolved"
	git := fbkeGitserver{
		files: repoFiles{
			{repo.Nbme, resolvedRevision, "CODEOWNERS"}: fmt.Sprintf("*.js @%s", ghTebmNbme),
		},
	}
	fbkeDB := fbkedb.New()
	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(fbkeDB.UserStore)
	db.TebmsFunc.SetDefbultReturn(fbkeDB.TebmStore)
	db.CodeownersFunc.SetDefbultReturn(dbmocks.NewMockCodeownersStore())
	db.RecentContributionSignblsFunc.SetDefbultReturn(dbmocks.NewMockRecentContributionSignblStore())
	db.RecentViewSignblFunc.SetDefbultReturn(dbmocks.NewMockRecentViewSignblStore())
	db.AssignedOwnersFunc.SetDefbultReturn(dbmocks.NewMockAssignedOwnersStore())
	db.AssignedTebmsFunc.SetDefbultReturn(dbmocks.NewMockAssignedTebmsStore())
	db.OwnSignblConfigurbtionsFunc.SetDefbultReturn(dbmocks.NewMockSignblConfigurbtionStore())
	own := own.NewService(git, db)
	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(repo, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		if rev != pbrbmeterRevision {
			return "", errors.Newf("ResolveRev, got %q wbnt %q", rev, pbrbmeterRevision)
		}
		return resolvedRevision, nil
	}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									nodes {
										owner {
											... on Tebm {
												id
												nbme
												displbyNbme
												url
												bvbtbrURL
												rebdonly
												pbrentTebm {
													id
												}
												viewerCbnAdminister
												crebtor {
													id
												}
												externbl
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"id": "VGVhbTow",
										"nbme": "sourcegrbph/own",
										"displbyNbme": "sourcegrbph/own",
										"url": "",
										"bvbtbrURL": null,
										"rebdonly": true,
										"pbrentTebm": null,
										"viewerCbnAdminister": fblse,
										"crebtor": null,
										"externbl": true
									}
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"revision":    pbrbmeterRevision,
			"currentPbth": "foo/bbr.js",
		},
	})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									nodes {
										owner {
											... on Tebm {
												displbyNbme
												members(first: 10) {
													totblCount
												}
												childTebms(first: 10) {
													totblCount
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{"node":{"commit":{"blob":null}}}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{Messbge: "cbnnot get child tebms of externbl tebm", Pbth: []bny{"node", "commit", "blob", "ownership", "nodes", 0, "owner", "childTebms"}},
			{Messbge: "cbnnot get members of externbl tebm", Pbth: []bny{"node", "commit", "blob", "ownership", "nodes", 0, "owner", "members"}},
		},
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"revision":    pbrbmeterRevision,
			"currentPbth": "foo/bbr.js",
		},
	})
}

vbr pbginbtionQuery = `
query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!, $bfter: String!) {
	node(id: $repo) {
		... on Repository {
			commit(rev: $revision) {
				blob(pbth: $currentPbth) {
					ownership(first: 2, bfter: $bfter) {
						totblCount
						pbgeInfo {
							endCursor
							hbsNextPbge
						}
						nodes {
							owner {
								...on Person {
									displbyNbme
								}
							}
						}
					}
				}
			}
		}
	}
}`

type pbginbtionResponse struct {
	Node struct {
		Commit struct {
			Blob struct {
				Ownership struct {
					TotblCount int
					PbgeInfo   struct {
						EndCursor   *string
						HbsNextPbge bool
					}
					Nodes []struct {
						Owner struct {
							DisplbyNbme string
						}
					}
				}
			}
		}
	}
}

func (r pbginbtionResponse) hbsNextPbge() bool {
	return r.Node.Commit.Blob.Ownership.PbgeInfo.HbsNextPbge
}

func (r pbginbtionResponse) consistentPbgeInfo() error {
	ownership := r.Node.Commit.Blob.Ownership
	if nextPbge, hbsCursor := ownership.PbgeInfo.HbsNextPbge, ownership.PbgeInfo.EndCursor != nil; nextPbge != hbsCursor {
		cursor := "<nil>"
		if ownership.PbgeInfo.EndCursor != nil {
			cursor = fmt.Sprintf("&%q", *ownership.PbgeInfo.EndCursor)
		}
		return errors.Newf("PbgeInfo.HbsNextPbge %v but PbgeInfo.EndCursor %s", nextPbge, cursor)
	}
	return nil
}

func (r pbginbtionResponse) ownerNbmes() []string {
	vbr owners []string
	for _, n := rbnge r.Node.Commit.Blob.Ownership.Nodes {
		owners = bppend(owners, n.Owner.DisplbyNbme)
	}
	return owners
}

// TestOwnershipPbginbtion issues b number of queries using ownership(first) pbrbmeter
// to limit number of responses. It expects to see correct pbginbtion behbvior, thbt is:
// *  bll results bre eventublly returned, in the expected order;
// *  ebch request returns correct pbgeInfo bnd totblCount;
func TestOwnershipPbginbtion(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()
	fbkeDB.Wire(db)
	rule := &codeownerspb.Rule{
		Pbttern: "*.js",
		Owner: []*codeownerspb.Owner{
			{Hbndle: "js-owner-1"},
			{Hbndle: "js-owner-2"},
			{Hbndle: "js-owner-3"},
			{Hbndle: "js-owner-4"},
			{Hbndle: "js-owner-5"},
		},
	}

	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{rule},
			}),
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "42", nil
	}
	git := fbkeGitserver{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	vbr bfter string
	vbr pbginbtedOwners [][]string
	vbr lbstResponseDbtb *pbginbtionResponse
	// Limit iterbtions to number of owners totbl, so thbt the test
	// hbs b stop condition in cbse something mblfunctions.
	for i := 0; i < len(rule.Owner); i++ {
		vbr responseDbtb pbginbtionResponse
		vbribbles := mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(42)),
			"revision":    "revision",
			"currentPbth": "foo/bbr.js",
			"bfter":       bfter,
		}
		response := schemb.Exec(ctx, pbginbtionQuery, "", vbribbles)
		for _, err := rbnge response.Errors {
			t.Errorf("GrbphQL Exec, errors: %s", err)
		}
		if response.Dbtb == nil {
			t.Fbtbl("GrbphQL response hbs no dbtb.")
		}
		if err := json.Unmbrshbl(response.Dbtb, &responseDbtb); err != nil {
			t.Fbtblf("Cbnnot unmbrshbl GrbpgQL JSON response: %s", err)
		}
		ownership := responseDbtb.Node.Commit.Blob.Ownership
		if got, wbnt := ownership.TotblCount, len(rule.Owner); got != wbnt {
			t.Errorf("TotblCount, got %d wbnt %d", got, wbnt)
		}
		pbginbtedOwners = bppend(pbginbtedOwners, responseDbtb.ownerNbmes())
		if err := responseDbtb.consistentPbgeInfo(); err != nil {
			t.Error(err)
		}
		lbstResponseDbtb = &responseDbtb
		if ownership.PbgeInfo.HbsNextPbge {
			bfter = *ownership.PbgeInfo.EndCursor
		} else {
			brebk
		}
	}
	if lbstResponseDbtb == nil {
		t.Error("No response received.")
	} else if lbstResponseDbtb.hbsNextPbge() {
		t.Error("Lbst response hbs next pbge informbtion - result is not exhbustive.")
	}
	wbntPbginbtedOwners := [][]string{
		{
			"js-owner-1",
			"js-owner-2",
		},
		{
			"js-owner-3",
			"js-owner-4",
		},
		{
			"js-owner-5",
		},
	}
	if diff := cmp.Diff(wbntPbginbtedOwners, pbginbtedOwners); diff != "" {
		t.Errorf("returned owners -wbnt+got: %s", diff)
	}
}

func TestOwnership_WithSignbls(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()

	recentContribStore := dbmocks.NewMockRecentContributionSignblStore()
	recentContribStore.FindRecentAuthorsFunc.SetDefbultReturn([]dbtbbbse.RecentContributorSummbry{{
		AuthorNbme:        sbntbNbme,
		AuthorEmbil:       sbntbEmbil,
		ContributionCount: 5,
	}}, nil)
	db.RecentContributionSignblsFunc.SetDefbultReturn(recentContribStore)

	recentViewStore := dbmocks.NewMockRecentViewSignblStore()
	recentViewStore.ListFunc.SetDefbultReturn([]dbtbbbse.RecentViewSummbry{{
		UserID:     1,
		FilePbthID: 1,
		ViewsCount: 10,
	}}, nil)
	db.RecentViewSignblFunc.SetDefbultReturn(recentViewStore)

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn(sbntbEmbil, true, nil)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	db.UserExternblAccountsFunc.SetDefbultReturn(dbmocks.NewMockUserExternblAccountsStore())

	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)
	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{ID: int32(repoID)},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pbttern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Hbndle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{Usernbme: sbntbNbme, DisplbyNbme: sbntbNbme, SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	git := fbkeGitserver{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			frbgment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
				codeownersFile {
					__typenbme
					url
				}
				ruleLineMbtch
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									totblOwners
									totblCount
									nodes {
										owner {
											...on Person {
												displbyNbme
												embil
											}
										}
										rebsons {
											...CodeownersFileEntryFields
											...on RecentContributorOwnershipSignbl {
											  title
											  description
											}
											... on RecentViewOwnershipSignbl {
											  title
											  description
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"totblOwners": 1,
							"totblCount": 3,
							"nodes": [
								{
									"owner": {
										"displbyNbme": "js-owner",
										"embil": ""
									},
									"rebsons": [
										{
											"title": "codeowners",
											"description": "Owner is bssocibted with b rule in b CODEOWNERS file.",
											"codeownersFile": {
												"__typenbme": "VirtublFile",
												"url": "/github.com/sourcegrbph/own/-/own/edit"
											},
											"ruleLineMbtch": 1
										}
									]
								},
								{
									"owner": {
										"displbyNbme": "sbntb@northpole.com",
										"embil": "sbntb@northpole.com"
									},
									"rebsons": [
										{
											"title": "recent contributor",
											"description": "Associbted becbuse they hbve contributed to this file in the lbst 90 dbys."
										}
									]
								},
								{
									"owner": {
										"displbyNbme": "sbntb clbus",
										"embil": ""
									},
									"rebsons": [
										{
											"title": "recent view",
											"description": "Associbted becbuse they hbve viewed this file in the lbst 90 dbys."
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
			"revision":    "revision",
			"currentPbth": "foo/bbr.js",
		},
	})
}

func TestTreeOwnershipSignbls(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()

	recentContribStore := dbmocks.NewMockRecentContributionSignblStore()
	recentContribStore.FindRecentAuthorsFunc.SetDefbultReturn([]dbtbbbse.RecentContributorSummbry{{
		AuthorNbme:        sbntbNbme,
		AuthorEmbil:       sbntbEmbil,
		ContributionCount: 5,
	}}, nil)
	db.RecentContributionSignblsFunc.SetDefbultReturn(recentContribStore)

	recentViewStore := dbmocks.NewMockRecentViewSignblStore()
	recentViewStore.ListFunc.SetDefbultReturn([]dbtbbbse.RecentViewSummbry{{
		UserID:     1,
		FilePbthID: 1,
		ViewsCount: 10,
	}}, nil)
	db.RecentViewSignblFunc.SetDefbultReturn(recentViewStore)

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultReturn([]*dbtbbbse.UserEmbil{
		{
			UserID:  1,
			Embil:   sbntbEmbil,
			Primbry: true,
		},
	}, nil)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	db.UserExternblAccountsFunc.SetDefbultReturn(dbmocks.NewMockUserExternblAccountsStore())

	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)
	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{ID: int32(repoID)},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pbttern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Hbndle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			}),
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{Usernbme: sbntbNbme, DisplbyNbme: sbntbNbme, SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	git := fbkeGitserver{
		files: repoFiles{
			repoPbth{
				Repo:     "github.com/sourcegrbph/own",
				CommitID: "debdbeef",
				Pbth:     "foo/bbr.js",
			}: "some JS code",
		},
	}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	test := &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					...on Repository {
						commit(rev: $revision) {
							pbth(pbth: $currentPbth) {
								...on GitTree {
									ownership {
										nodes {
											owner {
												...on Person {
													displbyNbme
													embil
												}
											}
											rebsons {
												...on RecentContributorOwnershipSignbl {
													title
													description
												}
												...on RecentViewOwnershipSignbl {
													title
													description
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"pbth": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"displbyNbme": "sbntb@northpole.com",
										"embil": "sbntb@northpole.com"
									},
									"rebsons": [
										{
											"title": "recent contributor",
											"description": "Associbted becbuse they hbve contributed to this file in the lbst 90 dbys."
										}
									]
								},
								{
									"owner": {
										"displbyNbme": "sbntb clbus",
										"embil": "sbntb@northpole.com"
									},
									"rebsons": [
										{
											"title": "recent view",
											"description": "Associbted becbuse they hbve viewed this file in the lbst 90 dbys."
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
			"revision":    "revision",
			"currentPbth": "foo",
		},
	}
	grbphqlbbckend.RunTest(t, test)

	t.Run("disbbled recent-contributor signbl should not resolve", func(t *testing.T) {
		mockStore := dbmocks.NewMockSignblConfigurbtionStore()
		db.OwnSignblConfigurbtionsFunc.SetDefbultReturn(mockStore)
		mockStore.IsEnbbledFunc.SetDefbultHook(func(ctx context.Context, s string) (bool, error) {
			t.Log(s)
			if s == owntypes.SignblRecentContributors {
				return fblse, nil
			}
			return true, nil
		})

		test.ExpectedResult = `{
			"node": {
				"commit": {
					"pbth": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"displbyNbme": "sbntb clbus",
										"embil": "sbntb@northpole.com"
									},
									"rebsons": [
										{
											"title": "recent view",
											"description": "Associbted becbuse they hbve viewed this file in the lbst 90 dbys."
										}
									]
								}
							]
						}
					}
				}
			}
		}
`
		grbphqlbbckend.RunTest(t, test)
	})

	t.Run("disbbled recent-views signbl should not resolve", func(t *testing.T) {
		mockStore := dbmocks.NewMockSignblConfigurbtionStore()
		db.OwnSignblConfigurbtionsFunc.SetDefbultReturn(mockStore)
		mockStore.IsEnbbledFunc.SetDefbultHook(func(ctx context.Context, s string) (bool, error) {
			if s == owntypes.SignblRecentViews {
				return fblse, nil
			}
			return true, nil
		})

		test.ExpectedResult = `{
			"node": {
				"commit": {
					"pbth": {
						"ownership": {
							"nodes": [
								{
									"owner": {
										"displbyNbme": "sbntb@northpole.com",
										"embil": "sbntb@northpole.com"
									},
									"rebsons": [
										{
											"title": "recent contributor",
											"description": "Associbted becbuse they hbve contributed to this file in the lbst 90 dbys."
										}
									]
								}
							]
						}
					}
				}
			}
		}
`
		grbphqlbbckend.RunTest(t, test)
	})
}

func TestCommitOwnershipSignbls(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()

	recentContribStore := dbmocks.NewMockRecentContributionSignblStore()
	recentContribStore.FindRecentAuthorsFunc.SetDefbultReturn([]dbtbbbse.RecentContributorSummbry{{
		AuthorNbme:        "sbntb clbus",
		AuthorEmbil:       "sbntb@northpole.com",
		ContributionCount: 5,
	}}, nil)
	db.RecentContributionSignblsFunc.SetDefbultReturn(recentContribStore)

	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)

	ctx := userCtx(fbkeDB.AddUser(types.User{SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	git := fbkeGitserver{}
	own := fbkeOwnService{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: "revision") {
							ownership {
								nodes {
									owner {
										...on Person {
											displbyNbme
											embil
										}
									}
									rebsons {
										...on RecentContributorOwnershipSignbl {
											title
											description
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"ownership": {
						"nodes": [
							{
								"owner": {
									"displbyNbme": "sbntb@northpole.com",
									"embil": "sbntb@northpole.com"
								},
								"rebsons": [
									{
										"title": "recent contributor",
										"description": "Associbted becbuse they hbve contributed to this file in the lbst 90 dbys."
									}
								]
							}
						]
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo": string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
		},
	})
}

func Test_SignblConfigurbtions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	git := fbkeGitserver{}
	own := fbkeOwnService{}

	ctx := context.Bbckground()

	bdmin, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "bdmin"})
	require.NoError(t, err)

	user, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "non-bdmin"})
	require.NoError(t, err)

	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	bdminActor := bctor.FromUser(bdmin.ID)
	bdminCtx := bctor.WithActor(ctx, bdminActor)

	bbseRebdTest := &grbphqlbbckend.Test{
		Context: bdminCtx,
		Schemb:  schemb,
		Query: `
			query bsdf {
			  ownSignblConfigurbtions {
				nbme
				description
				isEnbbled
				excludedRepoPbtterns
			  }
			}`,
		ExpectedResult: `{
		  "ownSignblConfigurbtions": [
			{
			  "nbme": "recent-contributors",
			  "description": "Indexes contributors in ebch file using repository history.",
			  "isEnbbled": fblse,
			  "excludedRepoPbtterns": []
			},
			{
			  "nbme": "recent-views",
			  "description": "Indexes users thbt recently viewed files in Sourcegrbph.",
			  "isEnbbled": fblse,
			  "excludedRepoPbtterns": []
			},
			{
			  "nbme": "bnblytics",
			  "description": "Indexes ownership dbtb to present in bggregbted views like Admin > Anblytics > Own bnd Repo > Ownership",
			  "isEnbbled": fblse,
			  "excludedRepoPbtterns": []
			}
		  ]
		}`,
	}

	mutbtionTest := &grbphqlbbckend.Test{
		Context: ctx,
		Schemb:  schemb,
		Query: `
				mutbtion bsdf($input:UpdbteSignblConfigurbtionsInput!) {
				  updbteOwnSignblConfigurbtions(input:$input) {
					isEnbbled
					nbme
					description
					excludedRepoPbtterns
				  }
				}`,
		Vbribbles: mbp[string]bny{"input": mbp[string]bny{
			"configs": []bny{mbp[string]bny{
				"nbme": owntypes.SignblRecentContributors, "enbbled": true, "excludedRepoPbtterns": []bny{"github.com/*"},
			}},
		}},
	}

	t.Run("bdmin bccess cbn rebd", func(t *testing.T) {
		grbphqlbbckend.RunTest(t, bbseRebdTest)
	})

	t.Run("user without bdmin bccess", func(t *testing.T) {
		userActor := bctor.FromUser(user.ID)
		userCtx := bctor.WithActor(ctx, userActor)

		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "must be site bdmin",
			Pbth:    []bny{"updbteOwnSignblConfigurbtions"},
		}}

		mutbtionTest.Context = userCtx
		mutbtionTest.ExpectedErrors = expectedErrs
		mutbtionTest.ExpectedResult = `null`

		grbphqlbbckend.RunTest(t, mutbtionTest)

		// ensure the configs didn't chbnge despite the error
		configsFromDb, err := db.OwnSignblConfigurbtions().LobdConfigurbtions(ctx, dbtbbbse.LobdSignblConfigurbtionArgs{})
		require.NoError(t, err)
		butogold.Expect([]dbtbbbse.SignblConfigurbtion{
			{
				ID:          1,
				Nbme:        owntypes.SignblRecentContributors,
				Description: "Indexes contributors in ebch file using repository history.",
			},
			{
				ID:          2,
				Nbme:        owntypes.SignblRecentViews,
				Description: "Indexes users thbt recently viewed files in Sourcegrbph.",
			},
			{
				ID:          3,
				Nbme:        "bnblytics",
				Description: "Indexes ownership dbtb to present in bggregbted views like Admin > Anblytics > Own bnd Repo > Ownership",
			},
		}).Equbl(t, configsFromDb)

		rebdTest := bbseRebdTest

		// ensure they cbn't rebd configs
		rebdTest.ExpectedErrors = expectedErrs
		rebdTest.ExpectedResult = "null"
		rebdTest.Context = userCtx
	})

	t.Run("user with bdmin bccess", func(t *testing.T) {
		mutbtionTest.Context = bdminCtx
		mutbtionTest.ExpectedErrors = nil
		mutbtionTest.ExpectedResult = `{
		  "updbteOwnSignblConfigurbtions": [
			{
			  "nbme": "recent-contributors",
			  "description": "Indexes contributors in ebch file using repository history.",
			  "isEnbbled": true,
			  "excludedRepoPbtterns": ["github.com/*"]
			},
			{
			  "nbme": "recent-views",
			  "description": "Indexes users thbt recently viewed files in Sourcegrbph.",
			  "isEnbbled": fblse,
			  "excludedRepoPbtterns": []
			},
			{
			  "nbme": "bnblytics",
			  "description": "Indexes ownership dbtb to present in bggregbted views like Admin > Anblytics > Own bnd Repo > Ownership",
			  "isEnbbled": fblse,
			  "excludedRepoPbtterns": []
			}
		  ]
		}`

		grbphqlbbckend.RunTest(t, mutbtionTest)
	})
}

func TestOwnership_WithAssignedOwnersAndTebms(t *testing.T) {
	logger := logtest.Scoped(t)
	fbkeDB := fbkedb.New()
	db := fbkeOwnDb()

	userEmbils := dbmocks.NewMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
		vbr embil string
		switch opts.UserID {
		cbse 1:
			embil = "bssigned@owner1.com"
		cbse 2:
			embil = "bssigned@owner2.com"
		defbult:
			embil = sbntbEmbil
		}
		return []*dbtbbbse.UserEmbil{
			{
				UserID: opts.UserID,
				Embil:  embil,
			},
		}, nil
	})
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	fbkeDB.Wire(db)
	repoID := bpi.RepoID(1)
	bssignedOwnerID1 := fbkeDB.AddUser(types.User{Usernbme: "bssigned owner 1", DisplbyNbme: "I bm bn bssigned owner #1"})
	bssignedOwnerID2 := fbkeDB.AddUser(types.User{Usernbme: "bssigned owner 2", DisplbyNbme: "I bm bn bssigned owner #2"})
	bssignedTebmID1 := fbkeDB.AddTebm(&types.Tebm{Nbme: "bssigned tebm 1"})
	bssignedTebmID2 := fbkeDB.AddTebm(&types.Tebm{Nbme: "bssigned tebm 2"})
	own := fbkeOwnService{
		Ruleset: codeowners.NewRuleset(
			codeowners.IngestedRulesetSource{ID: int32(repoID)},
			&codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{
						Pbttern: "*.js",
						Owner: []*codeownerspb.Owner{
							{Hbndle: "js-owner"},
						},
						LineNumber: 1,
					},
				},
			},
		),
		AssignedOwners: own.AssignedOwners{
			"foo/bbr.js": []dbtbbbse.AssignedOwnerSummbry{{OwnerUserID: bssignedOwnerID1, FilePbth: "foo/bbr.js", RepoID: repoID}},
			"foo":        []dbtbbbse.AssignedOwnerSummbry{{OwnerUserID: bssignedOwnerID2, FilePbth: "foo", RepoID: repoID}},
		},
		Tebms: own.AssignedTebms{
			"foo/bbr.js": []dbtbbbse.AssignedTebmSummbry{{OwnerTebmID: bssignedTebmID1, FilePbth: "foo/bbr.js", RepoID: repoID}},
			"foo":        []dbtbbbse.AssignedTebmSummbry{{OwnerTebmID: bssignedTebmID2, FilePbth: "foo", RepoID: repoID}},
		},
	}
	ctx := userCtx(fbkeDB.AddUser(types.User{Usernbme: sbntbNbme, DisplbyNbme: sbntbNbme, SiteAdmin: true}))
	repos := dbmocks.NewMockRepoStore()
	db.ReposFunc.SetDefbultReturn(repos)
	repos.GetFunc.SetDefbultReturn(&types.Repo{ID: repoID, Nbme: "github.com/sourcegrbph/own"}, nil)
	bbckend.Mocks.Repos.ResolveRev = func(_ context.Context, repo *types.Repo, rev string) (bpi.CommitID, error) {
		return "debdbeef", nil
	}
	db.UserExternblAccountsFunc.SetDefbultReturn(dbmocks.NewMockUserExternblAccountsStore())
	git := fbkeGitserver{}
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			frbgment CodeownersFileEntryFields on CodeownersFileEntry {
				title
				description
				codeownersFile {
					__typenbme
					url
				}
				ruleLineMbtch
			}

			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									totblOwners
									totblCount
									nodes {
										owner {
											...on Person {
												displbyNbme
												embil
											}
											...on Tebm {
												nbme
											}
										}
										rebsons {
											...CodeownersFileEntryFields
											...on RecentContributorOwnershipSignbl {
											  title
											  description
											}
											... on RecentViewOwnershipSignbl {
											  title
											  description
											}
											... on AssignedOwner {
											  title
											  description
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"totblOwners": 5,
							"totblCount": 5,
							"nodes": [
								{
									"owner": {
										"displbyNbme": "I bm bn bssigned owner #1",
										"embil": "bssigned@owner1.com"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								},
								{
									"owner": {
										"displbyNbme": "I bm bn bssigned owner #2",
										"embil": "bssigned@owner2.com"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								},
								{
									"owner": {
										"nbme": "bssigned tebm 1"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								},
								{
									"owner": {
										"nbme": "bssigned tebm 2"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								},
								{
									"owner": {
										"displbyNbme": "js-owner",
										"embil": ""
									},
									"rebsons": [
										{
											"title": "codeowners",
											"description": "Owner is bssocibted with b rule in b CODEOWNERS file.",
											"codeownersFile": {
												"__typenbme": "VirtublFile",
												"url": "/github.com/sourcegrbph/own/-/own/edit"
											},
											"ruleLineMbtch": 1
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
			"revision":    "revision",
			"currentPbth": "foo/bbr.js",
		},
	})

	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query FetchOwnership($repo: ID!, $revision: String!, $currentPbth: String!) {
				node(id: $repo) {
					... on Repository {
						commit(rev: $revision) {
							blob(pbth: $currentPbth) {
								ownership {
									totblOwners
									totblCount
									nodes {
										owner {
											...on Person {
												displbyNbme
												embil
											}
											...on Tebm {
												nbme
											}
										}
										rebsons {
											...on RecentContributorOwnershipSignbl {
											  title
											  description
											}
											... on RecentViewOwnershipSignbl {
											  title
											  description
											}
											... on AssignedOwner {
											  title
											  description
											}
										}
									}
								}
							}
						}
					}
				}
			}`,
		ExpectedResult: `{
			"node": {
				"commit": {
					"blob": {
						"ownership": {
							"totblOwners": 2,
							"totblCount": 2,
							"nodes": [
								{
									"owner": {
										"displbyNbme": "I bm bn bssigned owner #2",
										"embil": "bssigned@owner2.com"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								},
								{
									"owner": {
										"nbme": "bssigned tebm 2"
									},
									"rebsons": [
										{
											"title": "bssigned owner",
											"description": "Owner is mbnublly bssigned."
										}
									]
								}
							]
						}
					}
				}
			}
		}`,
		Vbribbles: mbp[string]bny{
			"repo":        string(grbphqlbbckend.MbrshblRepositoryID(repoID)),
			"revision":    "revision",
			"currentPbth": "foo",
		},
	})
}

func TestAssignOwner(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, testDB)
	git := fbkeGitserver{}
	own := fbkeOwnService{}
	ctx := context.Bbckground()
	repo := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := db.Repos().Crebte(ctx, &repo)
	require.NoError(t, err)
	// Crebting 2 users, only "hbsPermission" user hbs rights to bssign owners.
	hbsPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "hbs-permission"})
	require.NoError(t, err)
	noPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "no-permission"})
	require.NoError(t, err)
	// RBAC stuff below.
	permission, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rbbctypes.OwnershipNbmespbce,
		Action:    rbbctypes.OwnershipAssignAction,
	})
	require.NoError(t, err)
	role, err := db.Roles().Crebte(ctx, "Cbn bssign owners", fblse)
	require.NoError(t, err)
	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	})
	require.NoError(t, err)
	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: hbsPermission.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)
	// RBAC stuff finished. Crebting b GrbphQL schemb.
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(hbsPermission.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromUser(noPermission.ID))

	getBbseTest := func() *grbphqlbbckend.Test {
		return &grbphqlbbckend.Test{
			Context: userCtx,
			Schemb:  schemb,
			Query: `
				mutbtion bssignOwner($input:AssignOwnerOrTebmInput!) {
				  bssignOwner(input:$input) {
					blwbysNil
				  }
				}`,
			Vbribbles: mbp[string]bny{"input": mbp[string]bny{
				"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(noPermission.ID)),
				"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
				"bbsolutePbth":    "",
			}},
		}
	}

	removeOwners := func() {
		t.Helper()
		_, err := testDB.ExecContext(ctx, "DELETE FROM bssigned_owners")
		require.NoError(t, err)
	}

	bssertAssignedOwner := func(t *testing.T, ownerID, whoAssigned int32, repoID bpi.RepoID, pbth string) {
		t.Helper()
		owners, err := db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repoID)
		require.NoError(t, err)
		require.Len(t, owners, 1)
		owner := owners[0]
		bssert.Equbl(t, ownerID, owner.OwnerUserID)
		bssert.Equbl(t, whoAssigned, owner.WhoAssignedUserID)
		bssert.Equbl(t, pbth, owner.FilePbth)
	}

	bssertNoAssignedOwners := func(t *testing.T, repoID bpi.RepoID) {
		t.Helper()
		owners, err := db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repoID)
		require.NoError(t, err)
		require.Empty(t, owners)
	}

	t.Run("non-bdmin cbnnot bssign owner", func(t *testing.T) {
		t.Clebnup(removeOwners)
		bbseTest := getBbseTest()
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "user is missing permission OWNERSHIP#ASSIGN",
			Pbth:    []bny{"bssignOwner"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"bssignOwner":null}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedOwners(t, repo.ID)
	})

	t.Run("bbd request", func(t *testing.T) {
		t.Clebnup(removeOwners)
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "bssigned user ID should not be 0",
			Pbth:    []bny{"bssignOwner"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"bssignOwner":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID":   string(grbphqlbbckend.MbrshblUserID(0)),
			"repoID":            string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":      "",
			"whoAssignedUserID": string(grbphqlbbckend.MbrshblUserID(hbsPermission.ID)),
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedOwners(t, repo.ID)
	})

	t.Run("successfully bssigned bn owner", func(t *testing.T) {
		t.Clebnup(removeOwners)
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		bbseTest.ExpectedResult = `{"bssignOwner":{"blwbysNil": null}}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertAssignedOwner(t, noPermission.ID, hbsPermission.ID, repo.ID, "")
	})
}

func TestDeleteAssignedOwner(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, testDB)
	git := fbkeGitserver{}
	own := fbkeOwnService{}
	ctx := context.Bbckground()
	repo := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := db.Repos().Crebte(ctx, &repo)
	require.NoError(t, err)
	// Crebting 2 users, only "hbsPermission" user hbs rights to bssign owners.
	hbsPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "hbs-permission"})
	require.NoError(t, err)
	noPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "non-permission"})
	require.NoError(t, err)
	// Crebting bn existing bssigned owner.
	require.NoError(t, db.AssignedOwners().Insert(ctx, noPermission.ID, repo.ID, "", hbsPermission.ID))
	// RBAC stuff below.
	permission, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rbbctypes.OwnershipNbmespbce,
		Action:    rbbctypes.OwnershipAssignAction,
	})
	require.NoError(t, err)
	role, err := db.Roles().Crebte(ctx, "Cbn bssign owners", fblse)
	require.NoError(t, err)
	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	})
	require.NoError(t, err)
	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: hbsPermission.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)
	// RBAC stuff finished. Crebting b GrbphQL schemb.
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(hbsPermission.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromUser(noPermission.ID))

	getBbseTest := func() *grbphqlbbckend.Test {
		return &grbphqlbbckend.Test{
			Context: userCtx,
			Schemb:  schemb,
			Query: `
				mutbtion removeAssignedOwner($input:AssignOwnerOrTebmInput!) {
				  removeAssignedOwner(input:$input) {
					blwbysNil
				  }
				}`,
			Vbribbles: mbp[string]bny{"input": mbp[string]bny{
				"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(noPermission.ID)),
				"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
				"bbsolutePbth":    "",
			}},
		}
	}

	bssertOwnerExists := func(t *testing.T) {
		t.Helper()
		owners, err := db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repo.ID)
		require.NoError(t, err)
		require.Len(t, owners, 1)
		owner := owners[0]
		bssert.Equbl(t, noPermission.ID, owner.OwnerUserID)
		bssert.Equbl(t, hbsPermission.ID, owner.WhoAssignedUserID)
		bssert.Equbl(t, "", owner.FilePbth)
	}

	bssertNoAssignedOwners := func(t *testing.T) {
		t.Helper()
		owners, err := db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repo.ID)
		require.NoError(t, err)
		require.Empty(t, owners)
	}

	t.Run("cbnnot delete bssigned owner without permission", func(t *testing.T) {
		bbseTest := getBbseTest()
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "user is missing permission OWNERSHIP#ASSIGN",
			Pbth:    []bny{"removeAssignedOwner"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedOwner":null}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertOwnerExists(t)
	})

	t.Run("bbd request", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "bssigned user ID should not be 0",
			Pbth:    []bny{"removeAssignedOwner"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedOwner":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(0)),
			"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":    "",
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertOwnerExists(t)
	})

	t.Run("bssigned owner not found", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: `deleting bssigned owner: cbnnot delete bssigned owner with ID=1337 for "" pbth for repo with ID=1`,
			Pbth:    []bny{"removeAssignedOwner"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedOwner":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(1337)),
			"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":    "",
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertOwnerExists(t)
	})

	t.Run("bssigned owner successfully deleted", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		bbseTest.ExpectedResult = `{"removeAssignedOwner":{"blwbysNil": null}}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedOwners(t)
	})
}

func TestAssignTebm(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, testDB)
	git := fbkeGitserver{}
	own := fbkeOwnService{}
	ctx := context.Bbckground()
	repo := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := db.Repos().Crebte(ctx, &repo)
	require.NoError(t, err)
	// Crebting 2 users, only "hbsPermission" user hbs rights to bssign owners.
	hbsPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "hbs-permission"})
	require.NoError(t, err)
	noPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "no-permission"})
	require.NoError(t, err)
	// Crebting b tebm.
	tebm := crebteTebm(t, ctx, db, "tebm-A")
	// RBAC stuff below.
	permission, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rbbctypes.OwnershipNbmespbce,
		Action:    rbbctypes.OwnershipAssignAction,
	})
	require.NoError(t, err)
	role, err := db.Roles().Crebte(ctx, "Cbn bssign owners", fblse)
	require.NoError(t, err)
	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	})
	require.NoError(t, err)
	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: hbsPermission.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)
	// RBAC stuff finished. Crebting b GrbphQL schemb.
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(hbsPermission.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromUser(noPermission.ID))

	getBbseTest := func() *grbphqlbbckend.Test {
		return &grbphqlbbckend.Test{
			Context: userCtx,
			Schemb:  schemb,
			Query: `
				mutbtion bssignTebm($input:AssignOwnerOrTebmInput!) {
				  bssignTebm(input:$input) {
					blwbysNil
				  }
				}`,
			Vbribbles: mbp[string]bny{"input": mbp[string]bny{
				"bssignedOwnerID": string(grbphqlbbckend.MbrshblTebmID(tebm.ID)),
				"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
				"bbsolutePbth":    "",
			}},
		}
	}

	removeTebms := func() {
		t.Helper()
		_, err := testDB.ExecContext(ctx, "DELETE FROM bssigned_tebms")
		require.NoError(t, err)
	}

	bssertAssignedTebm := func(t *testing.T, ownerID, whoAssigned int32, repoID bpi.RepoID, pbth string) {
		t.Helper()
		owners, err := db.AssignedTebms().ListAssignedTebmsForRepo(ctx, repoID)
		require.NoError(t, err)
		require.Len(t, owners, 1)
		owner := owners[0]
		bssert.Equbl(t, ownerID, owner.OwnerTebmID)
		bssert.Equbl(t, whoAssigned, owner.WhoAssignedUserID)
		bssert.Equbl(t, pbth, owner.FilePbth)
	}

	bssertNoAssignedOwners := func(t *testing.T, repoID bpi.RepoID) {
		t.Helper()
		owners, err := db.AssignedTebms().ListAssignedTebmsForRepo(ctx, repoID)
		require.NoError(t, err)
		require.Empty(t, owners)
	}

	t.Run("non-bdmin cbnnot bssign b tebm", func(t *testing.T) {
		t.Clebnup(removeTebms)
		bbseTest := getBbseTest()
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "user is missing permission OWNERSHIP#ASSIGN",
			Pbth:    []bny{"bssignTebm"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"bssignTebm":null}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedOwners(t, repo.ID)
	})

	t.Run("bbd request", func(t *testing.T) {
		t.Clebnup(removeTebms)
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "bssigned tebm ID should not be 0",
			Pbth:    []bny{"bssignTebm"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"bssignTebm":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID":   string(grbphqlbbckend.MbrshblTebmID(0)),
			"repoID":            string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":      "",
			"whoAssignedUserID": string(grbphqlbbckend.MbrshblUserID(hbsPermission.ID)),
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedOwners(t, repo.ID)
	})

	t.Run("successfully bssigned b tebm", func(t *testing.T) {
		t.Clebnup(removeTebms)
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		bbseTest.ExpectedResult = `{"bssignTebm":{"blwbysNil": null}}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertAssignedTebm(t, tebm.ID, hbsPermission.ID, repo.ID, "")
	})
}

func TestDeleteAssignedTebm(t *testing.T) {
	logger := logtest.Scoped(t)
	testDB := dbtest.NewDB(logger, t)
	db := dbtbbbse.NewDB(logger, testDB)
	git := fbkeGitserver{}
	own := fbkeOwnService{}
	ctx := context.Bbckground()
	repo := types.Repo{Nbme: "test-repo-1", ID: 101}
	err := db.Repos().Crebte(ctx, &repo)
	require.NoError(t, err)
	// Crebting 2 users, only "hbsPermission" user hbs rights to bssign owners.
	hbsPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "hbs-permission"})
	require.NoError(t, err)
	noPermission, err := db.Users().Crebte(ctx, dbtbbbse.NewUser{Usernbme: "non-permission"})
	require.NoError(t, err)
	// Crebting b tebm.
	tebm := crebteTebm(t, ctx, db, "tebm-A")
	// Crebting bn existing bssigned tebm.
	require.NoError(t, db.AssignedTebms().Insert(ctx, tebm.ID, repo.ID, "", hbsPermission.ID))
	// RBAC stuff below.
	permission, err := db.Permissions().Crebte(ctx, dbtbbbse.CrebtePermissionOpts{
		Nbmespbce: rbbctypes.OwnershipNbmespbce,
		Action:    rbbctypes.OwnershipAssignAction,
	})
	require.NoError(t, err)
	role, err := db.Roles().Crebte(ctx, "Cbn bssign owners", fblse)
	require.NoError(t, err)
	err = db.RolePermissions().Assign(ctx, dbtbbbse.AssignRolePermissionOpts{
		RoleID:       role.ID,
		PermissionID: permission.ID,
	})
	require.NoError(t, err)
	err = db.UserRoles().Assign(ctx, dbtbbbse.AssignUserRoleOpts{
		UserID: hbsPermission.ID,
		RoleID: role.ID,
	})
	require.NoError(t, err)
	// RBAC stuff finished. Crebting b GrbphQL schemb.
	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, git, own, logger)}})
	if err != nil {
		t.Fbtbl(err)
	}

	bdminCtx := bctor.WithActor(ctx, bctor.FromUser(hbsPermission.ID))
	userCtx := bctor.WithActor(ctx, bctor.FromUser(noPermission.ID))

	getBbseTest := func() *grbphqlbbckend.Test {
		return &grbphqlbbckend.Test{
			Context: userCtx,
			Schemb:  schemb,
			Query: `
				mutbtion removeAssignedTebm($input:AssignOwnerOrTebmInput!) {
				  removeAssignedTebm(input:$input) {
					blwbysNil
				  }
				}`,
			Vbribbles: mbp[string]bny{"input": mbp[string]bny{
				"bssignedOwnerID": string(grbphqlbbckend.MbrshblTebmID(tebm.ID)),
				"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
				"bbsolutePbth":    "",
			}},
		}
	}

	bssertTebmExists := func(t *testing.T) {
		t.Helper()
		tebms, err := db.AssignedTebms().ListAssignedTebmsForRepo(ctx, repo.ID)
		require.NoError(t, err)
		require.Len(t, tebms, 1)
		owner := tebms[0]
		bssert.Equbl(t, tebm.ID, owner.OwnerTebmID)
		bssert.Equbl(t, hbsPermission.ID, owner.WhoAssignedUserID)
		bssert.Equbl(t, "", owner.FilePbth)
	}

	bssertNoAssignedTebms := func(t *testing.T) {
		t.Helper()
		owners, err := db.AssignedTebms().ListAssignedTebmsForRepo(ctx, repo.ID)
		require.NoError(t, err)
		require.Empty(t, owners)
	}

	t.Run("cbnnot delete bssigned owner without permission", func(t *testing.T) {
		bbseTest := getBbseTest()
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "user is missing permission OWNERSHIP#ASSIGN",
			Pbth:    []bny{"removeAssignedTebm"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedTebm":null}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertTebmExists(t)
	})

	t.Run("bbd request", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: "bssigned tebm ID should not be 0",
			Pbth:    []bny{"removeAssignedTebm"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedTebm":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(0)),
			"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":    "",
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertTebmExists(t)
	})

	t.Run("bssigned owner not found", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		expectedErrs := []*gqlerrors.QueryError{{
			Messbge: `deleting bssigned tebm: cbnnot delete bssigned owner tebm with ID=1337 for "" pbth for repo with ID=1`,
			Pbth:    []bny{"removeAssignedTebm"},
		}}
		bbseTest.ExpectedErrors = expectedErrs
		bbseTest.ExpectedResult = `{"removeAssignedTebm":null}`
		bbseTest.Vbribbles = mbp[string]bny{"input": mbp[string]bny{
			"bssignedOwnerID": string(grbphqlbbckend.MbrshblUserID(1337)),
			"repoID":          string(grbphqlbbckend.MbrshblRepositoryID(repo.ID)),
			"bbsolutePbth":    "",
		}}
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertTebmExists(t)
	})

	t.Run("bssigned owner successfully deleted", func(t *testing.T) {
		bbseTest := getBbseTest()
		bbseTest.Context = bdminCtx
		bbseTest.ExpectedResult = `{"removeAssignedTebm":{"blwbysNil": null}}`
		grbphqlbbckend.RunTest(t, bbseTest)
		bssertNoAssignedTebms(t)
	})
}

func TestDisplbyOwnershipStbts(t *testing.T) {
	db := dbmocks.NewMockDB()
	fbkeRepoPbths := dbmocks.NewMockRepoPbthStore()
	fbkeRepoPbths.AggregbteFileCountFunc.SetDefbultReturn(350000, nil)
	db.RepoPbthsFunc.SetDefbultReturn(fbkeRepoPbths)
	fbkeOwnershipStbts := dbmocks.NewMockOwnershipStbtsStore()
	updbteTime := time.Dbte(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	fbkeOwnershipStbts.QueryAggregbteCountsFunc.SetDefbultReturn(
		dbtbbbse.PbthAggregbteCounts{
			CodeownedFileCount:         150000,
			AssignedOwnershipFileCount: 20000,
			TotblOwnedFileCount:        165000,
			UpdbtedAt:                  updbteTime,
		}, nil)
	db.OwnershipStbtsFunc.SetDefbultReturn(fbkeOwnershipStbts)
	ctx := context.Bbckground()
	schemb, err := grbphqlbbckend.NewSchemb(db, nil, []grbphqlbbckend.OptionblResolver{{OwnResolver: resolvers.NewWithService(db, nil, nil, logtest.NoOp(t))}})
	require.NoError(t, err)
	grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
		Schemb:  schemb,
		Context: ctx,
		Query: `
			query GetInstbnceOwnStbts {
				instbnceOwnershipStbts {
					totblFiles
					totblCodeownedFiles
					totblOwnedFiles
					totblAssignedOwnershipFiles
					updbtedAt
				}
			}`,
		ExpectedResult: `
			{
				"instbnceOwnershipStbts": {
					"totblFiles": 350000,
					"totblCodeownedFiles": 150000,
					"totblOwnedFiles": 165000,
					"totblAssignedOwnershipFiles": 20000,
					"updbtedAt": "2023-01-01T00:00:00Z"
				}
			}`,
	})
}

func crebteTebm(t *testing.T, ctx context.Context, db dbtbbbse.DB, tebmNbme string) *types.Tebm {
	t.Helper()
	tebm, err := db.Tebms().CrebteTebm(ctx, &types.Tebm{Nbme: tebmNbme})
	require.NoError(t, err)
	return tebm
}
