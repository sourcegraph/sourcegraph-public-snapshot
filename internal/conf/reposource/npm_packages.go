package reposource

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// Exported for [NOTE: npm-tarball-filename-workaround].
	// . is allowed in scope names: for example https://www.npmjs.com/package/@dinero.js/core
	NpmScopeRegexString = `(?P<scope>[\w\-\.]+)`
	// . is allowed in package names: for example https://www.npmjs.com/package/highlight.js
	npmPackageNameRegexString = `(?P<name>[\w\-]+(\.[\w\-]+)*)`
)

var (
	npmScopeRegex          = lazyregexp.New(`^` + NpmScopeRegexString + `$`)
	npmPackageNameRegex    = lazyregexp.New(`^` + npmPackageNameRegexString + `$`)
	scopedPackageNameRegex = lazyregexp.New(
		`^(@` + NpmScopeRegexString + `/)?` +
			npmPackageNameRegexString +
			`@(?P<version>[\w\-]+(\.[\w\-]+)*)$`)
	npmURLRegex = lazyregexp.New(
		`^npm/(` + NpmScopeRegexString + `/)?` +
			npmPackageNameRegexString + `$`)
)

// An npm package of the form (@scope/)?name.
//
// The fields are kept private to reduce risk of not handling the empty scope
// case correctly.
type NpmPackage struct {
	// Optional scope () for a package, can potentially be "".
	// For more details, see https://docs.npmjs.com/cli/v8/using-npm/scope
	scope string
	// Required name for a package, always non-empty.
	name string
}

func NewNpmPackage(scope string, name string) (*NpmPackage, error) {
	if scope != "" && !npmScopeRegex.MatchString(scope) {
		return nil, errors.Errorf("illegal scope %s (allowed characters: 0-9, a-z, A-Z, _, -)", scope)
	}
	if !npmPackageNameRegex.MatchString(name) {
		return nil, errors.Errorf("illegal package name %s (allowed characters: 0-9, a-z, A-Z, _, -)", name)
	}
	return &NpmPackage{scope, name}, nil
}

func (pkg *NpmPackage) Equal(other *NpmPackage) bool {
	return pkg == other || (pkg != nil && other != nil && *pkg == *other)
}

// ParseNpmPackageFromRepoURL is a convenience function to parse a string in a
// 'npm/(scope/)?name' format into an NpmPackage.
func ParseNpmPackageFromRepoURL(urlPath string) (*NpmPackage, error) {
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
	return &NpmPackage{scope, name}, nil
}

// ParseNpmPackageFromPackageSyntax is a convenience function to parse a
// string in a '(@scope/)?name' format into an NpmPackage.
func ParseNpmPackageFromPackageSyntax(pkg string) (*NpmPackage, error) {
	dep, err := ParseNpmDependency(fmt.Sprintf("%s@0", pkg))
	if err != nil {
		return nil, err
	}
	return dep.NpmPackage, nil
}

type NpmPackageSerializationHelper struct {
	Scope string
	Name  string
}

var _ json.Marshaler = &NpmPackage{}
var _ json.Unmarshaler = &NpmPackage{}

func (pkg *NpmPackage) MarshalJSON() ([]byte, error) {
	return json.Marshal(NpmPackageSerializationHelper{pkg.scope, pkg.name})
}

func (pkg *NpmPackage) UnmarshalJSON(data []byte) error {
	var wrapper NpmPackageSerializationHelper
	err := json.Unmarshal(data, &wrapper)
	if err != nil {
		return err
	}
	newPkg, err := NewNpmPackage(wrapper.Scope, wrapper.Name)
	if err != nil {
		return err
	}
	*pkg = *newPkg
	return nil
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (pkg *NpmPackage) RepoName() api.RepoName {
	if pkg.scope != "" {
		return api.RepoName(fmt.Sprintf("npm/%s/%s", pkg.scope, pkg.name))
	}
	return api.RepoName("npm/" + pkg.name)
}

// CloneURL returns a "URL" that can later be used to download a repo.
func (pkg *NpmPackage) CloneURL() string {
	return string(pkg.RepoName())
}

// Format a package using (@scope/)?name syntax.
//
// This is largely for "lower-level" code interacting with the npm API.
//
// In most cases, you want to use NpmDependency's PackageManagerSyntax() instead.
func (pkg *NpmPackage) PackageSyntax() string {
	if pkg.scope != "" {
		return fmt.Sprintf("@%s/%s", pkg.scope, pkg.name)
	}
	return pkg.name
}

// NpmDependency is a "versioned package" for use by npm commands, such as
// `npm install`.
//
// See also: [NOTE: Dependency-terminology]
//
// Reference:  https://docs.npmjs.com/cli/v8/commands/npm-install
type NpmDependency struct {
	*NpmPackage

	// The version or tag (such as "latest") for a dependency.
	//
	// See https://docs.npmjs.com/cli/v8/using-npm/config#tag for more details
	// about tags.
	Version string

	// The URL of the tarball to download. Possibly empty.
	TarballURL string

	// The description of the package. Possibly empty.
	PackageDescription string
}

// ParseNpmDependency parses a string in a '(@scope/)?module@version' format into an NpmDependency.
//
// npm supports many ways of specifying dependencies (https://docs.npmjs.com/cli/v8/commands/npm-install)
// but we only support exact versions for now.
func ParseNpmDependency(dependency string) (*NpmDependency, error) {
	// We use slightly more restrictive validation compared to the official
	// rules (https://github.com/npm/validate-npm-package-name#naming-rules).
	//
	// For example, npm does not explicitly forbid package names with @ in them.
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
	return &NpmDependency{NpmPackage: &NpmPackage{scope, name}, Version: version}, nil
}

func (d *NpmDependency) Description() string {
	return d.PackageDescription
}

type NpmMetadata struct {
	Package *NpmPackage
}

// PackageManagerSyntax returns the dependency in npm/Yarn syntax. The returned
// string can (for example) be passed to `npm install`.
func (d *NpmDependency) PackageManagerSyntax() string {
	return fmt.Sprintf("%s@%s", d.PackageSyntax(), d.Version)
}

func (d *NpmDependency) Scheme() string {
	return "npm"
}

func (d *NpmDependency) PackageVersion() string {
	return d.Version
}

func (d *NpmDependency) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d *NpmDependency) Equal(o *NpmDependency) bool {
	return d == o || (d != nil && o != nil &&
		d.NpmPackage.Equal(o.NpmPackage) &&
		d.Version == o.Version)
}

// Less implements the Less method of the sort.Interface. It sorts
// dependencies by the semantic version in descending order.
// The latest version of a dependency becomes the first element of the slice.
func (d *NpmDependency) Less(other PackageDependency) bool {
	o := other.(*NpmDependency)

	if d.NpmPackage.Equal(o.NpmPackage) {
		return versionGreaterThan(d.Version, o.Version)
	}

	if d.scope == o.scope {
		return d.name > o.name
	}

	return d.scope > o.scope
}
