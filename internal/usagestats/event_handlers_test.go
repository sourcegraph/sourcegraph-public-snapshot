package usagestats

import "testing"

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
			url:  "https://sourcegraph.com/search?q=abcd&utm_source=test&utm_campaign=test&utm_medium=test&utm_content=test&utm_term=test",
			want: "https://sourcegraph.com/redacted?q=redacted&utm_campaign=test&utm_content=test&utm_medium=test&utm_source=test&utm_term=test",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			have, err := redactSensitiveInfoFromCloudURL(c.url)
			if err != nil {
				t.Fatal("Error in redactSensitiveInfoFromCloudURL")
			}
			if c.want != have {
				t.Fatalf("Failed to redact info from Cloud URL, got %s, want %s", have, c.want)
			}
		})
	}
}
