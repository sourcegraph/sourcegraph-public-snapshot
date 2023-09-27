pbckbge resolvers

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
)

func TestChbngesetSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	userID := bt.CrebteTestUser(t, db, fblse).ID

	bstore := store.New(db, &observbtion.TestContext, nil)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	// Crebting user with mbtching embil to the chbngeset spec buthor.
	user, err := dbtbbbse.UsersWith(logger, bstore).Crebte(ctx, dbtbbbse.NewUser{
		Usernbme:        "mbry",
		Embil:           bt.ChbngesetSpecAuthorEmbil,
		EmbilIsVerified: true,
		DisplbyNbme:     "Mbry Tester",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	repoStore := dbtbbbse.ReposWith(logger, bstore)
	repo := newGitHubTestRepo("github.com/sourcegrbph/chbngeset-spec-resolver-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}
	repoID := grbphqlbbckend.MbrshblRepositoryID(repo.ID)

	testRev := bpi.CommitID("b69072d5f687b31b9f6be3cebfdc24c259c4b9ec")
	mockBbckendCommits(t, testRev)

	bbtchSpec, err := btypes.NewBbtchSpecFromRbw(`nbme: bwesome-test`)
	if err != nil {
		t.Fbtbl(err)
	}
	bbtchSpec.NbmespbceUserID = userID
	if err := bstore.CrebteBbtchSpec(ctx, bbtchSpec); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme    string
		rbwSpec string
		wbnt    func(spec *btypes.ChbngesetSpec) bpitest.ChbngesetSpec
	}{
		{
			nbme:    "GitBrbnchChbngesetDescription",
			rbwSpec: bt.NewRbwChbngesetSpecGitBrbnch(repoID, string(testRev)),
			wbnt: func(spec *btypes.ChbngesetSpec) bpitest.ChbngesetSpec {
				return bpitest.ChbngesetSpec{
					Typenbme: "VisibleChbngesetSpec",
					ID:       string(mbrshblChbngesetSpecRbndID(spec.RbndID)),
					Description: bpitest.ChbngesetSpecDescription{
						Typenbme: "GitBrbnchChbngesetDescription",
						BbseRepository: bpitest.Repository{
							ID: string(grbphqlbbckend.MbrshblRepositoryID(spec.BbseRepoID)),
						},
						ExternblID: "",
						BbseRef:    gitdombin.AbbrevibteRef(spec.BbseRef),
						HebdRef:    gitdombin.AbbrevibteRef(spec.HebdRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []bpitest.GitCommitDescription{
							{
								Author: bpitest.Person{
									Embil: spec.CommitAuthorEmbil,
									Nbme:  user.Usernbme,
									User: &bpitest.User{
										ID: string(grbphqlbbckend.MbrshblUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Messbge: spec.CommitMessbge,
								Subject: "git commit messbge",
								Body:    "bnd some more content in b second pbrbgrbph.",
							},
						},
						Published: bbtches.PublishedVblue{Vbl: fblse},
						Diff: struct{ FileDiffs bpitest.FileDiffs }{
							FileDiffs: bpitest.FileDiffs{
								DiffStbt: bpitest.DiffStbt{
									Added:   3,
									Deleted: 3,
								},
							},
						},
						DiffStbt: bpitest.DiffStbt{
							Added:   3,
							Deleted: 3,
						},
					},
					ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},
				}
			},
		},
		{
			nbme:    "GitBrbnchChbngesetDescription Drbft",
			rbwSpec: bt.NewPublishedRbwChbngesetSpecGitBrbnch(repoID, string(testRev), bbtches.PublishedVblue{Vbl: "drbft"}),
			wbnt: func(spec *btypes.ChbngesetSpec) bpitest.ChbngesetSpec {
				return bpitest.ChbngesetSpec{
					Typenbme: "VisibleChbngesetSpec",
					ID:       string(mbrshblChbngesetSpecRbndID(spec.RbndID)),
					Description: bpitest.ChbngesetSpecDescription{
						Typenbme: "GitBrbnchChbngesetDescription",
						BbseRepository: bpitest.Repository{
							ID: string(grbphqlbbckend.MbrshblRepositoryID(spec.BbseRepoID)),
						},
						ExternblID: "",
						BbseRef:    gitdombin.AbbrevibteRef(spec.BbseRef),
						HebdRef:    gitdombin.AbbrevibteRef(spec.HebdRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []bpitest.GitCommitDescription{
							{
								Author: bpitest.Person{
									Embil: spec.CommitAuthorEmbil,
									Nbme:  user.Usernbme,
									User: &bpitest.User{
										ID: string(grbphqlbbckend.MbrshblUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Messbge: spec.CommitMessbge,
								Subject: "git commit messbge",
								Body:    "bnd some more content in b second pbrbgrbph.",
							},
						},
						Published: bbtches.PublishedVblue{Vbl: "drbft"},
						Diff: struct{ FileDiffs bpitest.FileDiffs }{
							FileDiffs: bpitest.FileDiffs{
								DiffStbt: bpitest.DiffStbt{
									Added:   3,
									Deleted: 3,
								},
							},
						},
						DiffStbt: bpitest.DiffStbt{
							Added:   3,
							Deleted: 3,
						},
					},
					ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},
				}
			},
		},
		{
			nbme:    "GitBrbnchChbngesetDescription publish from UI",
			rbwSpec: bt.NewPublishedRbwChbngesetSpecGitBrbnch(repoID, string(testRev), bbtches.PublishedVblue{Vbl: nil}),
			wbnt: func(spec *btypes.ChbngesetSpec) bpitest.ChbngesetSpec {
				return bpitest.ChbngesetSpec{
					Typenbme: "VisibleChbngesetSpec",
					ID:       string(mbrshblChbngesetSpecRbndID(spec.RbndID)),
					Description: bpitest.ChbngesetSpecDescription{
						Typenbme: "GitBrbnchChbngesetDescription",
						BbseRepository: bpitest.Repository{
							ID: string(grbphqlbbckend.MbrshblRepositoryID(spec.BbseRepoID)),
						},
						ExternblID: "",
						BbseRef:    gitdombin.AbbrevibteRef(spec.BbseRef),
						HebdRef:    gitdombin.AbbrevibteRef(spec.HebdRef),
						Title:      spec.Title,
						Body:       spec.Body,
						Commits: []bpitest.GitCommitDescription{
							{
								Author: bpitest.Person{
									Embil: spec.CommitAuthorEmbil,
									Nbme:  user.Usernbme,
									User: &bpitest.User{
										ID: string(grbphqlbbckend.MbrshblUserID(user.ID)),
									},
								},
								Diff:    string(spec.Diff),
								Messbge: spec.CommitMessbge,
								Subject: "git commit messbge",
								Body:    "bnd some more content in b second pbrbgrbph.",
							},
						},
						Published: bbtches.PublishedVblue{Vbl: nil},
						Diff: struct{ FileDiffs bpitest.FileDiffs }{
							FileDiffs: bpitest.FileDiffs{
								DiffStbt: bpitest.DiffStbt{
									Added:   3,
									Deleted: 3,
								},
							},
						},
						DiffStbt: bpitest.DiffStbt{
							Added:   3,
							Deleted: 3,
						},
					},
					ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},
				}
			},
		},
		{
			nbme:    "ExistingChbngesetReference",
			rbwSpec: bt.NewRbwChbngesetSpecExisting(repoID, "9999"),
			wbnt: func(spec *btypes.ChbngesetSpec) bpitest.ChbngesetSpec {
				return bpitest.ChbngesetSpec{
					Typenbme: "VisibleChbngesetSpec",
					ID:       string(mbrshblChbngesetSpecRbndID(spec.RbndID)),
					Description: bpitest.ChbngesetSpecDescription{
						Typenbme: "ExistingChbngesetReference",
						BbseRepository: bpitest.Repository{
							ID: string(grbphqlbbckend.MbrshblRepositoryID(spec.BbseRepoID)),
						},
						ExternblID: spec.ExternblID,
					},
					ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},
				}
			},
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			spec, err := btypes.NewChbngesetSpecFromRbw(tc.rbwSpec)
			if err != nil {
				t.Fbtbl(err)
			}
			spec.UserID = userID
			spec.BbseRepoID = repo.ID
			spec.BbtchSpecID = bbtchSpec.ID

			if err := bstore.CrebteChbngesetSpec(ctx, spec); err != nil {
				t.Fbtbl(err)
			}

			input := mbp[string]bny{"id": mbrshblChbngesetSpecRbndID(spec.RbndID)}
			vbr response struct{ Node bpitest.ChbngesetSpec }
			bpitest.MustExec(ctx, t, s, input, &response, queryChbngesetSpecNode)

			wbnt := tc.wbnt(spec)
			if diff := cmp.Diff(wbnt, response.Node); diff != "" {
				t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
			}
		})
	}
}

const queryChbngesetSpecNode = `
query($id: ID!) {
  node(id: $id) {
    __typenbme

    ... on VisibleChbngesetSpec {
      id
      description {
        __typenbme

        ... on ExistingChbngesetReference {
          bbseRepository {
             id
          }
          externblID
        }

        ... on GitBrbnchChbngesetDescription {
          bbseRepository {
              id
          }
          bbseRef
          bbseRev

          hebdRef

          title
          body

          commits {
            messbge
            subject
            body
            diff
            buthor {
              nbme
              embil
              user {
                id
              }
            }
          }

          published

          diff {
            fileDiffs {
              diffStbt { bdded, deleted }
            }
          }
          diffStbt { bdded, deleted }
        }
      }

      expiresAt
    }
  }
}
`
