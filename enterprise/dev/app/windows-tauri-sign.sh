#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

SRC_GLOB="win-msi\*msi"
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
target=$(find . -name "*.msi")

echo "--- :tauri::pencil: signing ${target}"
sign_artifacts ${DEST_DIR}
# this is also subtly works around the fact that in `create-github-release` our download glob is `dist/` (notice the forward) slash
# and since we're uploading it here we might as well upload the msi again
mkdir -p dist
mv "${DEST_DIR}/*" dist/
popd


echo "--- :satelite: uploading dist/"
ls -lah dist/
buildkite-agent artifact upload "${DIST_DIR}/*"
