package v1

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeInstanceDomain(t *testing.T) {
	for _, tc := range []struct {
		name       string
		domain     string
		wantDomain autogold.Value
		wantError  autogold.Value
	}{{
		name:       "normal URL",
		domain:     "https://souregraph.com/",
		wantDomain: autogold.Expect("souregraph.com"),
	}, {
		name:       "already a host",
		domain:     "sourcegraph.com",
		wantDomain: autogold.Expect("sourcegraph.com"),
	}, {
		name:       "subdomain",
		domain:     "foo.sourcegraph.com",
		wantDomain: autogold.Expect("foo.sourcegraph.com"),
	}, {
		name:       "host with trailing slash",
		domain:     "sourcegraph.com/",
		wantDomain: autogold.Expect("sourcegraph.com"),
	}, {
		name:       "normal URL with path",
		domain:     "https://souregraph.com/search",
		wantDomain: autogold.Expect("souregraph.com"),
	}, {
		name:      "clearly not a domain",
		domain:    "foo-bar",
		wantError: autogold.Expect("domain does contain a '.'"),
	}, {
		name:      "empty value",
		domain:    "",
		wantError: autogold.Expect("domain is empty"),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			gotDomain, err := NormalizeInstanceDomain(tc.domain)
			if tc.wantError != nil {
				require.Error(t, err)
				tc.wantError.Equal(t, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.wantDomain != nil {
				tc.wantDomain.Equal(t, gotDomain)
			} else {
				assert.Empty(t, gotDomain)
			}
		})
	}
}
