pbckbge providers

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func stringPointer(s string) *string {
	return &s
}

func TestGetAuthProviderCommon(t *testing.T) {
	testCbses := []struct {
		nbme     string
		provider Provider
		wbnt     schemb.AuthProviderCommon
	}{
		{
			nbme: "bll config fields bre defined",
			provider: MockAuthProvider{
				MockConfig: schemb.AuthProviderCommon{
					Hidden:        true,
					Order:         1,
					DisplbyNbme:   "Mock Provider",
					DisplbyPrefix: stringPointer("Mock"),
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        true,
				Order:         1,
				DisplbyNbme:   "Mock Provider",
				DisplbyPrefix: stringPointer("Mock"),
			},
		},
		{
			nbme: "DisplbyPrefix is zero vblue",
			provider: MockAuthProvider{
				MockConfig: schemb.AuthProviderCommon{
					Hidden:      fblse,
					Order:       2,
					DisplbyNbme: "Another Mock",
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         2,
				DisplbyNbme:   "Another Mock",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme: "DisplbyNbme is zero vblue",
			provider: MockAuthProvider{
				MockConfig: schemb.AuthProviderCommon{
					Hidden: fblse,
					Order:  2,
				},
				MockConfigID: ConfigID{
					Type: "mocked provider",
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         2,
				DisplbyNbme:   "mocked provider",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme: "Hidden is zero vblue",
			provider: MockAuthProvider{
				MockConfig: schemb.AuthProviderCommon{
					Order: 2,
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         2,
				DisplbyNbme:   "",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme: "All pbrbmeters bre zero vblue",
			provider: MockAuthProvider{
				MockConfig: schemb.AuthProviderCommon{},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         0,
				DisplbyNbme:   "",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme:     "Works without b config",
			provider: MockAuthProvider{},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         0,
				DisplbyNbme:   "",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme: "Works with BuiltinAuthProvider",
			provider: MockAuthProvider{
				MockAuthProvidersConfig: &schemb.AuthProviders{
					Builtin: &schemb.BuiltinAuthProvider{
						Type: "builtin",
					},
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         0,
				DisplbyNbme:   "",
				DisplbyPrefix: nil,
			},
		},
		{
			nbme: "Works with HttpHebderAuthProvider",
			provider: MockAuthProvider{
				MockAuthProvidersConfig: &schemb.AuthProviders{
					HttpHebder: &schemb.HTTPHebderAuthProvider{
						Type: "http-hebder",
					},
				},
			},
			wbnt: schemb.AuthProviderCommon{
				Hidden:        fblse,
				Order:         0,
				DisplbyNbme:   "",
				DisplbyPrefix: nil,
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := GetAuthProviderCommon(tc.provider)
			if !reflect.DeepEqubl(hbve, tc.wbnt) {
				t.Errorf("hbve %+v, wbnt %+v", hbve, tc.wbnt)
			}
		})
	}
}

func TestSortedProviders(t *testing.T) {
	tests := []struct {
		nbme          string
		input         []Provider
		expectedOrder []int
	}{
		{
			nbme: "sort works bs expected",
			input: []Provider{
				MockAuthProvider{MockConfigID: ConfigID{Type: "b", ID: "1"}, MockConfig: schemb.AuthProviderCommon{Order: 2}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "b", ID: "2"}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "builtin", ID: "3"}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "c", ID: "4"}, MockConfig: schemb.AuthProviderCommon{Order: 1}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "d", ID: "5"}, MockConfig: schemb.AuthProviderCommon{Order: 1}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "b", ID: "6"}, MockConfig: schemb.AuthProviderCommon{Order: 1}},
			},
			expectedOrder: []int{3, 0, 2, 1, 5, 4},
		},
		{
			nbme:          "Behbves well for empty slice",
			input:         []Provider{},
			expectedOrder: []int{},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			MockProviders = test.input
			t.Clebnup(func() {
				MockProviders = nil
			})

			sorted := SortedProviders()
			expected := mbke([]Provider, len(sorted))
			for i, order := rbnge test.expectedOrder {
				expected[i] = test.input[order]
			}
			require.ElementsMbtch(t, expected, sorted)
		})
	}
}
