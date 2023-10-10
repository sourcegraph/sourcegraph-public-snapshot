package usagestats

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSensitiveInfoFromCloudURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "path redacted",
			url:  "https://sourcegraph.com/github.com/test/test",
			want: "https://sourcegraph.com/github.com/test/test",
		},
		{
			name: "urls are not redacted from dotcom",
			url:  "https://sourcegraph.com/search?q=abcd",
			want: "https://sourcegraph.com/search/q=abcd",
		},
		{
			name: "path and non-approved query param redacted, approved params retained from cloud instances",
			url:  "https://sourcegraph.com/search?q=abcd&utm_source=test&utm_campaign=test&utm_medium=test&utm_content=test&utm_term=test&utm_cid=test",
			want: "https://sourcegraph.com/search/redacted?q=redacted&utm_campaign=test&utm_cid=test&utm_content=test&utm_medium=test&utm_source=test&utm_term=test",
		},
		{
			name: "path and query param not redacted from dotcom",
			url:  "https://sourcegraph.sourcegraph.com/search?q=abcd&utm_source=test&utm_campaign=test&utm_medium=test&utm_content=test&utm_term=test&utm_cid=test",
			want: "https://sourcegraph.sourcegraph.com/search?q=abcd&utm_source=test&utm_campaign=test&utm_medium=test&utm_content=test&utm_term=test&utm_cid=test",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			have, err := redactSensitiveInfoFromCloudURL(c.url)
			require.NoError(t, err)
			assert.Equal(t, c.want, have)
		})
	}
}
