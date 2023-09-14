package migrations

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// currentVersion returns the version that should be given to the out of band migration runner.
// In dev mode, we use the _next_ (unreleased) version so that we're always on the bleeding edge.
// When running from a tagged release, we'll use the baked-in version string constant.
func currentVersion(logger log.Logger) (oobmigration.Version, error) {
	if rawVersion := version.Version(); !version.IsDev(rawVersion) {
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

	// TODO: @jhchabran
	// The infer mechanism doesn't work in CI, because we weren't expecting to run a container
	// with a 0.0.0+dev version. This fixes it. We should come back to this.
	if version.IsDev(version.Version()) && os.Getenv("BAZEL_SKIP_OOB_INFER_VERSION") != "" {
		return oobmigration.NewVersion(5, 99), nil
	}

	version, err := inferNextReleaseVersion()
	if err != nil {
		return oobmigration.Version{}, err
	}

	logger.Info("Using latest tag as current version", log.String("version", version.String()))
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

// inferNextReleaseVersion returns the version AFTER the latest tagged release.
func inferNextReleaseVersion() (oobmigration.Version, error) {
	wd, err := os.Getwd()
	if err != nil {
		return oobmigration.Version{}, err
	}

	cmd := exec.Command("git", "tag", "--list", "v*")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return oobmigration.Version{}, err
	}

	tagMap := map[string]struct{}{}
	for _, tag := range strings.Split(string(output), "\n") {
		tag = strings.Split(tag, "-")[0] // strip off rc suffix if it exists

		if version, ok := oobmigration.NewVersionFromString(tag); ok {
			tagMap[version.String()] = struct{}{}
		}
	}

	versions := make([]oobmigration.Version, 0, len(tagMap))
	for tag := range tagMap {
		version, _ := oobmigration.NewVersionFromString(tag)
		versions = append(versions, version)
	}
	oobmigration.SortVersions(versions)

	if len(versions) == 0 {
		return oobmigration.Version{}, errors.New("failed to find tagged version")
	}

	// Get highest release and bump by one
	return versions[len(versions)-1].Next(), nil
}
