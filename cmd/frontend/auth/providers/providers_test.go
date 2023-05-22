package providers

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type mockAuthProvider struct {
	configID            ConfigID
	config              schema.AuthProviderCommon
	authProvidersConfig *schema.AuthProviders
}

func (m mockAuthProvider) ConfigID() ConfigID {
	return m.configID
}

func (m mockAuthProvider) Config() schema.AuthProviders {
	if m.authProvidersConfig != nil {
		return *m.authProvidersConfig
	}

	return schema.AuthProviders{
		Github: &schema.GitHubAuthProvider{
			Type:          m.configID.Type,
			DisplayName:   m.config.DisplayName,
			DisplayPrefix: m.config.DisplayPrefix,
			Hidden:        m.config.Hidden,
			Order:         m.config.Order,
		},
	}
}

func (m mockAuthProvider) CachedInfo() *Info {
	panic("should not be called")
}

func (m mockAuthProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

func (m mockAuthProvider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	panic("should not be called")
}

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
			provider: mockAuthProvider{
				config: schema.AuthProviderCommon{
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
			provider: mockAuthProvider{
				config: schema.AuthProviderCommon{
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
			provider: mockAuthProvider{
				config: schema.AuthProviderCommon{
					Hidden: false,
					Order:  2,
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
			name: "Hidden is zero value",
			provider: mockAuthProvider{
				config: schema.AuthProviderCommon{
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
			provider: mockAuthProvider{
				config: schema.AuthProviderCommon{},
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
			provider: mockAuthProvider{},
			want: schema.AuthProviderCommon{
				Hidden:        false,
				Order:         0,
				DisplayName:   "",
				DisplayPrefix: nil,
			},
		},
		{
			name: "Works with BuiltinAuthProvider",
			provider: mockAuthProvider{
				authProvidersConfig: &schema.AuthProviders{
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
			provider: mockAuthProvider{
				authProvidersConfig: &schema.AuthProviders{
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
				mockAuthProvider{configID: ConfigID{Type: "a", ID: "1"}, config: schema.AuthProviderCommon{Order: 2}},
				mockAuthProvider{configID: ConfigID{Type: "b", ID: "2"}},
				mockAuthProvider{configID: ConfigID{Type: "builtin", ID: "3"}},
				mockAuthProvider{configID: ConfigID{Type: "c", ID: "4"}, config: schema.AuthProviderCommon{Order: 1}},
				mockAuthProvider{configID: ConfigID{Type: "d", ID: "5"}, config: schema.AuthProviderCommon{Order: 1}},
				mockAuthProvider{configID: ConfigID{Type: "b", ID: "6"}, config: schema.AuthProviderCommon{Order: 1}},
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
