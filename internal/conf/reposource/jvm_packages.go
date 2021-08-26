package reposource

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/sourcegraph/sourcegraph/internal/api"
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

func (m *MavenModule) SortText() string {
	return fmt.Sprintf("%s:%s", m.GroupID, m.ArtifactID)
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

func (d MavenDependency) CoursierSyntax() string {
	return fmt.Sprintf("%s:%s:%s", d.MavenModule.GroupID, d.MavenModule.ArtifactID, d.Version)
}

func (d MavenDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d MavenDependency) LsifJavaDependencies() []string {
	if d.IsJDK() {
		return []string{}
	}
	return []string{d.CoursierSyntax()}
}

// ParseMavenDependency parses a dependency string in the Coursier format (colon seperated group ID, artifact ID and version)
// into a MavenDependency.
func ParseMavenDependency(dependency string) (MavenDependency, error) {
	parts := strings.Split(dependency, ":")
	if len(parts) < 3 {
		return MavenDependency{}, fmt.Errorf("dependency %q must contain at least two colon ':' characters", dependency)
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
		return MavenModule{}, fmt.Errorf("failed to parse a maven module from the path %s", urlPath)
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

// versionGreaterThan return true if the version string a is "greater" than the
// version string b using psuedo semantic versioning. Java package versions can
// be arbitrary strings so this method must always succeed even if a version is
// not using the semantic version format. The comparison is lexicographical by
// default except for the digit parts which are compared numerically. For
// example, versionGreaterThan return true when comparing 11.2.0 and 2.2.0
// because the number 11 is larger than 2 (even if "2" is lexicographically
// larger than "11").
func versionGreaterThan(version1, version2 string) bool {
	index := 0
	end := len(version1)
	if len(version2) < end {
		end = len(version2)
	}
	for index < end {
		rune1 := rune(version1[index])
		rune2 := rune(version2[index])
		if unicode.IsDigit(rune1) && unicode.IsDigit(rune2) {
			int1 := versionParseInt(index, version1)
			int2 := versionParseInt(index, version2)
			if int1 == int2 {
				index = versionNextNonDigitOffset(index, version1)
			} else {
				return int1 > int2
			}
		} else {
			if rune1 == rune2 {
				index += 1
			} else {
				return rune1 > rune2
			}
		}
	}
	return len(version1) < len(version2)
}

// versionParseInt returns the integer value of the number that appears at given
// index of the given string.
func versionParseInt(index int, a string) int {
	end := versionNextNonDigitOffset(index, a)
	value, _ := strconv.Atoi(a[index:end])
	return value
}

// versionNextNonDigitOffset returns the offset of the next non-digit character
// of the given string starting at the given index.
func versionNextNonDigitOffset(index int, b string) int {
	offset := index
	for offset < len(b) && unicode.IsDigit(rune(b[offset])) {
		offset += 1
	}
	return offset
}
