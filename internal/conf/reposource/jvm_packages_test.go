package reposource

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestDecomposeMavenPath(t *testing.T) {
	obtained, _ := ParseMavenModule("//maven/junit/junit")
	assert.Equal(t, api.RepoName("maven/junit/junit"), obtained.RepoName())
}

func ParseMavenDependencyOrPanic(value string) Dependency {
	dependency, err := ParseMavenDependency(value)
	if err != nil {
		panic(dependency)
	}
	return dependency
}

func TestSortDependencies(t *testing.T) {
	dependencies := []Dependency{
		ParseMavenDependencyOrPanic("a:c:1.2.0"),
		ParseMavenDependencyOrPanic("a:a:1.2.0"),
		ParseMavenDependencyOrPanic("a:b:1.2.0"),
		ParseMavenDependencyOrPanic("a:b:1.11.0"),
		ParseMavenDependencyOrPanic("a:b:1.2.0-M11"),
		ParseMavenDependencyOrPanic("a:b:1.2.0-M1"),
		ParseMavenDependencyOrPanic("a:b:1.1.0"),
	}
	expected := []Dependency{
		ParseMavenDependencyOrPanic("a:c:1.2.0"),
		ParseMavenDependencyOrPanic("a:b:1.11.0"),
		ParseMavenDependencyOrPanic("a:b:1.2.0"),
		ParseMavenDependencyOrPanic("a:b:1.2.0-M11"),
		ParseMavenDependencyOrPanic("a:b:1.2.0-M1"),
		ParseMavenDependencyOrPanic("a:b:1.1.0"),
		ParseMavenDependencyOrPanic("a:a:1.2.0"),
	}
	SortDependencies(dependencies)
	assert.Equal(t, expected, dependencies)
}
