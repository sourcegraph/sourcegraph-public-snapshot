package batches

import "testing"

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

		_, err := ParseBatchSpec([]byte(spec), featureFlags{})
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

		_, err := ParseBatchSpec([]byte(spec), featureFlags{})
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
}
