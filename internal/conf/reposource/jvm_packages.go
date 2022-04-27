package reposource

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MavenModule struct {
	GroupID    string
	ArtifactID string
}

func (m *MavenModule) Equal(other *MavenModule) bool {
	return m == other || (m != nil && other != nil && *m == *other)
}

func (m *MavenModule) IsJDK() bool {
	return m.Equal(jdkModule())
}

func (m *MavenModule) MatchesDependencyString(dependency string) bool {
	return strings.HasPrefix(dependency, fmt.Sprintf("%s:%s:", m.GroupID, m.ArtifactID))
}

func (m *MavenModule) CoursierSyntax() string {
	return fmt.Sprintf("%s:%s", m.GroupID, m.ArtifactID)
}

func (m *MavenModule) PackageSyntax() string {
	return m.CoursierSyntax()
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
	*MavenModule
	Version string
}

func (m *MavenDependency) Equal(o *MavenDependency) bool {
	return m == o || (m != nil && o != nil &&
		m.MavenModule.Equal(o.MavenModule) &&
		m.Version == o.Version)
}

func (m *MavenDependency) Less(other PackageDependency) bool {
	o := other.(*MavenDependency)

	if m.MavenModule.Equal(o.MavenModule) {
		return versionGreaterThan(m.Version, o.Version)
	}

	// TODO: This SortText method is quite inefficient and allocates.
	return m.SortText() > o.SortText()
}

func (d *MavenDependency) PackageManagerSyntax() string {
	return fmt.Sprintf("%s:%s", d.PackageSyntax(), d.Version)
}

func (d *MavenDependency) PackageVersion() string {
	return d.Version
}

func (d *MavenDependency) Scheme() string {
	return "semanticdb"
}

func (d *MavenDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d *MavenDependency) LsifJavaDependencies() []string {
	if d.IsJDK() {
		return []string{}
	}
	return []string{d.PackageManagerSyntax()}
}

// ParseMavenDependency parses a dependency string in the Coursier format (colon seperated group ID, artifact ID and version)
// into a MavenDependency.
func ParseMavenDependency(dependency string) (*MavenDependency, error) {
	parts := strings.Split(dependency, ":")
	if len(parts) < 3 {
		return nil, errors.Newf("dependency %q must contain at least two colon ':' characters", dependency)
	}
	version := parts[2]

	return &MavenDependency{
		MavenModule: &MavenModule{
			GroupID:    parts[0],
			ArtifactID: parts[1],
		},
		Version: version,
	}, nil
}

// ParseMavenModule returns a parsed JVM module from the provided URL path, without a leading `/`
func ParseMavenModule(urlPath string) (*MavenModule, error) {
	if urlPath == "jdk" {
		return jdkModule(), nil
	}
	parts := strings.SplitN(strings.TrimPrefix(urlPath, "maven/"), "/", 2)
	if len(parts) != 2 {
		return nil, errors.Newf("failed to parse a maven module from the path %s", urlPath)
	}

	return &MavenModule{
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
func jdkModule() *MavenModule {
	return &MavenModule{
		GroupID:    "jdk",
		ArtifactID: "jdk",
	}
}
