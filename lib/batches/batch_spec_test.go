package batches

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
    branches: [bar]
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
}

func TestOnQueryOrRepository_Branches(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			input *OnQueryOrRepository
			want  []string
		}{
			"no branches": {
				input: &OnQueryOrRepository{},
				want:  nil,
			},
			"single branch": {
				input: &OnQueryOrRepository{Branch: "foo"},
				want:  []string{"foo"},
			},
			"single branch, non-nil but empty branches": {
				input: &OnQueryOrRepository{
					Branch:   "foo",
					Branches: []string{},
				},
				want: []string{"foo"},
			},
			"multiple branches": {
				input: &OnQueryOrRepository{
					Branches: []string{"foo", "bar"},
				},
				want: []string{"foo", "bar"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				have, err := tc.input.GetBranches()
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		_, err := (&OnQueryOrRepository{
			Branch:   "foo",
			Branches: []string{"bar"},
		}).GetBranches()
		assert.Equal(t, ErrConflictingBranches, err)
	})
}
