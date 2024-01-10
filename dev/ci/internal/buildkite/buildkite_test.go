package buildkite_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/yaml"

	"github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
)

func TestStepSoftFail(t *testing.T) {
	t.Run("Explicit exit codes", func(t *testing.T) {
		pipeline := buildkite.Pipeline{}
		stepOpt := buildkite.SoftFail(1, 2, 3, 4)
		pipeline.AddStep("foo", stepOpt)
		step, ok := pipeline.Steps[0].(*buildkite.Step)
		if !ok {
			t.Fatal("Pipeline step is not a buildkite.Step")
		}
		want := "1 2 3 4"
		got := step.Env["SOFT_FAIL_EXIT_CODES"]
		if got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})
	t.Run("Any exit code", func(t *testing.T) {
		pipeline := buildkite.Pipeline{}
		stepOpt := buildkite.SoftFail()
		pipeline.AddStep("foo", stepOpt)
		step, ok := pipeline.Steps[0].(*buildkite.Step)
		if !ok {
			t.Fatal("Pipeline step is not a buildkite.Step")
		}
		want := "*"
		got := step.Env["SOFT_FAIL_EXIT_CODES"]
		if got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})
}

func TestOutputSanitization(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
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
					MetaData: map[string]any{"foo": "bar"},
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
					MetaData: map[string]any{"foo": "bar"},
					Env:      map[string]string{"FOO": "rire"},
				},
				want: `{
	"message": "incredibly complex markdown with some $$dollar",
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
				t.Fatalf("want \n%s\ngot\n%s\n", test.want, string(b))
			}
		}
	})

	t.Run("YAML", func(t *testing.T) {
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
					MetaData: map[string]any{"foo": "bar"},
					Env:      map[string]string{"FOO": "rire"},
				},
				want: `branch: tree
commit: "123456"
env:
	FOO: rire
message: incredibly complex markdown with some ` + "`" + `backticks` + "`" + `
meta_data:
	foo: bar
`,
			},
			{
				// dollar sign gets escaped
				input: buildkite.BuildOptions{
					Message:  "incredibly complex markdown with some $dollar",
					Commit:   "123456",
					Branch:   "tree",
					MetaData: map[string]any{"foo": "bar"},
					Env:      map[string]string{"FOO": "rire"},
				},
				want: `branch: tree
commit: "123456"
env:
	FOO: rire
message: incredibly complex markdown with some $$dollar
meta_data:
	foo: bar
`,
			},
		}

		for _, test := range tests {
			b, err := yaml.Marshal(test.input)
			if err != nil {
				t.Fatal(err)
			}
			test.want = strings.ReplaceAll(test.want, "\t", "  ")
			if string(b) != test.want {
				t.Fatalf("want \n%s\ngot\n%s\n", test.want, string(b))
			}
		}
	})
}
