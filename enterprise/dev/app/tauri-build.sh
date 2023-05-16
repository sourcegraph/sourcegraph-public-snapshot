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
  local bundle_path
  bundle_path="./src-tauri/target/release/bundle"
  mkdir -p dist
  src=$(find ${bundle_path} -type f \( -name "Sourcegraph*.dmg" -o -name "Sourcegraph*.app" -o -name "sourcegraph*.deb" -o -name "sourcegraph*.AppImage" -o -name "sourcegraph*.tar.gz" \) -exec realpath {} \;);
  while IFS= read -r from; do
    mv -vf "${from}" "./dist/"
  done <<< ${src}

  # # we have to handle Sourcegraph.App differently since it is a dir
  local app_bundle
  app_bundle="${bundle_path}/macos/Sourcegraph App.app"
  if [[ -d  ${app_bundle} ]]; then
    mv "${app_bundle}" "./dist/"
  fi

  echo --- Uploading artifacts from dist
  ls -al ./dist
  buildkite-agent artifact upload "./dist/*"

}

if [[ ${CI:-""} == "true" ]]; then
  download_artifacts
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
set_version ${VERSION}

echo "--- :tauri: Building Application (${VERSION})"]
NODE_ENV=production pnpm run build-app-shell
pnpm tauri build --bundles deb,appimage,app,dmg

if [[ ${CI:-""} == "true" ]]; then
  upload_dist
fi
