#!/usr/bin/env bash

set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../.. || exit 1

SUPPORTED_PLATFORMS=("aarch64-apple-darwin" "x86_64-apple-darwin" "x86_64-unknown-linux-gnu")

# deterines the Cody App filename based on the platform string given
# args:
# - 1: platform string
filename_from() {
  local arch
  local os
  local filename
  arch=$(echo "$1" | cut -d '-' -f1)
  os=$(echo "$1" | cut -d '-' -f3)
  # this is kind of annoying but tauri uses `x86_64` as a target platform but generates
  # bundles using `amd64` on linux
  if [[ ${arch} == "x86_64" && ${os} == "linux" ]]; then
    arch="amd64"
  fi

  case "${os}" in
    "darwin")
      filename="Cody.${version}.${arch}.app.tar.gz"
      ;;
    "linux")
      # note the UNDERSCORES
      filename="cody_${version}_${arch}.AppImage.tar.gz"
      ;;
    *)
      echo "cannot determine filename for unsupported os: ${os}"
      exit 1
      ;;
  esac

  echo "${filename}"
}
#
# short_platform breaks a long form platform strings like <arch>-<vendor>-<os> up into just <arch>-<os>
# ex. x86_64-unknown-linux-gnu becomes x86_64-linux
short_platform() {
  local arch
  local os
  arch=$(echo "$1" | cut -d '-' -f1)
  os=$(echo "$1" | cut -d '-' -f3)

  echo "${arch}-${os}"
}

# local_file given a dir and filename, local_file finds the file locally with that name
# args:
# 1: directory to search in
# 2: filename to search for
local_file() {
  local dir
  local filename
  dir=$1
  filename=$2

  find "${dir}" -type f -name "${filename}"
}

# platform_json_for creates the following json:
# { "<short-platform-key>": {
#   "signature": "<contents of signature file>",
#   "url": "<github release url>"
# }
#
# We determine the release artefact name for the particular platform. The artefact name is then used to
# locate the .sig file locally. If we have a signature for a artefact, get the contents of it as well as
# the github release url for the artefact. Once we have all the values, we return json using the values.
# args:
# 2: platform - platform to generate json for
platform_json_for() {
  local filename
  local key
  local signature_file
  local signature
  local platform

  platform="$1"
  filename=$(filename_from "${platform}")

  signature_file=$(local_file "dist" "${filename}.sig")
  if [[ -e "${signature_file}" ]]; then
    signature=$(cat "${signature_file}")
  else
    signature=""
  fi

  key=$(short_platform "${platform}")

  if [[ -n "${signature:-""}" ]]; then
    echo "${RELEASE_JSON}" | jq -r --arg key "${key}" --arg filename "${filename}" --arg sig "${signature}" \
      '.assets[] | select(.name == $filename) | { ($key): { "signature": $sig, "url": .url }}'
  else
    echo ""
  fi
}

# generate_manifest generates an update manifest with the following structure:
# {
#   "version": "1.2.3",
#   "pub_date": "<github release created At value>"
#   "platforms": {
#     "<short-platform name>": {
#       "signature": "<contents of release signature>"
#       "url": "<github release url>"
#     }
#   }
# }
#
# args:
# 1: version - the version for that should be used in the update manifest
generate_manifest() {
  local manifest
  local version
  local pub_date

  version="$1"

  pub_date="$(echo "${RELEASE_JSON}" | jq -r ".createdAt")"

  # This is the base manifest, which we will append platform json too
  manifest=$(
    cat <<EOF
{
  "version": "${version}",
  "pub_date": "${pub_date}",
  "platforms": {}
}
EOF
  )

  # we loop through our supported platforms. If we do have platform json for the particular platform, we append to our base manifest
  local json
  for platform in "${SUPPORTED_PLATFORMS[@]}"; do
    json="$(platform_json_for "${platform}")"

    if [[ -n ${json} ]]; then
      manifest=$(echo "${manifest}" | jq --argjson platform_json "${json}" '.platforms += $platform_json')
    fi
  done

  echo "${manifest}"
}

version=$(./dev/app/app-version.sh)
release_tag="app-v${version}"

echo "--- :github: fetching release information for tag: ${release_tag}"
RELEASE_JSON=$(gh release view "${release_tag}" --json name,createdAt,assets || echo '')

if [[ -z "${RELEASE_JSON}" ]]; then
  echo "No release information for ${release_tag} - skipping generating update manifest. Verify that there is a release for this tag"
  exit 0
fi

echo "--- generating app update manifest for version: ${version}"
echo "supported platforms in manifest are: ${SUPPORTED_PLATFORMS[*]}"
manifest=$(generate_manifest "${version}")

if [[ ${CI:-""} == "true" ]]; then
  mkdir -p manifest
  echo "${manifest}" | jq >>manifest/app.update.manifest
  buildkite-agent artifact upload manifest/app.update.manifest
else
  echo "--- app update manifest ---"
  echo "${manifest}" | jq
  echo "--- end ---"
fi
