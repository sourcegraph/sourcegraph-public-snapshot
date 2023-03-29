#!/usr/bin/env bash

# INPUT ENVIRONMENT VARIABLES
# - VERSION - required in order to find files on GCS
#   - also used by to other shell scripts
# - APPLE_DEV_ID_APPLICATION_CERT
#   - used by to the signing scripts
#   - defaults to /mnt/Apple-Developer-ID-Application.p12
#     - the file comes from Secrets in CI
# - APPLE_DEV_ID_APPLICATION_PASSWORD
#   - used by the signing scripts
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ID
#   - used by the notarize script
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ISSUER
#   - used by the notarize script
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_FILE
#   - used by the notarize script
#   - defaults to /mnt/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8
#     - the file comes from Secrets in CI

log() {
  # `log DEBUG|INFO|WARNING|ERROR <message>`
  printf '%s %s %s\n' "$(date +'%Y-%m-%d %H:%M:%S')" "${1}" "${2}"
}
debug() {
  log "DEBUG" "${1}"
}
info() {
  log "INFO" "${1}"
}
warn() {
  log "WARN" "${1}"
}
error() {
  log "ERROR" "${1}"
}

info "START processing macos artifacts"

# the shell scripts used to build, sign and notarize are in the same directory
exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# capture the checksums file so we can update it as we create and sign artifacts
gsutil cp "gs://sourcegraph-app-releases/${VERSION}/checksums.txt" .

# copy the binaries from GCS
# we're interested in the darwin zip files only
while IFS= read -r gcs_file; do
  zip_file_name=$(basename "${gcs_file}" .zip)
  info "signing ${zip_file_name}.zip"
  gsutil cp "${gcs_file}" "${zip_file_name}.zip" || {
    error "failed to download ${gcs_file}"
    exit 1
  }
  unzip -o "${zip_file_name}.zip" -d "${zip_file_name}"
  [ -f "${zip_file_name}/sourcegraph" ] || {
    error "failed to extract ${zip_file_name}.zip"
    exit 1
  }

  artifact="${PWD}/${zip_file_name}/sourcegraph" \
    "${exedir}/sign_macos_binary.sh" || {
    error "failed signing ${zip_file_name}/sourcegraph"
    exit 1
  }
  zip -jf9 "${zip_file_name}.zip" "${zip_file_name}/sourcegraph" || {
    error "failed to archive ${zip_file_name}/sourcegraph"
    exit 1
  }

  # notarize the binary
  # can't staple to standalone executables
  info "notarizing ${zip_file_name}.zip"

  "${exedir}/notarize_macos_artifact.sh" \
    "${PWD}/${zip_file_name}.zip" || {
    error "failed notarizing ${zip_file_name}.zip"
    exit 1
  }

  grep -v "${zip_file_name}[.]zip" checksums.txt >checksums.txt.2
  sha256sum "${zip_file_name}.zip" >>checksums.txt.2
  mv checksums.txt.2 checksums.txt
  # keep the macOS universal binary around because it'll also be copied into the app bundle
  [[ ${zip_file_name} = sourcegraph_${VERSION}_darwin_all ]] && mv "sourcegraph_${VERSION}_darwin_all/sourcegraph" sourcegraph
  rm -rf "${zip_file_name:-?}/"
  gsutil cp "${zip_file_name}.zip" checksums.txt "gs://sourcegraph-app-releases/${VERSION}/" || {
    error "failed to upload ${zip_file_name}.zip"
    exit 1
  }
  rm -f "${zip_file_name}.zip"
  ### limit binary signing to only the universal binary because the arch-specific ones are for Homebrew
done < <(gsutil ls "gs://sourcegraph-app-releases/${VERSION}/sourcegraph_${VERSION}_darwin_all.zip")

# the macOS universal binary should have been left by the binary signing process
[ -f "sourcegraph" ] || {
  error "unable to download sourcegraph binary from GCS"
  exit 1
}

info "building macOS App bundle"

artifact="${PWD}/sourcegraph" \
  signature="${PWD}/Sourcegraph App.app" \
  "${exedir}/build_macos_app.sh" || exit 1

[ -d "Sourcegraph App.app" ] || {
  error "failed building the macOS app bundle"
  exit 1
}

info "signing macOS App bundle"

artifact="${PWD}/Sourcegraph App.app" \
  "${exedir}/sign_macos_app.sh" || {
  error "failed signing the macOS app bundle"
  exit 1
}

info "notarizing macOS App bundle"

# this one can take awhile - 5+ minutes
"${exedir}/notarize_macos_artifact.sh" \
  --staple \
  "${PWD}/Sourcegraph App.app" || {
  error "failed notarizing the macOS app bundle"
  exit 1
}

info "uploading macOS App bundle"

# we really want to package in a dmg container, but that will take automation on macOS,
# which could work in GH actions, but not in buildkite
# Instead, compress into a zip archive for now
zip -ry9 "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" "Sourcegraph App.app"
sha256sum "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" >>checksums.txt
gsutil cp "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" checksums.txt "gs://sourcegraph-app-releases/${VERSION}/" || {
  error "failed uploading the macOS app bundle"
  exit 1
}

# if we're running on macOS, build a dmg container as well as a zip archive
command -v hdiutil 1>/dev/null 2>&1 && command -v osascript 1>/dev/null 2>&1 && {

  info "creating the macOS dmg container"
  "${exedir}/macos_app/create_sourcegraph_app_dmg.sh" "${PWD}/Sourcegraph App.app" || exit 1

  # sign the dmg
  info "signing the macOS dmg container"
  "${exedir}/sign_macos_artifact.sh" "${PWD}/Sourcegraph App.dmg" || exit 1

  # gotta notarize the dmg also, even though the app is already notarized!
  info "notarizing the macOS dmg container"
  "${exedir}/notarize_macos_artifact.sh" --staple "${PWD}/Sourcegraph App.dmg" || exit 1

  # add its checksum to the checksums file
  sha256sum "Sourcegraph App.dmg" >>checksums.txt || exit 1

  info "uploading the macOS dmg container"

  # and upload it
  gsutil cp "${PWD}/Sourcegraph App.dmg" checksums.txt "gs://sourcegraph-app-releases/${VERSION}/" || exit 1
}

info "replicating artifacts to /latest"

# replicate the artifacts in a "latest" bucket
while IFS= read -r gcs_file; do
  gsutil cp "${gcs_file}" "${gcs_file//${VERSION}/latest}"
done < <(gsutil ls "gs://sourcegraph-app-releases/${VERSION}/")
# change the checksum names and upload that
sed "s/${VERSION}/latest/g" checksums.txt >checksums-latest.txt
gsutil cp checksums-latest.txt gs://sourcegraph-app-releases/latest/checksums.txt
# include a file with the version number so we can track it for updates and such
printf '%s' "${VERSION}" >version.txt
gsutil cp version.txt gs://sourcegraph-app-releases/latest/

# whew; done!

info "DONE processing macOS artifacts"
