#!/usr/bin/env bash

set -euf -o pipefail

# by default, generate a report. this does not care if there are pending action items
COMMAND="report --format=csv --save=ThirdPartyLicenses.csv --write-headers"

# if LICENSE_CHECK=true, report unapproved dependencies and error if there are any ('action items')
if [[ "${LICENSE_CHECK:-''}" == "true" ]]; then
  COMMAND="action_items"
fi

# prepare dependencies
yarn --mutex network --frozen-lockfile
go mod vendor # go mod download does not work with license_finder

# run license check
license_finder ${COMMAND} --columns=package_manager name version licenses homepage approved
