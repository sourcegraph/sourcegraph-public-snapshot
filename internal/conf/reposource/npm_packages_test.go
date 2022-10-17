package reposource

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNpmPackageVersion(t *testing.T) {
	table := []struct {
		testName string
		expect   bool
	}{
		{"@scope/package@1.2.3-abc", true},
		{"package@latest", true},
		{"@scope/package@latest", true},
		{"package@1.2.3", true},
		{"package.js@1.2.3", true},
		{"package-1.2.3", false},
		{"@scope/package", false},
		{"@weird.scope/package@1.2.3", true},
		{"@scope/package.js@1.2.3", true},
		{"package@1$%", false},
		{"@scope-package@1.2.3", false},
		{"@/package@1.2.3", false},
		{"@scope/@1.2.3", false},
		{"@dashed-scope/abc@0", true},
		{"@a.b-c.d-e/f.g--h.ijk-l@0.1-abc", true},
		{"@A.B-C.D-E/F.G--H.IJK-L@0.1-ABC", true},
	}
	for _, entry := range table {
		dep, err := ParseNpmVersionedPackage(entry.testName)
		if entry.expect && (err != nil) {
			t.Errorf("expected success but got error '%s' when parsing %s",
				err.Error(), entry.testName)
		} else if !entry.expect && err == nil {
			t.Errorf("expected error but successfully parsed %s into %+v", entry.testName, dep)
		}
	}
}

func TestNpmDependency_Less(t *testing.T) {
	dependencies := []*NpmVersionedPackage{
		parseNpmDependencyOrPanic(t, "ac@1.2.0"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0.Final"),
		parseNpmDependencyOrPanic(t, "aa@1.2.0"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0"),
		parseNpmDependencyOrPanic(t, "ab@1.11.0"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-M11"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-M1"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-RC11"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-RC1"),
		parseNpmDependencyOrPanic(t, "ab@1.1.0"),
	}
	expected := []*NpmVersionedPackage{
		parseNpmDependencyOrPanic(t, "ac@1.2.0"),
		parseNpmDependencyOrPanic(t, "ab@1.11.0"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0.Final"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-RC11"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-RC1"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-M11"),
		parseNpmDependencyOrPanic(t, "ab@1.2.0-M1"),
		parseNpmDependencyOrPanic(t, "ab@1.1.0"),
		parseNpmDependencyOrPanic(t, "aa@1.2.0"),
	}

	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Less(dependencies[j])
	})

	assert.Equal(t, expected, dependencies)
}

func parseNpmDependencyOrPanic(t *testing.T, value string) *NpmVersionedPackage {
	dependency, err := ParseNpmVersionedPackage(value)
	if err != nil {
		t.Fatalf("error=%s", err)
	}
	return dependency
}
