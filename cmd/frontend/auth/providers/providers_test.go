package providers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type MockAuthProvider struct {
	configID ConfigID
	config   schema.AuthProviderCommon
}

func (m MockAuthProvider) ConfigID() ConfigID {
	return m.configID
}

func (m MockAuthProvider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Github: &schema.GitHubAuthProvider{
			Type: m.configID.Type,
		},
	}
}

func (m MockAuthProvider) CachedInfo() *Info {
	panic("should not be called")

	// return &providers.Info{ServiceID: m.serviceID}
}

func (m MockAuthProvider) Refresh(ctx context.Context) error {
	panic("should not be called")
}

type mockAuthProviderUser struct {
	Username string `json:"username,omitempty"`
	ID       int32  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
}

func (m MockAuthProvider) ExternalAccountInfo(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	data, err := encryption.DecryptJSON[mockAuthProviderUser](ctx, account.AccountData.Data)
	if err != nil {
		return nil, err
	}

	return &extsvc.PublicAccountData{
		Login:       &data.Username,
		DisplayName: &data.Name,
	}, nil
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
				MockAuthProvider{configID: ConfigID{Type: "a", ID: "1"}, config: schema.AuthProviderCommon{Order: 2}},
				MockAuthProvider{configID: ConfigID{Type: "b", ID: "2"}},
				MockAuthProvider{configID: ConfigID{Type: "builtin", ID: "3"}},
				MockAuthProvider{configID: ConfigID{Type: "c", ID: "4"}, config: schema.AuthProviderCommon{Order: 1}},
				MockAuthProvider{configID: ConfigID{Type: "d", ID: "5"}, config: schema.AuthProviderCommon{Order: 1}},
				MockAuthProvider{configID: ConfigID{Type: "b", ID: "6"}, config: schema.AuthProviderCommon{Order: 1}},
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
