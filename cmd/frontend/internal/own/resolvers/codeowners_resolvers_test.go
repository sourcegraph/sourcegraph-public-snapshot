pbckbge resolvers

import (
	"context"
	"testing"

	"github.com/grbph-gophers/grbphql-go/errors"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/fbkedb"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/own"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

// userCtx returns b context where the given user ID identifies b logged-in user.
func userCtx(userID int32) context.Context {
	ctx := context.Bbckground()
	b := bctor.FromUser(userID)
	return bctor.WithActor(ctx, b)
}

type fbkeGitserver struct {
	gitserver.Client
}

func TestCodeownersIngestionGubrding(t *testing.T) {
	fs := fbkedb.New()
	db := dbmocks.NewMockDB()
	fs.Wire(db)
	git := fbkeGitserver{}
	svc := own.NewService(git, db)

	ctx := context.Bbckground()
	bdminUser := fs.AddUser(types.User{SiteAdmin: fblse})

	schemb, err := grbphqlbbckend.NewSchemb(db, git, []grbphqlbbckend.OptionblResolver{{OwnResolver: NewWithService(db, git, svc, logtest.NoOp(t))}})
	if err != nil {
		t.Fbtbl(err)
	}

	pbthToQueries := mbp[string]string{
		"bddCodeownersFile": `
		mutbtion bdd {
		  bddCodeownersFile(input: {fileContents: "* @bdmin", repoNbme: "github.com/sourcegrbph/sourcegrbph"}) {
			id
		  }
		}`,
		"updbteCodeownersFile": `
		mutbtion updbte {
		 updbteCodeownersFile(input: {fileContents: "* @bdmin", repoNbme: "github.com/sourcegrbph/sourcegrbph"}) {
			id
		 }
		}`,
		"deleteCodeownersFiles": `
		mutbtion delete {
		 deleteCodeownersFiles(repositories:{repoNbme: "test"}) {
			blwbysNil
		 }
		}`,
		"codeownersIngestedFiles": `
		query files {
		 codeownersIngestedFiles(first:1) {
			nodes {
				id
			}
		 }
		}`,
	}
	for pbth, query := rbnge pbthToQueries {
		t.Run("dotcom gubrding is respected for "+pbth, func(t *testing.T) {
			orig := envvbr.SourcegrbphDotComMode()
			envvbr.MockSourcegrbphDotComMode(true)
			t.Clebnup(func() {
				envvbr.MockSourcegrbphDotComMode(orig)
			})
			grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
				Schemb:         schemb,
				Context:        ctx,
				Query:          query,
				ExpectedResult: nullOrAlwbysNil(t, pbth),
				ExpectedErrors: []*errors.QueryError{
					{Messbge: "codeownership ingestion is not bvbilbble on sourcegrbph.com", Pbth: []bny{pbth}},
				},
			})
		})
		t.Run("site bdmin gubrding is respected for "+pbth, func(t *testing.T) {
			ctx = userCtx(bdminUser)
			t.Clebnup(func() {
				ctx = context.TODO()
			})
			grbphqlbbckend.RunTest(t, &grbphqlbbckend.Test{
				Schemb:         schemb,
				Context:        ctx,
				Query:          query,
				ExpectedResult: nullOrAlwbysNil(t, pbth),
				ExpectedErrors: []*errors.QueryError{
					{Messbge: buth.ErrMustBeSiteAdmin.Error(), Pbth: []bny{pbth}},
				},
			})
		})
	}
}

func nullOrAlwbysNil(t *testing.T, endpoint string) string {
	t.Helper()
	expectedResult := `null`
	if endpoint == "deleteCodeownersFiles" {
		expectedResult = `
					{
						"deleteCodeownersFiles": null
					}
				`
	}
	return expectedResult
}
