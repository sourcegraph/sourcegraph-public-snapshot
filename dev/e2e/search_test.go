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
		found := false
		for _, r := range results {
			if r.Name == "github.com/sourcegraph/e2e-test-private-repository" {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("private repository not found")
		}
	})
}
