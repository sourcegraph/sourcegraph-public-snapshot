package run

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

func TestDetectSearchType(t *testing.T) {
	typeRegexp := "regexp"
	typeLiteral := "literal"
	testCases := []struct {
		name        string
		version     string
		patternType *string
		input       string
		want        query.SearchType
	}{
		{"V1, no pattern type", "V1", nil, "", query.SearchTypeRegex},
		{"V2, no pattern type", "V2", nil, "", query.SearchTypeLiteralDefault},
		{"V2, no pattern type, input does not produce parse error", "V2", nil, "/-/godoc", query.SearchTypeLiteralDefault},
		{"V1, regexp pattern type", "V1", &typeRegexp, "", query.SearchTypeRegex},
		{"V2, regexp pattern type", "V2", &typeRegexp, "", query.SearchTypeRegex},
		{"V1, literal pattern type", "V1", &typeLiteral, "", query.SearchTypeLiteralDefault},
		{"V2, override regexp pattern type", "V2", &typeLiteral, "patterntype:regexp", query.SearchTypeRegex},
		{"V2, override regex variant pattern type", "V2", &typeLiteral, "patterntype:regex", query.SearchTypeRegex},
		{"V2, override regex variant pattern type with double quotes", "V2", &typeLiteral, `patterntype:"regex"`, query.SearchTypeRegex},
		{"V2, override regex variant pattern type with single quotes", "V2", &typeLiteral, `patterntype:'regex'`, query.SearchTypeRegex},
		{"V1, override literal pattern type", "V1", &typeRegexp, "patterntype:literal", query.SearchTypeLiteralDefault},
		{"V1, override literal pattern type, with case-insensitive query", "V1", &typeRegexp, "pAtTErNTypE:literal", query.SearchTypeLiteralDefault},
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
			version:     "V3",
			patternType: nil,
			errorString: `unrecognized version: want "V1" or "V2", got "V3"`,
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
