#!/usr/bin/env bash

set -euf -o pipefail

COMMAND="report --format=csv --save=ThirdPartyLicenses.csv --write-headers"

# if LICENSE_CHECK=true, report unapproved dependencies and error if there are any
if [[ "${LICENSE_CHECK:-''}" == "true" ]]; then
  COMMAND="action_items"
fi

# based on https://github.com/pivotal/LicenseFinder/blob/master/dlf
docker run -v "$(pwd)":/scan -it licensefinder/license_finder \
  /bin/bash -lc "cd /scan && license_finder ${COMMAND} --prepare --columns=package_manager name version licenses homepage approved"
