package database

import (
	"context"
	"database/sql"

	"github.com/Masterminds/semver"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
)

type FreeLicenseStore interface {
	basestore.ShareableStore
	Init(ctx context.Context) (*FreeLicense, error)
}

type freeLicenseStore struct {
	*basestore.Store
}

var _ FreeLicenseStore = (*freeLicenseStore)(nil)

type FreeLicense struct {
	LicenseKey string
	Version    int
}

var free0License = FreeLicense{
	LicenseKey: "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJkRFBrQnZmMit0OTBSK3pibHI4bllZdHN0dnNia25LUVdaYlJhMHgzeXdMaUt1WmxYR1JMZTNFdmlkQkFZUXQ2KzZQdlBzRmpsWUJSNkF5YjF4V3haenNxYzN4alZUYXNjeGt6UUE2eFFDbDVQSkZKcFJmRDg3RVN1WjBMQXFLT2gydSs0cVBPczJPdFMrK1ltWEw2azBQODNRWGViVFdUN1AxeFVhSlduU1FpNUhkK21vOWlMNWJGQ1hVNFp3VWFaMnBoYmVUK3JkNTlMd2h1eFN4cGExQUhiNFdESTVQTjBJOEtXM2FuVk1HYmpaMVpKV1ZjTVRES3c4NkpOaDZ0SU53NlQ1Y2J1WXBNTHM5U3MvWmgzeVFHUWxZdEVOM2JVUUlDUmUvbGg0VVRLMVIzZ2sxUjZja2VGWG52VnNFM1h2MVYyZ3JoK3NvdmFSUkwydHhpOHc9PSIsIlJlc3QiOm51bGx9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTRMRE13TERJeU15d3hPU3c0TWl3eE9DdzNNeXc0TWwwc0luUWlPbHNpY0d4aGJqcG1jbVZsTFRBaVhTd2lkU0k2TVRBc0ltVWlPaUl5TVRNM0xUQXlMVEF4VkRBME9qQTRPakU1V2lKOSJ9",
	Version:    1,
}

var free0LicenseInfo = license.Info{
	Tags:      []string{"plan:free-0"},
	UserCount: 10,
}

var free1License = FreeLicense{
	LicenseKey: "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJkRFBrQnZmMit0OTBSK3pibHI4bllZdHN0dnNia25LUVdaYlJhMHgzeXdMaUt1WmxYR1JMZTNFdmlkQkFZUXQ2KzZQdlBzRmpsWUJSNkF5YjF4V3haenNxYzN4alZUYXNjeGt6UUE2eFFDbDVQSkZKcFJmRDg3RVN1WjBMQXFLT2gydSs0cVBPczJPdFMrK1ltWEw2azBQODNRWGViVFdUN1AxeFVhSlduU1FpNUhkK21vOWlMNWJGQ1hVNFp3VWFaMnBoYmVUK3JkNTlMd2h1eFN4cGExQUhiNFdESTVQTjBJOEtXM2FuVk1HYmpaMVpKV1ZjTVRES3c4NkpOaDZ0SU53NlQ1Y2J1WXBNTHM5U3MvWmgzeVFHUWxZdEVOM2JVUUlDUmUvbGg0VVRLMVIzZ2sxUjZja2VGWG52VnNFM1h2MVYyZ3JoK3NvdmFSUkwydHhpOHc9PSIsIlJlc3QiOm51bGx9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTRMRE13TERJeU15d3hPU3c0TWl3eE9DdzNNeXc0TWwwc0luUWlPbHNpY0d4aGJqcG1jbVZsTFRBaVhTd2lkU0k2TVRBc0ltVWlPaUl5TVRNM0xUQXlMVEF4VkRBME9qQTRPakU1V2lKOSJ9",
	Version:    1,
}

var free1LicenseInfo = license.Info{
	Tags:      []string{"plan:free-1"},
	UserCount: 10,
}

// Init initializes the Sourcegraph free license. This license is
// used when no license key is configured in the site configuration.
// The first time Init is called, the current free license is stored in
// the database. Subsequent calls to Init will return that same license for
// as long as the entry remains in the database, even if the free license
// plan changes.
// This function must be called before any code that needs to do license checks.
func (s *freeLicenseStore) Init(ctx context.Context) (*FreeLicense, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT license_key, license_version FROM free_license LIMIT 1"))
	var license FreeLicense

	err := row.Scan(&license.LicenseKey, &license.Version)
	if err == nil {
		licensing.FreeLicenseKey = license.LicenseKey
		return &license, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	freeLicense, freeLicenseInfo := s.getDefaultFreeLicense(ctx)

	// Try to generate a free license key first in case a custom generation key is set.
	license.LicenseKey, license.Version, err = licensing.GenerateProductLicenseKey(freeLicenseInfo)
	if err != nil {
		// If that does not work, fall back to the default free license.
		license.LicenseKey = freeLicense.LicenseKey
		license.Version = freeLicense.Version
	}
	err = s.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO free_license (license_key, license_version) VALUES (%s, %d)",
		license.LicenseKey,
		license.Version,
	))

	if err != nil {
		return nil, err
	}

	// Set the global free license key
	licensing.FreeLicenseKey = license.LicenseKey
	return &license, nil
}

func (s *freeLicenseStore) getDefaultFreeLicense(ctx context.Context) (FreeLicense, license.Info) {
	upgradeStore := upgradestore.NewWith(s.Store.Handle())
	firstVersion, ok, err := upgradeStore.GetFirstServiceVersion(ctx, "frontend")

	// If an error occurs, or it is a fresh instance, we use the latest default license
	if !ok || err != nil {
		return free1License, free1LicenseInfo
	}
	// If it is not a fresh instance, we try to parse the first version
	firstVersionSemver, err := semver.NewVersion(firstVersion)
	if err != nil {
		// If semver cannot be parsed (i.e. it's not a release build)
		// we use the latest default license
		return free1License, free1LicenseInfo
	}

	// If the first version is before 4.4, we use the old default license
	if firstVersionSemver.LessThan(semver.MustParse("4.4.0")) {
		return free0License, free0LicenseInfo
	}

	return free1License, free1LicenseInfo
}
