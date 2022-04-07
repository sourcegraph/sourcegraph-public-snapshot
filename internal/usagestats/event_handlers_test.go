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
			want: "https://sourcegraph.com/redacted",
		},
		{
			name: "path and non-approved query param redacted",
			url:  "https://sourcegraph.com/search?q=abcd",
			want: "https://sourcegraph.com/redacted?q=redacted",
		},
		{
			name: "path and non-approved query param redacted, approved params retained",
			url:  "https://sourcegraph.com/search?q=abcd&utm_source=test&utm_campaign=test&utm_medium=test&utm_content=test&utm_term=test&utm_cid=test",
			want: "https://sourcegraph.com/redacted?q=redacted&utm_campaign=test&utm_cid=test&utm_content=test&utm_medium=test&utm_source=test&utm_term=test",
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
