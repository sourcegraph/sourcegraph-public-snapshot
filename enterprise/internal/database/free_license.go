package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type FreeLicenseStore interface {
	basestore.ShareableStore
	Init(ctx context.Context) (*FreeLicense, error)
	Get(ctx context.Context) (*FreeLicense, error)
}

type freeLicenseStore struct {
	*basestore.Store
}

var _ FreeLicenseStore = (*freeLicenseStore)(nil)

func (s *freeLicenseStore) Init(ctx context.Context) (*FreeLicense, error) {
	expiresAt, _ := time.Parse(time.RFC3339, "2137-02-01T04:08:19Z")
	freeLicense := &FreeLicense{
		LicenseKey: "eyJzaWciOnsiRm9ybWF0Ijoic3NoLXJzYSIsIkJsb2IiOiJkRFBrQnZmMit0OTBSK3pibHI4bllZdHN0dnNia25LUVdaYlJhMHgzeXdMaUt1WmxYR1JMZTNFdmlkQkFZUXQ2KzZQdlBzRmpsWUJSNkF5YjF4V3haenNxYzN4alZUYXNjeGt6UUE2eFFDbDVQSkZKcFJmRDg3RVN1WjBMQXFLT2gydSs0cVBPczJPdFMrK1ltWEw2azBQODNRWGViVFdUN1AxeFVhSlduU1FpNUhkK21vOWlMNWJGQ1hVNFp3VWFaMnBoYmVUK3JkNTlMd2h1eFN4cGExQUhiNFdESTVQTjBJOEtXM2FuVk1HYmpaMVpKV1ZjTVRES3c4NkpOaDZ0SU53NlQ1Y2J1WXBNTHM5U3MvWmgzeVFHUWxZdEVOM2JVUUlDUmUvbGg0VVRLMVIzZ2sxUjZja2VGWG52VnNFM1h2MVYyZ3JoK3NvdmFSUkwydHhpOHc9PSIsIlJlc3QiOm51bGx9LCJpbmZvIjoiZXlKMklqb3hMQ0p1SWpwYk1qSTRMRE13TERJeU15d3hPU3c0TWl3eE9DdzNNeXc0TWwwc0luUWlPbHNpY0d4aGJqcG1jbVZsTFRBaVhTd2lkU0k2TVRBc0ltVWlPaUl5TVRNM0xUQXlMVEF4VkRBME9qQTRPakU1V2lKOSJ9",
		Version:    1,
		Info: license.Info{
			Tags:      []string{"plan:free-0"},
			UserCount: 10,
			ExpiresAt: expiresAt,
		},
	}

	err := s.Exec(ctx, sqlf.Sprintf(
		"INSERT INTO free_license (license_key, version, info) VALUES (%s, %d, %s)",
		freeLicense.LicenseKey,
		freeLicense.Version,
		freeLicense.Info,
	))

	if err != nil {
		return nil, err
	}

	return freeLicense, nil
}

type FreeLicense struct {
	LicenseKey string
	Version    int
	Info       license.Info
}

func (s *freeLicenseStore) Get(ctx context.Context) (*FreeLicense, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf("SELECT * FROM free_license"))

	var freeLicense FreeLicense
	if err := row.Scan(&freeLicense.LicenseKey, &freeLicense.Version, &freeLicense.Info); err != nil {
		return nil, err
	}

	return &freeLicense, nil
}
