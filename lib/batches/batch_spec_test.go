package batches

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestParseBatchSpec(t *testing.T) {
	t.Run("valid_without_version", func(t *testing.T) {
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
  fork: false
`

		_, err := ParseBatchSpec([]byte(spec))
		if err != nil {
			t.Fatalf("parsing valid spec returned error: %s", err)
		}
	})

	t.Run("valid with version 1", func(t *testing.T) {
		const spec = `
version: 1
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
  fork: false
`

		_, err := ParseBatchSpec([]byte(spec))
		if err != nil {
			t.Fatalf("parsing valid spec returned error: %s", err)
		}
	})

	t.Run("valid with version 2", func(t *testing.T) {
		const spec = `
version: 2
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
  fork: false
`

		_, err := ParseBatchSpec([]byte(spec))
		if err != nil {
			t.Fatalf("parsing valid spec returned error: %s", err)
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		const spec = `
version: 99
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
  fork: false
`
		_, err := ParseBatchSpec([]byte(spec))
		assert.Equal(t, "version: version must be one of the following: 1, 2", err.Error())
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

		_, err := ParseBatchSpec([]byte(spec))
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `batch spec includes steps but no changesetTemplate`
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

		_, err := ParseBatchSpec([]byte(spec))
		if err == nil {
			t.Fatal("no error returned")
		}

		// We expect this error to be user-friendly, which is why we test for
		// it specifically here.
		wantErr := `The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed.`
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
			batchSpec, err := ParseBatchSpec([]byte(spec))
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

		_, err := ParseBatchSpec([]byte(spec))
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `3 errors occurred:
	* on.0: Must validate one and only one schema (oneOf)
	* on.0: Must validate at least one schema (anyOf)
	* on.0: Must validate one and only one schema (oneOf)`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("mount path contains comma", func(t *testing.T) {
		const spec = `
name: test-spec
description: A test spec
steps:
  - run: /tmp/sample.sh
    container: alpine:3
    mount:
      - path: /foo,bar/
        mountpoint: /tmp
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`
		_, err := ParseBatchSpec([]byte(spec))
		assert.Equal(t, "step 1 mount path contains invalid characters", err.Error())
	})

	t.Run("mount mountpoint contains comma", func(t *testing.T) {
		const spec = `
name: test-spec
description: A test spec
steps:
  - run: /tmp/foo,bar/sample.sh
    container: alpine:3
    mount:
      - path: /valid/sample.sh
        mountpoint: /tmp/foo,bar/sample.sh
changesetTemplate:
  title: Test Mount
  body: Test a mounted path
  branch: test
  commit:
    message: Test
`
		_, err := ParseBatchSpec([]byte(spec))
		assert.Equal(t, "step 1 mount mountpoint contains invalid characters", err.Error())
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

func TestSkippedStepsForRepo(t *testing.T) {
	tests := map[string]struct {
		spec        *BatchSpec
		wantSkipped []int
	}{
		"no if": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
				},
			},
			wantSkipped: []int{},
		},

		"if has static true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSkipped: []int{},
		},

		"one of many steps has if with static true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int{},
		},

		"if has static non-true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSkipped: []int{0},
		},

		"one of many steps has if with static non-true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int{1},
		},

		"if expression that can be partially evaluated to true": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
				},
			},
			wantSkipped: []int{},
		},

		"if expression that can be partially evaluated to false": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSkipped: []int{0},
		},

		"one of many steps has if expression that can be evaluated to false": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: `${{ matches repository.name "horse" }}`},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int{1},
		},

		"if expression that can NOT be partially evaluated": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSkipped: []int{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			haveSkipped, err := SkippedStepsForRepo(tt.spec, "github.com/sourcegraph/src-cli", []string{})
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			want := tt.wantSkipped
			sort.Sort(sortableInt(want))
			have := make([]int, 0, len(haveSkipped))
			for s := range haveSkipped {
				have = append(have, s)
			}
			sort.Sort(sortableInt(have))
			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type sortableInt []int

func (s sortableInt) Len() int { return len(s) }

func (s sortableInt) Less(i, j int) bool { return s[i] < s[j] }

func (s sortableInt) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func TestBatchSpec_RequiredEnvVars(t *testing.T) {
	for name, tc := range map[string]struct {
		in   string
		want []string
	}{
		"no steps": {
			in:   `steps:`,
			want: []string{},
		},
		"no env vars": {
			in:   `steps: [run: asdf]`,
			want: []string{},
		},
		"static variable": {
			in:   `steps: [{run: asdf, env: [a: b]}]`,
			want: []string{},
		},
		"dynamic variable": {
			in:   `steps: [{run: asdf, env: [a]}]`,
			want: []string{"a"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var spec BatchSpec
			err := yaml.Unmarshal([]byte(tc.in), &spec)
			if err != nil {
				t.Fatal(err)
			}
			have := spec.RequiredEnvVars()

			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Errorf("unexpected value: have=%q want=%q", have, tc.want)
			}
		})
	}
}
