#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

SRC_GLOB="release-msi\*msi"
DEST_DIR="artifacts"

download_artifacts() {
  src=$1
  dest=$2
  mkdir -p "${dest}"
  buildkite-agent artifact download "${src}" "${dest}"
}



sign_artifacts() {
  local target
  target=$1

  if [[ ${TAURI_PRIVATE_KEY} == "" ]]; then
    echo "Private key not found"
    exit 1
  fi
  if [[ ${TAURI_KEY_PASSWORD} == "" ]]; then
    echo "Private key password not found"
    exit 1
  fi

  pnpm tauri signer sign -k ${TAURI_PRIVATE_KEY} -p ${TAURI_KEY_PASSWORD} ${target}
}

if [[ ${CI:-""} == "true" ]]; then
  mkdir -p ${DEST_DIR}
  download_artifacts "${SRC_GLOB}" "${DEST_DIR}"
else
  echo "This script is meant to run in CI."
  exit 1
fi

pushd .
cd ${DEST_DIR}
# due to how the artifact is downloaded from buildkite, the msi will be in release-msi\*
# so we move everything from release-msi\* to the current dir
mv release-msi/* .
# now lets get the full path to the msi
target=$(find . -name "*.msi")

echo "--- :tauri::pencil: signing ${target}"
sign_artifacts "${target}"
mkdir -p dist
# lets move everything in the current dir to dist ... which *should* be the .msi and the .msi.sig
mv "${DEST_DIR}/*" dist/
popd


echo "--- :satelite: uploading dist/"
ls -lah dist/
buildkite-agent artifact upload "${DIST_DIR}/*"
