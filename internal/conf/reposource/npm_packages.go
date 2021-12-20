package reposource

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

const (
	// Exported for [NOTE: npm-tarball-filename-workaround].
	NPMScopeRegexString       = `(?P<scope>[0-9a-z_\\-]+)`
	npmPackageNameRegexString = `(?P<name>[0-9a-z_\\-]+)`
)

var (
	npmScopeRegex          = lazyregexp.New(`^` + NPMScopeRegexString + `$`)
	npmPackageNameRegex    = lazyregexp.New(`^` + npmPackageNameRegexString + `$`)
	scopedPackageNameRegex = lazyregexp.New(
		`^(@` + NPMScopeRegexString + `/)?` +
			npmPackageNameRegexString +
			`@(?P<version>[0-9a-zA-Z_\\-]+(\.[0-9a-zA-Z_\\-]+)*)$`)
	npmURLRegex = lazyregexp.New(
		`^npm/(` + NPMScopeRegexString + `/)?` +
			npmPackageNameRegexString + `$`)
)

// An NPM package of the form (@scope/)?name.
//
// The fields are kept private to reduce risk of not handling the empty scope
// case correctly.
type NPMPackage struct {
	// Optional scope () for a package, can potentially be "".
	// For more details, see https://docs.npmjs.com/cli/v8/using-npm/scope
	scope string
	// Required name for a package, always non-empty.
	name string
}

func NewNPMPackage(scope string, name string) (*NPMPackage, error) {
	if scope != "" && !npmScopeRegex.MatchString(scope) {
		return nil, errors.Errorf("illegal scope %s (allowed characters: 0-9, a-z, _, -)", scope)
	}
	if !npmPackageNameRegex.MatchString(name) {
		return nil, errors.Errorf("illegal package name %s (allowed characters: 0-9, a-z, _, -)", name)
	}
	return &NPMPackage{scope, name}, nil
}

// ParseNPMPackageFromRepoURL is a convenience function to parse a string in a
// 'npm/(scope/)?name' format into an NPMPackage.
func ParseNPMPackageFromRepoURL(urlPath string) (*NPMPackage, error) {
	match := npmURLRegex.FindStringSubmatch(urlPath)
	if match == nil {
		return nil, errors.Errorf("expected path in npm/(scope/)?name format but found %s", urlPath)
	}
	result := make(map[string]string)
	for i, groupName := range npmURLRegex.SubexpNames() {
		if i != 0 && groupName != "" {
			result[groupName] = match[i]
		}
	}
	scope, name := result["scope"], result["name"]
	return &NPMPackage{scope, name}, nil
}

// ParseNPMPackageFromPackageSyntax is a convenience function to parse a
// string in a '(@scope/)?name' format into an NPMPackage.
func ParseNPMPackageFromPackageSyntax(pkg string) (*NPMPackage, error) {
	dep, err := ParseNPMDependency(fmt.Sprintf("%s@0", pkg))
	if err != nil {
		return nil, err
	}
	return &dep.NPMPackage, nil
}

type NPMPackageSerializationHelper struct {
	Scope string
	Name  string
}

var _ json.Marshaler = &NPMPackage{}
var _ json.Unmarshaler = &NPMPackage{}

func (pkg *NPMPackage) MarshalJSON() ([]byte, error) {
	return json.Marshal(NPMPackageSerializationHelper{pkg.scope, pkg.name})
}

func (pkg *NPMPackage) UnmarshalJSON(data []byte) error {
	var wrapper NPMPackageSerializationHelper
	err := json.Unmarshal(data, &wrapper)
	if err != nil {
		return err
	}
	newPkg, err := NewNPMPackage(wrapper.Scope, wrapper.Name)
	if err != nil {
		return err
	}
	*pkg = *newPkg
	return nil
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (pkg *NPMPackage) RepoName() api.RepoName {
	if pkg.scope != "" {
		return api.RepoName(fmt.Sprintf("npm/%s/%s", pkg.scope, pkg.name))
	}
	return api.RepoName("npm/" + pkg.name)
}

// CloneURL returns a "URL" that can later be used to download a repo.
func (pkg *NPMPackage) CloneURL() string {
	return string(pkg.RepoName())
}

// MatchesDependencyString checks if a dependency (= package + version pair)
// refers to the same package as pkg.
func (pkg NPMPackage) MatchesDependencyString(depPackageSyntax string) bool {
	return strings.HasPrefix(depPackageSyntax, pkg.PackageSyntax()+"@")
}

func (pkg NPMPackage) PackageSyntax() string {
	if pkg.scope != "" {
		return fmt.Sprintf("@%s/%s", pkg.scope, pkg.name)
	}
	return pkg.name
}

// NPMDependency is a "versioned package" for use by npm commands, such as
// `npm install`.
//
// See also: [NOTE: Dependency-terminology]
//
// Reference:  https://docs.npmjs.com/cli/v8/commands/npm-install
type NPMDependency struct {
	NPMPackage

	// The version or tag (such as "latest") for a dependency.
	//
	// See https://docs.npmjs.com/cli/v8/using-npm/config#tag for more details
	// about tags.
	Version string
}

// ParseNPMDependency parses a string in a '(@scope/)?module@version' format into an NPMDependency.
//
// NPM supports many ways of specifying dependencies (https://docs.npmjs.com/cli/v8/commands/npm-install)
// but we only support exact versions for now.
func ParseNPMDependency(dependency string) (*NPMDependency, error) {
	// We use slightly more restrictive validation compared to the official
	// rules (https://github.com/npm/validate-npm-package-name#naming-rules).
	//
	// For example, NPM does not explicitly forbid package names with @ in them.
	// However, there don't seem to be any such packages in practice (I searched
	// 100k+ packages and got 0 hits). The web frontend relies on using '@' to
	// split between the package and rev-like part of the URL, such as
	// https://sourcegraph.com/github.com/golang/go@master, so avoiding '@' is
	// important.
	//
	// Scope names follow the same rules as package names.
	// (source: https://docs.npmjs.com/cli/v8/using-npm/scope)
	match := scopedPackageNameRegex.FindStringSubmatch(dependency)
	if match == nil {
		return nil, errors.Errorf("expected dependency in (@scope/)?name@version format but found %s", dependency)
	}
	result := make(map[string]string)
	for i, groupName := range scopedPackageNameRegex.SubexpNames() {
		if i != 0 && groupName != "" {
			result[groupName] = match[i]
		}
	}
	scope, name, version := result["scope"], result["name"], result["version"]
	return &NPMDependency{NPMPackage{scope, name}, version}, nil
}

// PackageManagerSyntax returns the dependency in NPM/Yarn syntax. The returned
// string can (for example) be passed to `npm install`.
func (d NPMDependency) PackageManagerSyntax() string {
	return fmt.Sprintf("%s@%s", d.NPMPackage.PackageSyntax(), d.Version)
}

func (d NPMDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

// SortDependencies sorts the dependencies by the semantic version in descending
// order. The latest version of a dependency becomes the first element of the
// slice.
func SortNPMDependencies(dependencies []NPMDependency) {
	sort.Slice(dependencies, func(i, j int) bool {
		iPkg, jPkg := dependencies[i].NPMPackage, dependencies[j].NPMPackage
		if iPkg == jPkg {
			return versionGreaterThan(dependencies[i].Version, dependencies[j].Version)
		}
		if iPkg.scope == jPkg.scope {
			return iPkg.name > jPkg.name
		}
		return iPkg.scope > jPkg.scope
	})
}
