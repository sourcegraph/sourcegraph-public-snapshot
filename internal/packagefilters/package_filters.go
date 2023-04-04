package packagefilters

import (
	"github.com/gobwas/glob"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PackageFilters struct {
	allowlists map[string][]PackageMatcher
	blocklists map[string][]PackageMatcher
}

func NewFilterLists(filters []shared.PackageRepoFilter) (p PackageFilters, err error) {
	p.allowlists = make(map[string][]PackageMatcher)
	p.blocklists = make(map[string][]PackageMatcher)

	for _, filter := range filters {
		var matcher PackageMatcher
		if filter.NameFilter != nil {
			matcher, err = NewPackageNameGlob(filter.NameFilter.PackageGlob)
			if err != nil {
				return PackageFilters{}, errors.Wrapf(err, "error building glob matcher for %q", filter.NameFilter.PackageGlob)
			}
		} else {
			matcher, err = NewVersionGlob(filter.VersionFilter.PackageName, filter.VersionFilter.VersionGlob)
			if err != nil {
				return PackageFilters{}, errors.Wrapf(err, "error building glob matcher for %q %q", filter.VersionFilter.PackageName, filter.VersionFilter.VersionGlob)
			}
		}
		switch filter.Behaviour {
		case "ALLOW":
			p.allowlists[filter.PackageScheme] = append(p.allowlists[filter.PackageScheme], matcher)
		case "BLOCK":
			p.blocklists[filter.PackageScheme] = append(p.blocklists[filter.PackageScheme], matcher)
		}
	}

	return
}

func IsPackageAllowed(scheme string, pkgName reposource.PackageName, filters PackageFilters) (allowed bool) {
	// blocklist takes priority
	for _, block := range filters.blocklists[scheme] {
		// non-all-encompassing version globs don't apply to unversioned packages,
		// likely we're at too-early point in the syncing process to know, but also
		// we may still want the package to browse versions that _dont_ match this
		if vglob, ok := block.(versionGlob); ok && vglob.globStr != "*" {
			continue
		}

		if block.Matches(pkgName, "") {
			return false
		}
	}

	// package is not blocked; we'll now check for (preliminarily) allowing the package.
	//
	// - allow if allow filters are empty (no restrictions)
	// - allow if any name filter matches it
	// - allow if any version filter applies to this name (it _may_ allow at least one version, but we can't know that yet)

	var (
		namesAllowlist    []PackageMatcher
		versionsAllowlist []versionGlob
	)
	for _, allow := range filters.allowlists[scheme] {
		if _, ok := allow.(packageNameGlob); ok {
			namesAllowlist = append(namesAllowlist, allow)
		} else {
			versionsAllowlist = append(versionsAllowlist, allow.(versionGlob))
		}
	}

	isAllowed := len(filters.allowlists[scheme]) == 0
	for _, allow := range namesAllowlist {
		isAllowed = isAllowed || allow.Matches(pkgName, "")
	}

	for _, allow := range versionsAllowlist {
		isAllowed = isAllowed || allow.packageName == string(pkgName)
	}

	return isAllowed
}

func IsVersionedPackageAllowed(scheme string, pkgName reposource.PackageName, version string, filters PackageFilters) (allowed bool) {
	// blocklist takes priority
	for _, block := range filters.blocklists[scheme] {
		if block.Matches(pkgName, version) {
			return false
		}
	}

	// by default, anything is allowed unless specific allowlist exists
	isAllowed := len(filters.allowlists[scheme]) == 0
	for _, allow := range filters.allowlists[scheme] {
		isAllowed = isAllowed || allow.Matches(pkgName, version)
	}

	return isAllowed
}

type PackageMatcher interface {
	Matches(pkg reposource.PackageName, version string) bool
}

type packageNameGlob struct {
	g glob.Glob
}

func NewPackageNameGlob(nameGlob string) (PackageMatcher, error) {
	g, err := glob.Compile(nameGlob)
	if err != nil {
		return nil, err
	}
	return packageNameGlob{g}, nil
}

func (p packageNameGlob) Matches(pkg reposource.PackageName, _ string) bool {
	// when the package name is to be glob matched, we dont
	// care about the version
	return p.g.Match(string(pkg))
}

type versionGlob struct {
	packageName string
	globStr     string
	g           glob.Glob
}

func NewVersionGlob(packageName, vglob string) (PackageMatcher, error) {
	g, err := glob.Compile(vglob)
	if err != nil {
		return nil, err
	}
	return versionGlob{packageName, vglob, g}, nil
}

func (v versionGlob) Matches(pkg reposource.PackageName, version string) bool {
	// when the version is to be glob matched, the package name
	// has to match exactly
	return string(pkg) == v.packageName && v.g.Match(version)
}
