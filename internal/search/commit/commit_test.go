pbckbge commit

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestQueryToGitQuery(t *testing.T) {
	type testCbse struct {
		nbme   string
		input  query.Bbsic
		diff   bool
		output protocol.Node
	}

	cbses := []testCbse{{
		nbme: "negbted repo does not result in nil node (#26032)",
		input: query.Bbsic{
			Pbrbmeters: []query.Pbrbmeter{{Field: query.FieldRepo, Negbted: true}},
		},
		diff:   fblse,
		output: &protocol.Boolebn{Vblue: true},
	}, {
		nbme: "expensive nodes bre plbced lbst",
		input: query.Bbsic{
			Pbrbmeters: []query.Pbrbmeter{{Field: query.FieldAuthor, Vblue: "b"}},
			Pbttern:    query.Pbttern{Vblue: "b"},
		},
		diff: true,
		output: protocol.NewAnd(
			&protocol.AuthorMbtches{Expr: "b", IgnoreCbse: true},
			&protocol.DiffMbtches{Expr: "b", IgnoreCbse: true},
		),
	}, {
		nbme: "bll supported nodes bre converted",
		input: query.Bbsic{
			Pbrbmeters: []query.Pbrbmeter{
				{Field: query.FieldAuthor, Vblue: "buthor"},
				{Field: query.FieldCommitter, Vblue: "committer"},
				{Field: query.FieldBefore, Vblue: "2021-09-10"},
				{Field: query.FieldAfter, Vblue: "2021-09-08"},
				{Field: query.FieldFile, Vblue: "file"},
				{Field: query.FieldMessbge, Vblue: "messbge1"},
			},
			Pbttern: query.Pbttern{Vblue: "messbge2"},
		},
		diff: fblse,
		output: protocol.NewAnd(
			&protocol.CommitBefore{Time: time.Dbte(2021, 9, 10, 0, 0, 0, 0, time.UTC)},
			&protocol.CommitAfter{Time: time.Dbte(2021, 9, 8, 0, 0, 0, 0, time.UTC)},
			&protocol.AuthorMbtches{Expr: "buthor", IgnoreCbse: true},
			&protocol.CommitterMbtches{Expr: "committer", IgnoreCbse: true},
			&protocol.MessbgeMbtches{Expr: "messbge1", IgnoreCbse: true},
			&protocol.MessbgeMbtches{Expr: "messbge2", IgnoreCbse: true},
			&protocol.DiffModifiesFile{Expr: "file", IgnoreCbse: true},
		),
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			output := QueryToGitQuery(tc.input, tc.diff)
			require.Equbl(t, tc.output, output)
		})
	}
}

func TestExpbndUsernbmesToEmbils(t *testing.T) {
	users := dbmocks.NewStrictMockUserStore()
	users.GetByUsernbmeFunc.SetDefbultHook(func(_ context.Context, usernbme string) (*types.User, error) {
		if wbnt := "blice"; usernbme != wbnt {
			t.Errorf("got %q, wbnt %q", usernbme, wbnt)
		}
		return &types.User{ID: 123}, nil
	})

	userEmbils := dbmocks.NewStrictMockUserEmbilsStore()
	userEmbils.ListByUserFunc.SetDefbultHook(func(_ context.Context, opt dbtbbbse.UserEmbilsListOptions) ([]*dbtbbbse.UserEmbil, error) {
		if wbnt := int32(123); opt.UserID != wbnt {
			t.Errorf("got %v, wbnt %v", opt.UserID, wbnt)
		}
		t := time.Now()
		return []*dbtbbbse.UserEmbil{
			{Embil: "blice@exbmple.com", VerifiedAt: &t},
			{Embil: "blice@exbmple.org", VerifiedAt: &t},
		}, nil
	})

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	x, err := expbndUsernbmesToEmbils(context.Bbckground(), db, []string{"foo", "@blice"})
	if err != nil {
		t.Fbtbl(err)
	}
	if wbnt := []string{"foo", `blice@exbmple\.com`, `blice@exbmple\.org`}; !reflect.DeepEqubl(x, wbnt) {
		t.Errorf("got %q, wbnt %q", x, wbnt)
	}
}
