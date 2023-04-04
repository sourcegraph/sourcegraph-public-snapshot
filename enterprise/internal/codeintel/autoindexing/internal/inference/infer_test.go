package inference

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
)

func TestInferRepositoryAndRevision(t *testing.T) {
	t.Run("Go", func(t *testing.T) {
		testCases := []struct {
			pkg      dependencies.MinimialVersionedPackageRepo
			repoName string
			revision string
		}{
			{
				pkg: dependencies.MinimialVersionedPackageRepo{
					Scheme:  "gomod",
					Name:    "https://github.com/sourcegraph/sourcegraph",
					Version: "v2.3.2",
				},
				repoName: "github.com/sourcegraph/sourcegraph",
				revision: "v2.3.2",
			},
			{
				pkg: dependencies.MinimialVersionedPackageRepo{
					Scheme:  "gomod",
					Name:    "https://github.com/aws/aws-sdk-go-v2/credentials",
					Version: "v0.1.0",
				},
				repoName: "github.com/aws/aws-sdk-go-v2",
				revision: "v0.1.0",
			},
			{
				pkg: dependencies.MinimialVersionedPackageRepo{
					Scheme:  "gomod",
					Name:    "https://github.com/sourcegraph/sourcegraph",
					Version: "v0.0.0-de0123456789",
				},
				repoName: "github.com/sourcegraph/sourcegraph",
				revision: "de0123456789",
			},
			{
				pkg: dependencies.MinimialVersionedPackageRepo{
					Scheme:  "npm",
					Name:    "mypackage",
					Version: "1.0.0",
				},
				repoName: "npm/mypackage",
				revision: "v1.0.0",
			},
			{
				pkg: dependencies.MinimialVersionedPackageRepo{
					Scheme:  "npm",
					Name:    "@myscope/mypackage",
					Version: "1.0.0",
				},
				repoName: "npm/myscope/mypackage",
				revision: "v1.0.0",
			},
		}

		for _, testCase := range testCases {
			repoName, revision, ok := InferRepositoryAndRevision(testCase.pkg)
			if !ok {
				t.Fatalf("expected repository to be inferred")
			}

			if string(repoName) != testCase.repoName {
				t.Errorf("unexpected repo name. want=%q have=%q", testCase.repoName, string(repoName))
			}
			if revision != testCase.revision {
				t.Errorf("unexpected revision. want=%q have=%q", testCase.revision, revision)
			}
		}
	})
}
