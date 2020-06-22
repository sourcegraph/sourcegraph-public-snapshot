#!/usr/bin/env bash

# This script either generates a report of third-party dependencies, or runs a check that fails
# if there are any unapproved dependencies ('action items').
#
# Please refer to the handbook entry for more details: https://about.sourcegraph.com/handbook/engineering/continuous_integration#third-party-licenses

set -euf -o pipefail

# by default, generate a report. this does not care if there are pending action items
COMMAND="report --format=csv --save=./third-party-licenses/ThirdPartyLicenses.csv --write-headers"

# if LICENSE_CHECK=true, report unapproved dependencies and error if there are any ('action items')
if [[ "${LICENSE_CHECK:-''}" == "true" ]]; then
  COMMAND="action_items"
fi

# prepare dependencies
yarn --mutex network --frozen-lockfile
go clean -modcache # stuff in cache causes strange things to happen
go mod vendor      # go mod download does not work with license_finder

# report license_finder configuration
license_finder permitted_licenses list
license_finder restricted_licenses list
license_finder ignored_groups list
license_finder ignored_dependencies list
license_finder dependencies list

# run license check
license_finder ${COMMAND} --columns=package_manager name version licenses homepage approved
