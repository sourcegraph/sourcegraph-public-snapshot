package lspext

import (
	"fmt"
	"sort"
	"strings"
)

// WorkspacePackagesParams is parameters for the `workspace/xpackages` extension.
//
// See: https://github.com/sourcegraph/language-server-protocol/blob/7fd3c1/extension-workspace-references.md
type WorkspacePackagesParams struct{}

// PackageInformation is the metadata associated with a build-system-
// or package-manager-level package. Sometimes, languages have
// abstractions called "packages" as well, but this refers
// specifically to packages as defined by the build system or package
// manager. E.g., Python pip packages (NOT Python language packages or
// modules), Go packages, Maven packages (NOT Java language packages),
// npm modules (NOT JavaScript language modules).  PackageInformation
// includes both attributes of the package itself and attributes of
// the package's dependencies.
type PackageInformation struct {
	// Package is the set of attributes of the package
	Package PackageDescriptor `json:"package,omitempty"`

	// Dependencies is the list of dependency attributes
	Dependencies []DependencyReference `json:"dependencies,omitempty"`
}

// PackageDescriptor identifies a package (usually but not always uniquely).
type PackageDescriptor map[string]interface{}

// String returns a consistently ordered string representation of the
// PackageDescriptor. It is useful for testing.
func (s PackageDescriptor) String() string {
	sm := make(sortedMap, 0, len(s))
	for k, v := range s {
		sm = append(sm, mapValue{key: k, value: v})
	}
	sort.Sort(sm)
	var str string
	for _, v := range sm {
		str += fmt.Sprintf("%s:%v ", v.key, v.value)
	}
	return strings.TrimSpace(str)
}

type mapValue struct {
	key   string
	value interface{}
}

type sortedMap []mapValue

func (s sortedMap) Len() int           { return len(s) }
func (s sortedMap) Less(i, j int) bool { return s[i].key < s[j].key }
func (s sortedMap) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
