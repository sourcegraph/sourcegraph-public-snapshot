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
	scopedPackageNameWithoutVersionRegex = lazyregexp.New(
		`^(@` + NpmScopeRegexString + `/)?` +
			npmPackageNameRegexString)
	npmURLRegex = lazyregexp.New(
		`^npm/(` + NpmScopeRegexString + `/)?` +
			npmPackageNameRegexString + `$`)
)

// An npm package of the form (@scope/)?name.
//
// The fields are kept private to reduce risk of not handling the empty scope
// case correctly.
type NpmPackageName struct {
	// Optional scope () for a package, can potentially be "".
	// For more details, see https://docs.npmjs.com/cli/v8/using-npm/scope
	scope string
	// Required name for a package, always non-empty.
	name string
}

func NewNpmPackageName(scope string, name string) (*NpmPackageName, error) {
	if scope != "" && !npmScopeRegex.MatchString(scope) {
		return nil, errors.Errorf("illegal scope %s (allowed characters: 0-9, a-z, A-Z, _, -)", scope)
	}
	if !npmPackageNameRegex.MatchString(name) {
		return nil, errors.Errorf("illegal package name %s (allowed characters: 0-9, a-z, A-Z, _, -)", name)
	}
	return &NpmPackageName{scope, name}, nil
}

func (pkg *NpmPackageName) Equal(other *NpmPackageName) bool {
	return pkg == other || (pkg != nil && other != nil && *pkg == *other)
}

// ParseNpmPackageNameWithoutVersion parses a package name with optional scope
// into NpmPackageName.
func ParseNpmPackageNameWithoutVersion(input string) (NpmPackageName, error) {
	match := scopedPackageNameWithoutVersionRegex.FindStringSubmatch(input)
	if match == nil {
		return NpmPackageName{}, errors.Errorf("expected dependency in (@scope/)?name format but found %s", input)
	}
	result := make(map[string]string)
	for i, groupName := range scopedPackageNameWithoutVersionRegex.SubexpNames() {
		if i != 0 && groupName != "" {
			result[groupName] = match[i]
		}
	}
	return NpmPackageName{result["scope"], result["name"]}, nil
}

// ParseNpmPackageFromRepoURL is a convenience function to parse a string in a
// 'npm/(scope/)?name' format into an NpmPackageName.
func ParseNpmPackageFromRepoURL(repoName api.RepoName) (*NpmPackageName, error) {
	match := npmURLRegex.FindStringSubmatch(string(repoName))
	if match == nil {
		return nil, errors.Errorf("expected path in npm/(scope/)?name format but found %s", repoName)
	}
	result := make(map[string]string)
	for i, groupName := range npmURLRegex.SubexpNames() {
		if i != 0 && groupName != "" {
			result[groupName] = match[i]
		}
	}
	scope, name := result["scope"], result["name"]
	return &NpmPackageName{scope, name}, nil
}

// ParseNpmPackageFromPackageSyntax is a convenience function to parse a
// string in a '(@scope/)?name' format into an NpmPackageName.
func ParseNpmPackageFromPackageSyntax(pkg PackageName) (*NpmPackageName, error) {
	dep, err := ParseNpmVersionedPackage(fmt.Sprintf("%s@0", pkg))
	if err != nil {
		return nil, err
	}
	return dep.NpmPackageName, nil
}

type NpmPackageSerializationHelper struct {
	Scope string
	Name  string
}

var _ json.Marshaler = &NpmPackageName{}
var _ json.Unmarshaler = &NpmPackageName{}

func (pkg *NpmPackageName) MarshalJSON() ([]byte, error) {
	return json.Marshal(NpmPackageSerializationHelper{pkg.scope, pkg.name})
}

