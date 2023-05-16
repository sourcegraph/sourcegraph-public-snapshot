#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

download_artifacts() {
  mkdir -p .bin
  buildkite-agent artifact download ".bin/sourcegraph-backend-*" .bin/
}

set_version() {
  local version
  local tauri_conf
  local tmp
  version=$1
  tauri_conf="./src-tauri/tauri.conf.json"
  tmp=$(mktemp)
  echo "--- Updating package version in '${tauri_conf}' to ${version}"
  jq --arg version "${version}" '.package.version = $version' "${tauri_conf}" > "${tmp}"
  mv "${tmp}" ./src-tauri/tauri.conf.json
}

upload_dist() {
  mkdir -p dist
  src=$(find ./src-tauri/target/release -type f \( -name "Sourcegraph*.dmg" -o -name "sourcegraph*.deb" -o -name "sourcegraph*.AppImage" -o -name "sourcegraph*.tar.gz" \) -exec realpath {} \;)
  for from in ${src}; do
    # remove everything until the last slash
    filename=$(echo ${from} | sed 's|.*/||')
    mv -vf "$from" "./dist/${filename}"
  done
  echo --- Uploading artifacts from dist
  ls -al ./dist
  buildkite-agent artifact upload "./dist/*"

}

if [[ ${CI:-""} == "true" ]]; then
  download_artifacts
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
set_version ${VERSION}

echo "--- [Tauri] Building Application (${VERSION})"]
NODE_ENV=production pnpm run build-app-shell
pnpm tauri build --bundles deb,appimage,app,dmg
upload_dist
