package providers

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/schema"
)

func stringPointer(s string) *string {
	return &s
}

func TestGetAuthProviderCommon(t *testing.T) {
	testCases := []struct {
		name     string
		provider Provider
		want     schema.AuthProviderCommon
	}{
		{
			name: "all config fields are defined",
			provider: MockAuthProvider{
				MockConfig: schema.AuthProviderCommon{
					Hidden:        true,
					Order:         1,
					DisplayName:   "Mock Provider",
					DisplayPrefix: stringPointer("Mock"),
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        true,
				Order:         1,
				DisplayName:   "Mock Provider",
				DisplayPrefix: stringPointer("Mock"),
			},
		},
		{
			name: "DisplayPrefix is zero value",
			provider: MockAuthProvider{
				MockConfig: schema.AuthProviderCommon{
					Hidden:      false,
					Order:       2,
					DisplayName: "Another Mock",
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         2,
				DisplayName:   "Another Mock",
				DisplayPrefix: nil,
			},
		},
		{
			name: "DisplayName is zero value",
			provider: MockAuthProvider{
				MockConfig: schema.AuthProviderCommon{
					Hidden: false,
					Order:  2,
				},
				MockConfigID: ConfigID{
					Type: "mocked provider",
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         2,
				DisplayName:   "mocked provider",
				DisplayPrefix: nil,
			},
		},
		{
			name: "Hidden is zero value",
			provider: MockAuthProvider{
				MockConfig: schema.AuthProviderCommon{
					Order: 2,
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         2,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
		{
			name: "All parameters are zero value",
			provider: MockAuthProvider{
				MockConfig: schema.AuthProviderCommon{},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         0,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
		{
			name:     "Works without a config",
			provider: MockAuthProvider{},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         0,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
		{
			name: "Works with BuiltinAuthProvider",
			provider: MockAuthProvider{
				MockAuthProvidersConfig: &schema.AuthProviders{
					Builtin: &schema.BuiltinAuthProvider{
						Type: "builtin",
					},
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         0,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
		{
			name: "Works with HttpHeaderAuthProvider",
			provider: MockAuthProvider{
				MockAuthProvidersConfig: &schema.AuthProviders{
					HttpHeader: &schema.HTTPHeaderAuthProvider{
						Type: "http-header",
					},
				},
			},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         0,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			have := GetAuthProviderCommon(tc.provider)
			if !reflect.DeepEqual(have, tc.want) {
				t.Errorf("have %+v, want %+v", have, tc.want)
			}
		})
	}
}

func TestSortedProviders(t *testing.T) {
	tests := []struct {
		name          string
		input         []Provider
		expectedOrder []int
	}{
		{
			name: "sort works as expected",
			input: []Provider{
				MockAuthProvider{MockConfigID: ConfigID{Type: "a", ID: "1"}, MockConfig: schema.AuthProviderCommon{Order: 2}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "b", ID: "2"}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "builtin", ID: "3"}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "c", ID: "4"}, MockConfig: schema.AuthProviderCommon{Order: 1}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "d", ID: "5"}, MockConfig: schema.AuthProviderCommon{Order: 1}},
				MockAuthProvider{MockConfigID: ConfigID{Type: "b", ID: "6"}, MockConfig: schema.AuthProviderCommon{Order: 1}},
			},
			expectedOrder: []int{3, 0, 2, 1, 5, 4},
		},
		{
			name:          "Behaves well for empty slice",
			input:         []Provider{},
			expectedOrder: []int{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			MockProviders = test.input
			t.Cleanup(func() {
				MockProviders = nil
			})

			sorted := SortedProviders()
			expected := make([]Provider, len(sorted))
			for i, order := range test.expectedOrder {
				expected[i] = test.input[order]
			}
			require.ElementsMatch(t, expected, sorted)
		})
	}
}
