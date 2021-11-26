package batches

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestParseBatchSpec(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err != nil {
			t.Fatalf("parsing valid spec returned error: %s", err)
		}
	})

	t.Run("missing changesetTemplate", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `1 error occurred:
	* batch spec includes steps but no changesetTemplate

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("invalid batch change name", func(t *testing.T) {
		const spec = `
name: this name is invalid cause it contains whitespace
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		// We expect this error to be user-friendly, which is why we test for
		// it specifically here.
		wantErr := `1 error occurred:
	* The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed.

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("uses unsupported conditional exec", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    if: "false"
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `1 error occurred:
	* step 1 in batch spec uses the 'if' attribute for conditional execution, which is not supported in this Sourcegraph version

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("parsing if attribute", func(t *testing.T) {
		const specTemplate = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    if: %s
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		for _, tt := range []struct {
			raw  string
			want string
		}{
			{raw: `"true"`, want: "true"},
			{raw: `"false"`, want: "false"},
			{raw: `true`, want: "true"},
			{raw: `false`, want: "false"},
			{raw: `"${{ foobar }}"`, want: "${{ foobar }}"},
			{raw: `${{ foobar }}`, want: "${{ foobar }}"},
			{raw: `foobar`, want: "foobar"},
		} {
			spec := fmt.Sprintf(specTemplate, tt.raw)
			batchSpec, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{AllowConditionalExec: true})
			if err != nil {
				t.Fatal(err)
			}

			if batchSpec.Steps[0].IfCondition() != tt.want {
				t.Fatalf("wrong IfCondition. want=%q, got=%q", tt.want, batchSpec.Steps[0].IfCondition())
			}
		}
	})
	t.Run("uses conflicting branch attributes", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repository: github.com/foo/bar
    branch: foo
    branches: bar
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `3 errors occurred:
	* on.0: Must validate one and only one schema (oneOf)
	* on.0: Must validate at least one schema (anyOf)
	* on.0: Must validate one and only one schema (oneOf)

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})
	t.Run("uses unsupported files attribute", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
    files:
      /tmp/horse.txt: yipeeee

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{AllowFiles: false})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `1 error occurred:
	* step 1 in batch spec uses the 'files' attribute to create files in the step container, which is not supported in this Batch Changes version

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})
}

func TestOnQueryOrRepository(t *testing.T) {
	// Note that we're not testing the full set of branch value possibilities
	// here: there are tests in branch_test.go to handle that. This is just
	// ensuring that the unmarshalling does sensible things.
	for name, tc := range map[string]struct {
		input string
		want  OnQueryOrRepository
	}{
		"branch": {
			input: `{"repository": "github.com/a/b", "branch": "foo"}`,
			want: OnQueryOrRepository{
				Repository: "github.com/a/b",
				Branches:   []string{"foo"},
			},
		},
		"branches": {
			input: `{"repository": "github.com/a/b", "branches": "foo"}`,
			want: OnQueryOrRepository{
				Repository: "github.com/a/b",
				Branches:   []string{"foo"},
			},
		},
		"no branches": {
			input: `{"repositoriesMatchingQuery": "foo"}`,
			want: OnQueryOrRepository{
				RepositoriesMatchingQuery: "foo",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Run("json", func(t *testing.T) {
				have := OnQueryOrRepository{}
				err := json.Unmarshal([]byte(tc.input), &have)
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})

			t.Run("yaml", func(t *testing.T) {
				have := OnQueryOrRepository{}
				err := yaml.Unmarshal([]byte(tc.input), &have)
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})
		})
	}
}
