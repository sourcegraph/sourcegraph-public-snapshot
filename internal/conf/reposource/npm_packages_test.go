package reposource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNPMDependency(t *testing.T) {
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
		dep, err := ParseNPMDependency(entry.testName)
		if entry.expect && (err != nil) {
			t.Errorf("expected success but got error '%s' when parsing %s",
				err.Error(), entry.testName)
		} else if !entry.expect && err == nil {
			t.Errorf("expected error but successfully parsed %s into %+v", entry.testName, dep)
		}
	}
}

func TestSortNPMDependencies(t *testing.T) {
	dependencies := []*NPMDependency{
		parseNPMDependencyOrPanic(t, "ac@1.2.0"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0.Final"),
		parseNPMDependencyOrPanic(t, "aa@1.2.0"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0"),
		parseNPMDependencyOrPanic(t, "ab@1.11.0"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-M11"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-M1"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-RC11"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-RC1"),
		parseNPMDependencyOrPanic(t, "ab@1.1.0"),
	}
	expected := []*NPMDependency{
		parseNPMDependencyOrPanic(t, "ac@1.2.0"),
		parseNPMDependencyOrPanic(t, "ab@1.11.0"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0.Final"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-RC11"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-RC1"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-M11"),
		parseNPMDependencyOrPanic(t, "ab@1.2.0-M1"),
		parseNPMDependencyOrPanic(t, "ab@1.1.0"),
		parseNPMDependencyOrPanic(t, "aa@1.2.0"),
	}
	SortNPMDependencies(dependencies)
	assert.Equal(t, expected, dependencies)
}

func parseNPMDependencyOrPanic(t *testing.T, value string) *NPMDependency {
	dependency, err := ParseNPMDependency(value)
	if err != nil {
		t.Fatalf("error=%s", err)
	}
	return dependency
}
