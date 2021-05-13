package reposource

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("(-want +got):\n%s", diff)
	}
}

func TestDecomposeMavenPath(t *testing.T) {
	obtained, _ := ParseMavenModule("//maven/junit/junit")
	assertEqual(t, obtained.RepoName(), "maven/junit/junit")
}

func TestSortDependencies(t *testing.T) {
	dependencies := []Dependency{
		ParseMavenDependency("a:c:1.2.0"),
		ParseMavenDependency("a:a:1.2.0"),
		ParseMavenDependency("a:b:1.2.0"),
		ParseMavenDependency("a:b:1.11.0"),
		ParseMavenDependency("a:b:1.2.0-M11"),
		ParseMavenDependency("a:b:1.2.0-M1"),
		ParseMavenDependency("a:b:1.1.0"),
	}
	expected := []Dependency{
		ParseMavenDependency("a:c:1.2.0"),
		ParseMavenDependency("a:b:1.11.0"),
		ParseMavenDependency("a:b:1.2.0"),
		ParseMavenDependency("a:b:1.2.0-M11"),
		ParseMavenDependency("a:b:1.2.0-M1"),
		ParseMavenDependency("a:b:1.1.0"),
		ParseMavenDependency("a:a:1.2.0"),
	}
	SortDependencies(dependencies)
	assertEqual(t, dependencies, expected)
}
