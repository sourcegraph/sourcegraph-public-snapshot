package client

import (
	"context"
	"testing"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDetectSearchType(t *testing.T) {
	typeRegexp := "regexp"
	typeLiteral := "literal"
	typeLucky := "lucky"
	testCases := []struct {
		name        string
		version     string
		patternType *string
		input       string
		want        query.SearchType
	}{
		{"V1, no pattern type", "V1", nil, "", query.SearchTypeRegex},
		{"V2, no pattern type", "V2", nil, "", query.SearchTypeLiteral},
		{"V3, no pattern type", "V3", nil, "", query.SearchTypeStandard},
		{"V2, no pattern type, input does not produce parse error", "V2", nil, "/-/godoc", query.SearchTypeLiteral},
		{"V1, regexp pattern type", "V1", &typeRegexp, "", query.SearchTypeRegex},
		{"V2, regexp pattern type", "V2", &typeRegexp, "", query.SearchTypeRegex},
		{"V1, literal pattern type", "V1", &typeLiteral, "", query.SearchTypeLiteral},
		{"V2, override regexp pattern type", "V2", &typeLiteral, "patterntype:regexp", query.SearchTypeRegex},
		{"V2, override regex variant pattern type", "V2", &typeLiteral, "patterntype:regex", query.SearchTypeRegex},
		{"V2, override regex variant pattern type with double quotes", "V2", &typeLiteral, `patterntype:"regex"`, query.SearchTypeRegex},
		{"V2, override regex variant pattern type with single quotes", "V2", &typeLiteral, `patterntype:'regex'`, query.SearchTypeRegex},
		{"V1, override literal pattern type", "V1", &typeRegexp, "patterntype:literal", query.SearchTypeLiteral},
		{"V1, override literal pattern type, with case-insensitive query", "V1", &typeRegexp, "pAtTErNTypE:literal", query.SearchTypeLiteral},
		{"V1, lucky pattern type should be mapped to standard", "V1", &typeLucky, "", query.SearchTypeStandard},
		{"V1, lucky pattern type in query should be mapped to standard", "V1", nil, "patternType:lucky", query.SearchTypeStandard},
	}

	for _, test := range testCases {
		t.Run(test.name, func(*testing.T) {
			got, err := detectSearchType(test.version, test.patternType)
			got = overrideSearchType(test.input, got)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Errorf("failed %v, got %v, expected %v", test.name, got, test.want)
			}
		})
	}

	t.Run("errors", func(t *testing.T) {
		typeInvalid := "invalid"

		cases := []struct {
			version     string
			patternType *string
			errorString string
		}{{
			version:     "",
			patternType: &typeInvalid,
			errorString: `unrecognized patternType "invalid"`,
		}, {
			version:     "V99",
			patternType: nil,
			errorString: "unrecognized version: want \"V1\", \"V2\", or \"V3\", got \"V99\"",
		}}

		for _, tc := range cases {
			t.Run("", func(t *testing.T) {
				_, err := detectSearchType(tc.version, tc.patternType)
				require.Error(t, err)
				require.Equal(t, tc.errorString, err.Error())
			})
		}
	})
}

func TestSanitizeSearchPatterns(t *testing.T) {
	mockConf := &conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				SearchSanitization: &schema.SearchSanitization{
					SanitizePatterns: []string{"it's Morbin' time"},
					OrgName:          "Thirty Seconds to Mars",
				},
			},
		},
	}
	mockCompiledPatternList := []*regexp.Regexp{regexp.MustCompile("it's Morbin' time")}

	tests := []struct {
		name        string
		conf        *conf.Unified
		user        *types.User
		userDBError bool
		userOrgs    []*types.Org
		orgDBError  bool
		want        []*regexp.Regexp
	}{
		{
			name:     "nil if feature is not enabled",
			conf:     &conf.Unified{},
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{},
			want:     nil,
		},
		{
			name:     "empty slice if user is site admin",
			conf:     mockConf,
			user:     &types.User{ID: 1, SiteAdmin: true},
			userOrgs: []*types.Org{},
			want:     []*regexp.Regexp{},
		},
		{
			name:     "empty slice if user is non-admin but member of allowlist org",
			conf:     mockConf,
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{{Name: "Thirty Seconds to Mars"}},
			want:     []*regexp.Regexp{},
		},
		{
			name:     "populated slice if user is non-admin and not member of allowlist org",
			conf:     mockConf,
			user:     &types.User{ID: 1},
			userOrgs: []*types.Org{{Name: "Bring Me the Horizon"}, {Name: "Linkin Park"}},
			want:     mockCompiledPatternList,
		},
		{
			name:        "populated slice if error on get user from db",
			conf:        mockConf,
			user:        &types.User{ID: 1},
			userDBError: true,
			userOrgs:    []*types.Org{},
			want:        mockCompiledPatternList,
		},
		{
			name:       "populated slice if error on get user orgs from db",
			conf:       mockConf,
			user:       &types.User{ID: 1},
			userOrgs:   []*types.Org{},
			orgDBError: true,
			want:       mockCompiledPatternList,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conf.DefaultClient().Mock(tc.conf)

			mockUserStore := dbmocks.NewMockUserStore()
			if tc.userDBError {
				mockUserStore.GetByIDFunc.SetDefaultReturn(nil, errors.New("test error"))
			} else {
				mockUserStore.GetByIDFunc.SetDefaultReturn(tc.user, nil)
			}

			mockOrgStore := dbmocks.NewMockOrgStore()
			if tc.orgDBError {
				mockOrgStore.GetByUserIDFunc.SetDefaultReturn(nil, errors.New("test error"))
			} else {
				mockOrgStore.GetByUserIDFunc.SetDefaultReturn(tc.userOrgs, nil)
			}

			mockDB := dbmocks.NewMockDB()
			mockDB.UsersFunc.SetDefaultReturn(mockUserStore)
			mockDB.OrgsFunc.SetDefaultReturn(mockOrgStore)

			require.Equal(t, tc.want, sanitizeSearchPatterns(actor.WithActor(context.Background(), actor.FromMockUser(tc.user.ID)), mockDB, logtest.Scoped(t)))
		})
	}
}
