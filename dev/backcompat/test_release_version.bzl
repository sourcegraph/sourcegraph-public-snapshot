"""Defines the minimum ugradeable version of Sourcegraph.

This designates the mininum version from which we guarantees the newest database
schema can run.

See https://docs.sourcegraph.com/dev/background-information/sql/migrations
"""

# Defines which version we target with the backward compatibility tests.
MINIMUM_UPGRADEABLE_VERSION = "5.0.0"

# Defines a reproducible reference to clone Sourcegraph at to run those tests.
MINIMUM_UPGRADEABLE_VERSION_REF = "3a3606c543bae07d70faab5c147ad91ea5d6eb7d"
