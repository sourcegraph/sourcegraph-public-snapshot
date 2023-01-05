package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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

var defaultFreeLicense = &FreeLicense{
	LicenseKey: "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJkRFBrQnZmMit0OTBSK3pibHI4bllZdHN0dnNia25LUVdaYlJhMHgzeXdMaUt1WmxYR1JMZTNFdmlkQkFZUXQ2KzZQdlBzRmpsWUJSNkF5YjF4V3haenNxYzN4alZUYXNjeGt6UUE2eFFDbDVQSkZKcFJmRDg3RVN1WjBMQXFLT2gydSs0cVBPczJPdFMrK1ltWEw2azBQODNRWGViVFdUN1AxeFVhSlduU1FpNUhkK21vOWlMNWJGQ1hVNFp3VWFaMnBoYmVUK3JkNTlMd2h1eFN4cGExQUhiNFdESTVQTjBJOEtXM2FuVk1HYmpaMVpKV1ZjTVRES3c4NkpOaDZ0SU53NlQ1Y2J1WXBNTHM5U3MvWmgzeVFHUWxZdEVOM2JVUUlDUmUvbGg0VVRLMVIzZ2sxUjZja2VGWG52VnNFM1h2MVYyZ3JoK3NvdmFSUkwydHhpOHc9PSIsIlJlc3QiOm51bGx9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTRMRE13TERJeU15d3hPU3c0TWl3eE9DdzNNeXc0TWwwc0luUWlPbHNpY0d4aGJqcG1jbVZsTFRBaVhTd2lkU0k2TVRBc0ltVWlPaUl5TVRNM0xUQXlMVEF4VkRBME9qQTRPakU1V2lKOSJ9",
	Version:    1,
}

var defaultLicenseInfo = license.Info{
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

	// Try to generate a free license first in case a custom generation key is set.
	license.LicenseKey, license.Version, err = licensing.GenerateProductLicenseKey(defaultLicenseInfo)
	if err != nil {
		// If that does not work, fall back to the default free license.
		license.LicenseKey = defaultFreeLicense.LicenseKey
		license.Version = defaultFreeLicense.Version
	}
	err = s.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO free_license (id, license_key, license_version) VALUES (%s, %s, %d)",
		uuid.New().String(),
		license.LicenseKey,
		license.Version,
	))

	if err != nil {
		return nil, err
	}

	licensing.FreeLicenseKey = license.LicenseKey
	return &license, nil
}
