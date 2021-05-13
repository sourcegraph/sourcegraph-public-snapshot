package reposource

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Masterminds/semver"

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

func ParseMavenDependency(dependency string) Dependency {
	parts := strings.Split(dependency, ":")
	version := parts[2]
	semanticVersion, _ := semver.NewVersion(version)
	return Dependency{
		Module: Module{
			GroupId:    parts[0],
			ArtifactId: parts[1],
		},
		Version:         version,
		SemanticVersion: semanticVersion,
	}
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
