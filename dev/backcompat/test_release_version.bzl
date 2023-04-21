"""Defines the minimum ugradeable version of Sourcegraph.

This designates the mininum version from which we guarantees the newest database
schema can run.

See https://docs.sourcegraph.com/dev/background-information/sql/migrations
"""

# Defines which version we target with the backward compatibility tests.
MINIMUM_UPGRADEABLE_VERSION = "5.0.0"

# Defines a reproducible reference to clone Sourcegraph at to run those tests.
MINIMUM_UPGRADEABLE_VERSION_REF = "653cb673b7785cf27d9b2e4b43e703c8dc23beaf"
