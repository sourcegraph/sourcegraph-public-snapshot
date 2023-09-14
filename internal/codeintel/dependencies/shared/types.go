package shared

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
)

type PackageRepoReference struct {
	ID            int
	Scheme        string
	Name          reposource.PackageName
	Versions      []PackageRepoRefVersion
	Blocked       bool
	LastCheckedAt *time.Time
}

type PackageRepoRefVersion struct {
	ID            int
	PackageRefID  int
	Version       string
	Blocked       bool
	LastCheckedAt *time.Time
}

type MinimalPackageRepoRef struct {
	Scheme        string
	Name          reposource.PackageName
	Versions      []MinimalPackageRepoRefVersion
	Blocked       bool
	LastCheckedAt *time.Time
}

type MinimalPackageRepoRefVersion struct {
	Version       string
	Blocked       bool
	LastCheckedAt *time.Time
}

type MinimialVersionedPackageRepo struct {
	Scheme  string
	Name    reposource.PackageName
	Version string
}

type MinimalPackageFilter struct {
	PackageScheme string
	Behaviour     *string
	NameFilter    *struct {
		PackageGlob string
	}
	VersionFilter *struct {
		PackageName string
		VersionGlob string
	}
}

type PackageRepoFilter struct {
	ID            int
	Behaviour     string
	PackageScheme string
	NameFilter    *struct {
		PackageGlob string
	}
	VersionFilter *struct {
		PackageName string
		VersionGlob string
	}
	DeletedAt *time.Time
	UpdatedAt time.Time
}
