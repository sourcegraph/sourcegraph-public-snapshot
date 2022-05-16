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

func (m *MavenModule) Description() string { return "" }

type MavenMetadata struct {
	Module *MavenModule
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

// ParseMavenDependency parses a dependency string in the Coursier format
// (colon seperated group ID, artifact ID and an optional version) into a MavenDependency.
func ParseMavenDependency(dependency string) (*MavenDependency, error) {
	dep := &MavenDependency{MavenModule: &MavenModule{}}

	switch ps := strings.Split(dependency, ":"); len(ps) {
	case 3:
		dep.Version = ps[2]
		fallthrough
	case 2:
		dep.MavenModule.GroupID = ps[0]
		dep.MavenModule.ArtifactID = ps[1]
	default:
		return nil, errors.Newf("dependency %q must contain at least one colon ':' character", dependency)
	}

	return dep, nil
}

// ParseMavenDependencyFromRepoName is a convenience function to parse a repo name in a
// 'maven/<name>' format into a MavenDependency.
func ParseMavenDependencyFromRepoName(name string) (*MavenDependency, error) {
	if name == "jdk" {
		return &MavenDependency{MavenModule: jdkModule()}, nil
	}

	dep := strings.ReplaceAll(strings.TrimPrefix(name, "maven/"), "/", ":")
	if len(dep) == len(name) {
		return nil, errors.New("invalid maven dependency repo name, missing maven/ prefix")
	}

	return ParseMavenDependency(dep)
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
