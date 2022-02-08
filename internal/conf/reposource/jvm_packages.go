package reposource

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MavenModule struct {
	GroupID    string
	ArtifactID string
}

func (m *MavenModule) IsJDK() bool {
	return *m == jdkModule()
}

func (m *MavenModule) MatchesDependencyString(dependency string) bool {
	return strings.HasPrefix(dependency, fmt.Sprintf("%s:%s:", m.GroupID, m.ArtifactID))
}

func (m *MavenModule) CoursierSyntax() string {
	return fmt.Sprintf("%s:%s", m.GroupID, m.ArtifactID)
}

func (m *MavenModule) SortText() string {
	return m.CoursierSyntax()
}

func (m *MavenModule) LsifJavaKind() string {
	if m.IsJDK() {
		return "jdk"
	}
	return "maven"
}

func (m *MavenModule) RepoName() api.RepoName {
	if m.IsJDK() {
		return "jdk"
	}
	return api.RepoName(fmt.Sprintf("maven/%s/%s", m.GroupID, m.ArtifactID))
}

func (m *MavenModule) CloneURL() string {
	cloneURL := url.URL{Path: string(m.RepoName())}
	return cloneURL.String()
}

// See [NOTE: Dependency-terminology]
type MavenDependency struct {
	MavenModule
	Version string
}

// SortDependencies sorts the dependencies by the semantic version in descending
// order. The latest version of a dependency becomes the first element of the
// slice
func SortDependencies(dependencies []MavenDependency) {
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].MavenModule == dependencies[j].MavenModule {
			return versionGreaterThan(dependencies[i].Version, dependencies[j].Version)
		}
		return dependencies[i].MavenModule.SortText() > dependencies[j].MavenModule.SortText()
	})
}

func (d MavenDependency) IsJDK() bool {
	return d.MavenModule.IsJDK()
}

func (d MavenDependency) PackageManagerSyntax() string {
	return fmt.Sprintf("%s:%s:%s", d.MavenModule.GroupID, d.MavenModule.ArtifactID, d.Version)
}

func (d MavenDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d MavenDependency) LsifJavaDependencies() []string {
	if d.IsJDK() {
		return []string{}
	}
	return []string{d.PackageManagerSyntax()}
}

// ParseMavenDependency parses a dependency string in the Coursier format (colon seperated group ID, artifact ID and version)
// into a MavenDependency.
func ParseMavenDependency(dependency string) (MavenDependency, error) {
	parts := strings.Split(dependency, ":")
	if len(parts) < 3 {
		return MavenDependency{}, errors.Newf("dependency %q must contain at least two colon ':' characters", dependency)
	}
	version := parts[2]

	return MavenDependency{
		MavenModule: MavenModule{
			GroupID:    parts[0],
			ArtifactID: parts[1],
		},
		Version: version,
	}, nil
}

// ParseMavenModule returns a parsed JVM module from the provided URL path, without a leading `/`
func ParseMavenModule(urlPath string) (MavenModule, error) {
	if urlPath == "jdk" {
		return jdkModule(), nil
	}
	parts := strings.SplitN(strings.TrimPrefix(urlPath, "maven/"), "/", 2)
	if len(parts) != 2 {
		return MavenModule{}, errors.Newf("failed to parse a maven module from the path %s", urlPath)
	}

	return MavenModule{
		GroupID:    parts[0],
		ArtifactID: parts[1],
	}, nil
}

// jdkModule returns the module for the Java standard library (JDK). This module
// is technically not a "maven module" because the JDK is not published as a
// Maven library. The only difference that's relevant for Sourcegraph is that we
// use a different coursier command to download JDK sources compared to normal
// maven modules:
// - JDK sources: `coursier java-home --jvm VERSION`
// - Maven sources: `coursier fetch MAVEN_MODULE:VERSION --classifier=sources`
// Since the difference is so small, the code is easier to read/maintain if we
// model the JDK as a Maven module.
func jdkModule() MavenModule {
	return MavenModule{
		GroupID:    "jdk",
		ArtifactID: "jdk",
	}
}
