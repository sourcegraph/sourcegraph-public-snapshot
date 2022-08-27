package main

import "github.com/sourcegraph/sourcegraph/internal/oobmigration"

// MinVersion is the minimum version a migrator can support upgrading to a newer version of Sourcegraph.
var MinVersion = oobmigration.NewVersion(3, 25)

// MaxVersion is the highest known released version at the time the migrator was built.
var MaxVersion = oobmigration.NewVersion(4, 0)
