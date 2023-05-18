#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

BIN_DIR=".bin"
DIST_DIR="dist"

download_artifacts() {
  mkdir -p .bin
  buildkite-agent artifact download "${BIN_DIR}/sourcegraph-backend-*" .bin/
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
  # we're using a while and read here because the paths may contain spaces and this handles it properly
  while IFS= read -r from; do
    mv -vf "${from}" "./${DIST_DIR}/"
  done <<< ${src}

  # # we have to handle Sourcegraph.App differently since it is a dir
  local app_bundle
  app_bundle="${bundle_path}/macos/Sourcegraph App.app"
  if [[ -d  ${app_bundle} ]]; then
    mv "${app_bundle}" "./${DIST_DIR}/"
  fi

  echo --- Uploading artifacts from dist
  ls -al ./dist
  buildkite-agent artifact upload "./${DIST_DIR}/*"

}

cleanup_codesigning() {
    if [[ -n $(security list-keychains -d user | grep "my_temporary_keychain") ]]; then
      set +e
      echo "--- :broom: cleaning up keychains"
      security delete-keychain my_temporary_keychain.keychain
      set -e
    fi
}

pre_codesign() {
  local binary_path
  binary_path=$1
  # Tauri won't code sign our sidecar sourcegraph-backend Go binary for us, so we need to do it on
  # our own. https://github.com/tauri-apps/tauri/discussions/2269
  # For details on code signing, see doc/dev/background-information/app/codesigning.md
  if [[ $(uname -s) == "Darwin" ]]; then
    trap 'cleanup_codesigning' ERR INT TERM EXIT

    if [[ ${CI} == "true" ]]; then
      local secrets
      echo "--- :aws: Retrieving signing secrets"
      secrets=$(aws secretsmanager get-secret-value --secret-id sourcegraph/mac-codesigning | jq '.SecretString |  fromjson')
      export APPLE_SIGNING_IDENTITY="$(echo ${secrets} | jq -r '.APPLE_SIGNING_IDENTITY')"
      export APPLE_CERTIFICATE="$(echo ${secrets} | jq -r '.APPLE_CERTIFICATE')"
      export APPLE_CERTIFICATE_PASSWORD="$(echo ${secrets} | jq -r  '.APPLE_CERTIFICATE_PASSWORD')"
      export APPLE_ID="$(echo ${secrets} | jq -r '.APPLE_ID')"
      export APPLE_PASSWORD="$(echo ${secrets} | jq -r '.APPLE_PASSWORD')"
    fi
    # We expect the same APPLE_ env vars that Tauri does here, see https://tauri.app/v1/guides/distribution/sign-macos
    cleanup_codesigning
    security create-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
    security set-keychain-settings my_temporary_keychain.keychain
    security unlock-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
    security list-keychains -d user -s my_temporary_keychain.keychain "$(security list-keychains -d user | sed 's/["]//g')"

    echo "$APPLE_CERTIFICATE" >cert.p12.base64
    base64 -d -i cert.p12.base64 -o cert.p12

    security import ./cert.p12 -k my_temporary_keychain.keychain -P "$APPLE_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
    security set-key-partition-list -S apple-tool:,apple:, -s -k my_temporary_keychain_password -D "$APPLE_SIGNING_IDENTITY" -t private my_temporary_keychain.keychain

    echo "--- :mac::pencil2: binary: ${binary_path}"
    codesign --force -s "$APPLE_SIGNING_IDENTITY" --keychain my_temporary_keychain.keychain --deep "${binary_path}"

    security delete-keychain my_temporary_keychain.keychain
    security list-keychains -d user -s login.keychain
  fi
}

secret_value() {
  # secrets are fetched slightly different depending on host
  local name
  local value
  name=$1
  if [[ $(uname s) == "Darwin" ]]; then
    # host is in aws - probably
    value=$(aws secretsmanager get-secret-value --secret-id ${target} | jq '.SecretString | fromjson')
  else
    # On Linux we assume we're in GCP thus the secret should be injected as an evironment variable. Please check the instance configuration"
    value=""
  fi
  echo ${value}
}

build() {
  if [[ ${CI} == "true" ]]; then
    local secrets
    echo "--- :aws::gcp::tauri: Retrieving tauri signing secrets"
    secrets=$(secret_value "sourcegraph/tauri-key")
    # if the value is not found in secrets we default to an empty string
    export TAURI_PRIVATE_KEY="${TAURI_PRIVATE_KEY:-"$(echo ${secrets} | jq -r '.TAURI_PRIVATE_KEY' | base64 -d || echo '')"}"
    export TAURI_KEY_PASSWORD="${TAURI_KEY_PASSWORD:-"$(echo ${secrets} | jq -r '.TAURI_KEY_PASSWORD' || echo '')"}"
  fi
  echo "--- :tauri: Building Application (${VERSION})"]
  NODE_ENV=production pnpm run build-app-shell
  pnpm tauri build --bundles deb,appimage,app,dmg,updater
}

if [[ ${CI:-""} == "true" ]]; then
  download_artifacts
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
set_version ${VERSION}

# only perform codesigning on mac
if [[ ${CODESIGNING:-"0"} == 1 ]]; then
  # If on a macOS host, Tauri will invoke the base64 command as part of the code signing process.
  # it expects the macOS base64 command, not the gnutils one provided by homebrew, so we prefer
  # that one here:
  export PATH="/usr/bin/:$PATH"

  echo "--- :tauri::mac: Performing code signing"
  binaries=$(find ${BIN_DIR} -type f -name "*apple*")
  # if the paths contain spaces this for loop will fail, but we're pretty sure the binaries in bin don't contain spaces
  for binary in ${binaries}; do
    pre_codesign "${binary}"
  done
fi

echo "xcode DEBUG: $(xcode-select -p)"

build

if [[ ${CI:-""} == "true" ]]; then
  upload_dist
fi
