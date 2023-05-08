#!/usr/bin/env bash

function version() {
  date=$(date '+%Y.%m.%d')

  if [[ ${CI} == "true" ]]; then
    sha=$(git rev-parse --short HEAD)
  else
    sha="dev"
  fi


  echo "${date}+${sha}"
}

appVersion="$(version)"


stamp_version="${VERSION:-${appVersion}}"

echo STABLE_VERSION "$stamp_version"
echo VERSION_TIMESTAMP "$(date +%s)"