func (pkg *NpmPackageName) UnmarshalJSON(data []byte) error {
	var wrapper NpmPackageSerializationHelper
	err := json.Unmarshal(data, &wrapper)
	if err != nil {
		return err
	}
	newPkg, err := NewNpmPackageName(wrapper.Scope, wrapper.Name)
	if err != nil {
		return err
	}
	*pkg = *newPkg
	return nil
}

// RepoName provides a name that is "globally unique" for a Sourcegraph instance.
//
// The returned value is used for repo:... in queries.
func (pkg *NpmPackageName) RepoName() api.RepoName {
	if pkg.scope != "" {
		return api.RepoName(fmt.Sprintf("npm/%s/%s", pkg.scope, pkg.name))
	}
	return api.RepoName("npm/" + pkg.name)
}

// CloneURL returns a "URL" that can later be used to download a repo.
func (pkg *NpmPackageName) CloneURL() string {
	return string(pkg.RepoName())
}

// Format a package using (@scope/)?name syntax.
//
// This is largely for "lower-level" code interacting with the npm API.
//
// In most cases, you want to use NpmVersionedPackage's VersionedPackageSyntax() instead.
func (pkg *NpmPackageName) PackageSyntax() PackageName {
	if pkg.scope != "" {
		return PackageName(fmt.Sprintf("@%s/%s", pkg.scope, pkg.name))
	}
	return PackageName(pkg.name)
}

// NpmVersionedPackage is a "versioned package" for use by npm commands, such as
// `npm install`.
//
// Reference:  https://docs.npmjs.com/cli/v8/commands/npm-install
type NpmVersionedPackage struct {
	*NpmPackageName

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

// ParseNpmVersionedPackage parses a string in a '(@scope/)?module@version' format into an NpmVersionedPackage.
//
// npm supports many ways of specifying dependencies (https://docs.npmjs.com/cli/v8/commands/npm-install)
// but we only support exact versions for now.
//
// Some packages have names containing multiple '/' characters.
// (https://sourcegraph.com/search?q=context:global+file:package.json%24+%22name%22:+%22%40%5B%5E%5Cn/%5D%2B/%5B%5E%5Cn/%5D%2B/%5B%5E%5Cn%5D%2B%5C%22&patternType=regexp)
// So it is possible for indexes to reference packages by that name,
// but such names are not supported by recent npm versions, so we don't
// allow those here.
func ParseNpmVersionedPackage(dependency string) (*NpmVersionedPackage, error) {
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
	return &NpmVersionedPackage{NpmPackageName: &NpmPackageName{scope, name}, Version: version}, nil
}

func (d *NpmVersionedPackage) Description() string {
	return d.PackageDescription
}

type NpmMetadata struct {
	Package *NpmPackageName
}

// PackageManagerSyntax returns the dependency in npm/Yarn syntax. The returned
// string can (for example) be passed to `npm install`.
func (d *NpmVersionedPackage) VersionedPackageSyntax() string {
	return fmt.Sprintf("%s@%s", d.PackageSyntax(), d.Version)
}

func (d *NpmVersionedPackage) Scheme() string {
	return "npm"
}

func (d *NpmVersionedPackage) PackageVersion() string {
	return d.Version
}

func (d *NpmVersionedPackage) GitTagFromVersion() string {
	return "v" + d.Version
}

func (d *NpmVersionedPackage) Equal(o *NpmVersionedPackage) bool {
	return d == o || (d != nil && o != nil &&
		d.NpmPackageName.Equal(o.NpmPackageName) &&
		d.Version == o.Version)
}

// Less implements the Less method of the sort.Interface. It sorts
// dependencies by the semantic version in descending order.
// The latest version of a dependency becomes the first element of the slice.
func (d *NpmVersionedPackage) Less(other VersionedPackage) bool {
	o := other.(*NpmVersionedPackage)

	if d.NpmPackageName.Equal(o.NpmPackageName) {
		return versionGreaterThan(d.Version, o.Version)
	}

	if d.scope == o.scope {
		return d.name > o.name
	}

	return d.scope > o.scope
}
