// +build e2e

package main

import (
	"testing"
)

func TestSearch_VisibilityFilter(t *testing.T) {
	t.Run("type:repo visibility:private", func(t *testing.T) {
		results, err := client.SearchRepositories("type:repo visibility:private")
		if err != nil {
			t.Fatal(err)
		}
		missing := results.Exists("github.com/sourcegraph/e2e-test-private-repository")
		if len(missing) > 0 {
			t.Fatalf("private repository not found: %v", missing)
		}
	})
}
