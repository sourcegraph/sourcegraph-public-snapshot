package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// NOTE: This should be kept up-to-date with cmd/migrator/build.sh so that we "bake in"
// fallback schemas everything we support migrating to.
const maxVersionString = "4.4.0"

// MaxVersion is the highest known released version at the time the migrator was built.
var MaxVersion = func() oobmigration.Version {
	if version, ok := oobmigration.NewVersionFromString(maxVersionString); ok {
		return version
	}

	panic(fmt.Sprintf("malformed maxVersionString %q", maxVersionString))
}()

// MinVersion is the minimum version a migrator can support upgrading to a newer version of
// Sourcegraph.
var MinVersion = oobmigration.NewVersion(3, 20)

// FrozenRevisions are schemas at a point-in-time for which out-of-band migration unit tests
// can continue to run on their last pre-deprecation version. This code is still ran by the
// migrator, but only on a schema shape that existed in the past.
var FrozenRevisions = []string{
	"4.3.0",
}
