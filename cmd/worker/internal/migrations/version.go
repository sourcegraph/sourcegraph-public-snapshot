package migrations

import (
	"os"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// currentVersion returns the version that should be given to the out of band migration runner.
// In dev mode, we use the _next_ (unreleased) version so that we're always on the bleeding edge.
// When running from a tagged release, we'll use the baked-in version string constant.
func currentVersion(logger log.Logger) (oobmigration.Version, error) {
	if rawVersion := version.Version(); !version.IsDev(rawVersion) && !deploy.IsApp() {
		version, ok := parseVersion(rawVersion)
		if !ok {
			return oobmigration.Version{}, errors.Newf("failed to parse current version: %q", rawVersion)
		}

		return version, nil
	}

	if rawVersion := os.Getenv("SRC_OOBMIGRATION_CURRENT_VERSION"); rawVersion != "" {
		version, ok := oobmigration.NewVersionFromString(rawVersion)
		if !ok {
			return oobmigration.Version{}, errors.Newf("failed to parse force-supplied version: %q", rawVersion)
		}

		return version, nil
	}

	// This should cover both app and dev although there may be some situations where the old infer logic
	// is better for dev.
	version, ok := parseVersion(version.MigrationEndVersion())
	if !ok {
		return oobmigration.Version{}, errors.Newf("failed to parse latest release version: %s", version)
	}
	logger.Info("Using latest release to determine current version", log.String("version", version.String()))
	return version, nil

}

// parseVersion reads the Sourcegraph instance version set at build time. If the given string cannot
// be parsed as one of the following formats, a false-valued flag is returned.
//
// Tagged release format: `v1.2.3`
// Continuous release format: `(ef-feat_)?12345_2006-01-02-1.2-deadbeefbabe(_patch)?`
// App release format: `2023.03.23+204874.db2922`
// App insiders format: `2023.03.23-insiders+204874.db2922`
func parseVersion(rawVersion string) (oobmigration.Version, bool) {
	version, ok := oobmigration.NewVersionFromString(rawVersion)
	if ok {
		return version, true
	}

	parts := strings.Split(rawVersion, "_")
	if len(parts) > 0 && parts[len(parts)-1] == "patch" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) > 0 {
		return oobmigration.NewVersionFromString(strings.Split(parts[len(parts)-1], "-")[0])
	}

	return oobmigration.Version{}, false
}
