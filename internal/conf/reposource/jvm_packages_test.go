package reposource

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestDecomposeMavenPath(t *testing.T) {
	obtained, _ := ParseMavenPackageFromRepoName("maven/org.hamcrest/hamcrest-core")
	assert.Equal(t, obtained.GroupID, "org.hamcrest")
	assert.Equal(t, obtained.ArtifactID, "hamcrest-core")
	assert.Equal(t, api.RepoName("maven/org.hamcrest/hamcrest-core"), obtained.RepoName())
}

func parseMavenDependencyOrPanic(t *testing.T, value string) *MavenVersionedPackage {
	dependency, err := ParseMavenVersionedPackage(value)
	if err != nil {
		t.Fatalf("error=%s", err)
	}
	return dependency
}

func TestGreaterThan(t *testing.T) {
	assert.True(t, versionGreaterThan("11.2.0", "1.2.0"))
	assert.True(t, versionGreaterThan("11.2.0", "2.2.0"))
	assert.True(t, versionGreaterThan("11.2.0", "11.2.0-M1"))
	assert.False(t, versionGreaterThan("11.2.0-M11", "11.2.0"))
}

func TestMavenDependency_Less(t *testing.T) {
	dependencies := []*MavenVersionedPackage{
		parseMavenDependencyOrPanic(t, "a:c:1.2.0"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0.Final"),
		parseMavenDependencyOrPanic(t, "a:a:1.2.0"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0"),
		parseMavenDependencyOrPanic(t, "a:b:1.11.0"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0-M11"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0-M1"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0-RC11"),
		parseMavenDependencyOrPanic(t, "a:b:1.2.0-RC1"),
		parseMavenDependencyOrPanic(t, "a:b:1.1.0"),
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Less(dependencies[j])
	})

	have := make([]string, 0, len(dependencies))
	for _, dep := range dependencies {
		have = append(have, dep.VersionedPackageSyntax())
	}

	want := []string{
		"a:c:1.2.0",
		"a:b:1.11.0",
		"a:b:1.2.0",
		"a:b:1.2.0.Final",
		"a:b:1.2.0-RC11",
		"a:b:1.2.0-RC1",
		"a:b:1.2.0-M11",
		"a:b:1.2.0-M1",
		"a:b:1.1.0",
		"a:a:1.2.0",
	}

	assert.Equal(t, want, have)
}
