package license

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractLicenseKeyBasedAccessTokenContents(t *testing.T) {
	for _, tc := range []struct {
		name         string
		token        string
		wantContents autogold.Value
		wantError    autogold.Value
	}{
		{
			name:         "from real token",
			token:        GenerateLicenseKeyBasedAccessToken("key"),
			wantContents: autogold.Expect(`,p�+zF�"y�'ǳ�s4��8�z�s�&��`),
		},
		{
			name:      "from invalid prefix",
			token:     "abc_1234",
			wantError: autogold.Expect("invalid token prefix"),
		},
		{
			name:      "from invalid encoding",
			token:     "slk_asdfasdfasdfasdf",
			wantError: autogold.Expect("invalid token encoding: encoding/hex: invalid byte: U+0073 's'"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			contents, err := ExtractLicenseKeyBasedAccessTokenContents(tc.token)
			if tc.wantError != nil {
				require.Error(t, err)
				tc.wantError.Equal(t, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.wantContents != nil {
				tc.wantContents.Equal(t, contents)
			} else {
				assert.Empty(t, contents)
			}
		})
	}
}
