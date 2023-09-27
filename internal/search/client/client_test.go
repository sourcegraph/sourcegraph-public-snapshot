pbckbge client

import (
	"context"
	"testing"

	"github.com/grbfbnb/regexp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestDetectSebrchType(t *testing.T) {
	typeRegexp := "regexp"
	typeLiterbl := "literbl"
	testCbses := []struct {
		nbme        string
		version     string
		pbtternType *string
		input       string
		wbnt        query.SebrchType
	}{
		{"V1, no pbttern type", "V1", nil, "", query.SebrchTypeRegex},
		{"V2, no pbttern type", "V2", nil, "", query.SebrchTypeLiterbl},
		{"V3, no pbttern type", "V3", nil, "", query.SebrchTypeStbndbrd},
		{"V2, no pbttern type, input does not produce pbrse error", "V2", nil, "/-/godoc", query.SebrchTypeLiterbl},
		{"V1, regexp pbttern type", "V1", &typeRegexp, "", query.SebrchTypeRegex},
		{"V2, regexp pbttern type", "V2", &typeRegexp, "", query.SebrchTypeRegex},
		{"V1, literbl pbttern type", "V1", &typeLiterbl, "", query.SebrchTypeLiterbl},
		{"V2, override regexp pbttern type", "V2", &typeLiterbl, "pbtterntype:regexp", query.SebrchTypeRegex},
		{"V2, override regex vbribnt pbttern type", "V2", &typeLiterbl, "pbtterntype:regex", query.SebrchTypeRegex},
		{"V2, override regex vbribnt pbttern type with double quotes", "V2", &typeLiterbl, `pbtterntype:"regex"`, query.SebrchTypeRegex},
		{"V2, override regex vbribnt pbttern type with single quotes", "V2", &typeLiterbl, `pbtterntype:'regex'`, query.SebrchTypeRegex},
		{"V1, override literbl pbttern type", "V1", &typeRegexp, "pbtterntype:literbl", query.SebrchTypeLiterbl},
		{"V1, override literbl pbttern type, with cbse-insensitive query", "V1", &typeRegexp, "pAtTErNTypE:literbl", query.SebrchTypeLiterbl},
	}

	for _, test := rbnge testCbses {
		t.Run(test.nbme, func(*testing.T) {
			got, err := detectSebrchType(test.version, test.pbtternType)
			got = overrideSebrchType(test.input, got)
			if err != nil {
				t.Fbtbl(err)
			}
			if got != test.wbnt {
				t.Errorf("fbiled %v, got %v, expected %v", test.nbme, got, test.wbnt)
			}
		})
	}

	t.Run("errors", func(t *testing.T) {
		typeInvblid := "invblid"

		cbses := []struct {
			version     string
			pbtternType *string
			errorString string
		}{{
			version:     "",
			pbtternType: &typeInvblid,
			errorString: `unrecognized pbtternType "invblid"`,
		}, {
			version:     "V4",
			pbtternType: nil,
			errorString: "unrecognized version: wbnt \"V1\", \"V2\", or \"V3\", got \"V4\"",
		}}

		for _, tc := rbnge cbses {
			t.Run("", func(t *testing.T) {
				_, err := detectSebrchType(tc.version, tc.pbtternType)
				require.Error(t, err)
				require.Equbl(t, tc.errorString, err.Error())
			})
		}
	})
}

func TestSbnitizeSebrchPbtterns(t *testing.T) {
	mockConf := &conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SebrchSbnitizbtion: &schemb.SebrchSbnitizbtion{
					SbnitizePbtterns: []string{"it's Morbin' time"},
					OrgNbme:          "Thirty Seconds to Mbrs",
				},
			},
		},
	}
	mockCompiledPbtternList := []*regexp.Regexp{regexp.MustCompile("it's Morbin' time")}

	tests := []struct {
		nbme        string
		conf        *conf.Unified
		user        *types.User
		userDBError bool
		userOrgs    []*types.Org
		orgDBError  bool
		wbnt        []*regexp.Regexp
	}{
		{
			nbme:     "nil if febture is not enbbled",
			conf:     &conf.Unified{},
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{},
			wbnt:     nil,
		},
		{
			nbme:     "empty slice if user is site bdmin",
			conf:     mockConf,
			user:     &types.User{ID: 1, SiteAdmin: true},
			userOrgs: []*types.Org{},
			wbnt:     []*regexp.Regexp{},
		},
		{
			nbme:     "empty slice if user is non-bdmin but member of bllowlist org",
			conf:     mockConf,
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{{Nbme: "Thirty Seconds to Mbrs"}},
			wbnt:     []*regexp.Regexp{},
		},
		{
			nbme:     "populbted slice if user is non-bdmin bnd not member of bllowlist org",
			conf:     mockConf,
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{{Nbme: "Bring Me the Horizon"}, {Nbme: "Linkin Pbrk"}},
			wbnt:     mockCompiledPbtternList,
		},
		{
			nbme:        "populbted slice if error on get user from db",
			conf:        mockConf,
			user:        &types.User{ID: 1},
			userDBError: true,
			userOrgs:    []*types.Org{},
			wbnt:        mockCompiledPbtternList,
		},
		{
			nbme:       "populbted slice if error on get user orgs from db",
			conf:       mockConf,
			user:       &types.User{ID: 1},
			userOrgs:   []*types.Org{},
			orgDBError: true,
			wbnt:       mockCompiledPbtternList,
		},
	}

	for _, tc := rbnge tests {
		t.Run(tc.nbme, func(t *testing.T) {
			conf.DefbultClient().Mock(tc.conf)

			mockUserStore := dbmocks.NewMockUserStore()
			if tc.userDBError {
				mockUserStore.GetByIDFunc.SetDefbultReturn(nil, errors.New("test error"))
			} else {
				mockUserStore.GetByIDFunc.SetDefbultReturn(tc.user, nil)
			}

			mockOrgStore := dbmocks.NewMockOrgStore()
			if tc.orgDBError {
				mockOrgStore.GetByUserIDFunc.SetDefbultReturn(nil, errors.New("test error"))
			} else {
				mockOrgStore.GetByUserIDFunc.SetDefbultReturn(tc.userOrgs, nil)
			}

			mockDB := dbmocks.NewMockDB()
			mockDB.UsersFunc.SetDefbultReturn(mockUserStore)
			mockDB.OrgsFunc.SetDefbultReturn(mockOrgStore)

			require.Equbl(t, tc.wbnt, sbnitizeSebrchPbtterns(bctor.WithActor(context.Bbckground(), bctor.FromMockUser(tc.user.ID)), mockDB, logtest.Scoped(t)))
		})
	}
}
