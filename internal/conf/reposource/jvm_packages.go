package reposource

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Module struct {
	GroupId    string
	ArtifactId string
}

func (m Module) MatchesDependencyString(dependency string) bool {
	return strings.HasPrefix(dependency, fmt.Sprintf("%s:%s:", m.GroupId, m.ArtifactId))
}

func (m Module) SortText() string {
	return fmt.Sprintf("%s:%s", m.GroupId, m.ArtifactId)
}

func (m Module) RepoName() api.RepoName {
	return api.RepoName(fmt.Sprintf("maven/%s/%s", m.GroupId, m.ArtifactId))
}

func (m Module) CloneURL() string {
	cloneURL := url.URL{Path: string(m.RepoName())}
	return cloneURL.String()
}

type Dependency struct {
	Module
	Version         string
	SemanticVersion *semver.Version
}

// SortDependencies sorts the dependencies by the semantic version in descending
// order. The latest version of a dependency becomes the first element of the
// slice
func SortDependencies(dependencies []Dependency) {
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].Module == dependencies[j].Module {
			return dependencies[i].SemanticVersion.GreaterThan(dependencies[j].SemanticVersion)
		}
		return dependencies[i].Module.SortText() > dependencies[j].Module.SortText()
	})
}

func (d Dependency) CoursierSyntax() string {
	return fmt.Sprintf("%s:%s:%s", d.Module.GroupId, d.Module.ArtifactId, d.Version)
}

func (d Dependency) GitTagFromVersion() string {
	return "v" + d.Version
}

func ParseMavenDependency(dependency string) (Dependency, error) {
	parts := strings.Split(dependency, ":")
	if len(parts) < 3 {
		return Dependency{}, errors.Errorf("dependency %s must contain at least two colon ':' characters", dependency)

	}
	version := parts[2]

	// Ignore error from semantic version parsing because we only use the
	// semantic version for sorting dependencies, which falls back to
	// lexicographical ordering if the semantic version is missing. We can't
	// guarantee that every published Java package has a valid semantic
	// version according to the implementation of the Go-lang semver
	// package.
	semanticVersion, _ := semver.NewVersion(version)

	return Dependency{
		Module: Module{
			GroupId:    parts[0],
			ArtifactId: parts[1],
		},
		Version:         version,
		SemanticVersion: semanticVersion,
	}, nil
}

func ParseMavenModule(path string) (Module, error) {
	parts := strings.SplitN(strings.TrimPrefix(path, "maven/"), "/", 2)
	if len(parts) != 2 {
		return Module{}, fmt.Errorf("failed to parse a maven module from the path %s", path)
	}

	return Module{
		GroupId:    parts[0],
		ArtifactId: parts[1],
	}, nil
}
