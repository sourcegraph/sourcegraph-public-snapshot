#!/usr/bin/env bash

# INPUT ENVIRONMENT VARIABLES
# - VERSION - required in order to find files on GCS
#   - also passed on to other shell scripts
# - APPLE_DEV_ID_APPLICATION_CERT
#   - passed on to the signing scripts
#   - defaults to /mnt/Apple-Developer-ID-Application.p12
#     - the file comes from Secrets in CI
# - APPLE_DEV_ID_APPLICATION_PASSWORD
#   - passed on to the signing scripts
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ID
#   - passed on to the notarize script
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ISSUER
#   - passed on to the notarize script
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_FILE
#   - passed on to the notarize script
#   - defaults to /mnt/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8
#     - the file comes from Secrets in CI

# the shell scripts used to build, sign and notarize are in the same directory
exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# capture the checksums file so we can update it as we create and sign artifacts
gsutil cp "gs://sourcegraph-app-releases/${VERSION}/checksums.txt" .

# which means we need to copy the binaries from GCS
# we're interested in the darwin zip files only
while IFS= read -r gcs_file; do
  zip_file_name=$(basename "${gcs_file}" .zip)
  gsutil cp "${gcs_file}" "${zip_file_name}.zip" || exit 1
  unzip -o "${zip_file_name}.zip" -d "${zip_file_name}"
  [ -f "${zip_file_name}/sourcegraph" ] || exit 1
  artifact="${PWD}/${zip_file_name}/sourcegraph" \
    "${exedir}/sign_macos_binary.sh" || exit 1
  zip -jf9 "${zip_file_name}.zip" "${zip_file_name}/sourcegraph" || exit 1
  grep -v "${zip_file_name}[.]zip" checksums.txt >checksums.txt.2
  sha256sum "${zip_file_name}.zip" >>checksums.txt.2
  mv checksums.txt.2 checksums.txt
  # keep the macOS universal binary around because it'll also be copied into the app bundle
  [[ ${zip_file_name} = sourcegraph_${VERSION}_darwin_all ]] && mv "sourcegraph_${VERSION}_darwin_all/sourcegraph" sourcegraph
  rm -rf "${zip_file_name:-?}/"
  gsutil cp "${zip_file_name}.zip" checksums.txt "gs://sourcegraph-app-releases/${VERSION}/" || exit 1
  rm -f "${zip_file_name}.zip"
done < <(gsutil ls "gs://sourcegraph-app-releases/${VERSION}/*darwin*.zip")

# the macOS universal binary should have been left by the binary signing process
[ -f "sourcegraph" ] || {
  echo "unable to download sourcegraph binary from GCS" 1>&2
  exit 1
}

artifact="${PWD}/sourcegraph" \
  signature="${PWD}/Sourcegraph App.app" \
  "${exedir}/build_macos_app.sh" || exit 1

[ -d "Sourcegraph App.app" ] || {
  echo "failed building the macOS app bundle 1>&2"
  exit 1
}

artifact="${PWD}/Sourcegraph App.app" \
  "${exedir}/sign_macos_app.sh" || exit 1

# this one can take awhile - 5+ minutes
artifact="${PWD}/Sourcegraph App.app" \
  "${exedir}/notarize_macos_app.sh" || exit 1

# we really want to package in a dmg container, but that will take automation on macOS,
# which could work in GH actions, but not in buildkite
# Instead, compress into a zip archive for now
zip -r9 "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" "Sourcegraph App.app"
sha256sum "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" >>checksums.txt
gsutil cp "sourcegraph_${VERSION}_macOS_universal_app_bundle.zip" checksums.txt "gs://sourcegraph-app-releases/${VERSION}/" || exit 1

# replicate the artifacts in a "latest" bucket
while IFS= read -r gcs_file; do
  gsutil cp "${gcs_file}" "${gcs_file//${VERSION}/latest}"
done < <(gsutil ls "gs://sourcegraph-app-releases/${VERSION}/")
# change the checksum names and upload that
sed "s/${VERSION}/latest/g" checksums.txt >checksums.txt.2
mv checksums.txt.2 checksums.txt
gsutil cp checksums.txt "gs://sourcegraph-app-releases/latest/"

# whew; done!
