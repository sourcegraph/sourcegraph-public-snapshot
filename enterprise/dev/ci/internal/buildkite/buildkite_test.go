package buildkite_test

import (
	"encoding/json"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
)

func TestOutputSanitization(t *testing.T) {
	tests := []struct {
		input buildkite.BuildOptions
		want  string
	}{
		{
			// backticks are left unchanged
			input: buildkite.BuildOptions{
				Message:  "incredibly complex markdown with some `backticks`",
				Commit:   "123456",
				Branch:   "tree",
				MetaData: map[string]interface{}{"foo": "bar"},
				Env:      map[string]string{"FOO": "rire"},
			},
			want: `{
	"message": "incredibly complex markdown with some ` + "`" + `backticks` + "`" + `",
	"commit": "123456",
	"branch": "tree",
	"meta_data": {
		"foo": "bar"
	},
	"env": {
		"FOO": "rire"
	}
}`,
		},
		{
			// dollar sign gets escaped
			input: buildkite.BuildOptions{
				Message:  "incredibly complex markdown with some $dollar",
				Commit:   "123456",
				Branch:   "tree",
				MetaData: map[string]interface{}{"foo": "bar"},
				Env:      map[string]string{"FOO": "rire"},
			},
			want: `{
	"message": "incredibly complex markdown with some \\$dollar",
	"commit": "123456",
	"branch": "tree",
	"meta_data": {
		"foo": "bar"
	},
	"env": {
		"FOO": "rire"
	}
}`,
		},
	}

	for _, test := range tests {
		b, err := json.MarshalIndent(test.input, "", "\t")
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != test.want {
			t.Fatalf("want %s got %s", test.want, string(b))
		}
	}
}
