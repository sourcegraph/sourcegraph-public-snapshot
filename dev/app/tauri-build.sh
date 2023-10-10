#!/usr/bin/env bash
set -eu

cd "$(dirname "${BASH_SOURCE[0]}")"/../.. || exit 1

BIN_DIR=".bin"
DIST_DIR="dist"

download_artifacts() {
  mkdir -p .bin
  buildkite-agent artifact download "${BIN_DIR}/sourcegraph-backend-*" .bin/
  chmod -R +x .bin/
}

set_version() {
  local version
  local tauri_conf
  local tmp
  version=$1
  tauri_conf="./src-tauri/tauri.conf.json"
  tmp=$(mktemp)
  echo "--- Updating package version in '${tauri_conf}' to ${version}"
  jq --arg version "${version}" '.package.version = $version' "${tauri_conf}" >"${tmp}"
  mv "${tmp}" ./src-tauri/tauri.conf.json
}

bundle_path() {
  local platform
  platform="$(./dev/app/detect-platform.sh)"
  echo "./src-tauri/target/${platform}/release/bundle"
}

upload_dist() {
  local path
  local target_dir
  path="$(bundle_path)"
  echo "searching for artefacts in '${path}' and moving them to dist/"
  src=$(find "${path}" -type f \( -name "Cody*.dmg" -o -name "Cody*.tar.gz" -o -name "cody*.deb" -o -name "cody*.AppImage" -o -name "cody*.tar.gz" -o -name "*.sig" \))
  target_dir="./${DIST_DIR}"

  mkdir -p "${target_dir}"
  for from in ${src}; do
    mv -vf "${from}" "${target_dir}/"
  done

  echo --- Uploading artifacts from dist
  ls -al "./${DIST_DIR}"
  buildkite-agent artifact upload "${DIST_DIR}/*"
}

create_app_archive() {
  local version
  local platform
  local path
  local app_path
  local arch
  local target

  version=$1
  platform=$2
  path="$(bundle_path)"
  app_path=$(find "${path}" -type d -name "Cody.app")
  app_tar_gz=$(find "${path}" -type f -name "Cody.app.tar.gz")

  # we extract the arch from the platform
  arch=$(echo "${platform}" | cut -d '-' -f1)
  if [[ -z ${arch} ]]; then
    arch=$(uname -m)
  fi

  target="Cody.${version}.${arch}.app.tar.gz"
  # # we have to handle Cody.app differently since it is a dir
  if [[ -d ${app_path} && -z ${app_tar_gz} ]]; then
    pushd .
    cd "${path}/macos/"
    echo "--- :file_cabinet: Creating archive ${target}"
    tar -czvf "${target}" "Cody.app"
    popd
  elif [[ -e ${app_tar_gz} ]]; then
    echo "--- :file_cabinet: Moving existing archive/signatures to ${target}"
    mv -vf "${app_tar_gz}" "$(dirname "${app_tar_gz}")/${target}"
    mv -vf "${app_tar_gz}.sig" "$(dirname "${app_tar_gz}")/${target}.sig" || echo "--- signature not found - skipping"
  fi
}

cleanup_codesigning() {
  # shellcheck disable=SC2143
  if [[ $(security list-keychains -d user | grep -q "my_temporary_keychain") ]]; then
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
  trap 'cleanup_codesigning' ERR INT TERM EXIT

  if [[ ${CI} == "true" ]]; then
    local secrets
    echo "--- :aws: Retrieving signing secrets"
    secrets=$(aws secretsmanager get-secret-value --secret-id sourcegraph/mac-codesigning | jq '.SecretString |  fromjson')
    APPLE_SIGNING_IDENTITY="$(echo "${secrets}" | jq -r '.APPLE_SIGNING_IDENTITY')"
    APPLE_CERTIFICATE="$(echo "${secrets}" | jq -r '.APPLE_CERTIFICATE')"
    APPLE_CERTIFICATE_PASSWORD="$(echo "${secrets}" | jq -r '.APPLE_CERTIFICATE_PASSWORD')"
    APPLE_ID="$(echo "${secrets}" | jq -r '.APPLE_ID')"
    APPLE_PASSWORD="$(echo "${secrets}" | jq -r '.APPLE_PASSWORD')"

    export APPLE_SIGNING_IDENTITY
    export APPLE_CERTIFICATE
    export APPLE_CERTIFICATE_PASSWORD
    export APPLE_ID
    export APPLE_PASSWORD
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

  echo "--- :mac::spiral_note_pad::lower_left_fountain_pen: binary: ${binary_path}"
  codesign --force -s "$APPLE_SIGNING_IDENTITY" --keychain my_temporary_keychain.keychain --deep "${binary_path}"

  security delete-keychain my_temporary_keychain.keychain
  security list-keychains -d user -s login.keychain
}

secret_value() {
  local name
  local value
  name=$1
  if [[ $(uname -s) == "Darwin" ]]; then
    # host is in aws - probably
    value=$(aws secretsmanager get-secret-value --secret-id "${name}" | jq '.SecretString | fromjson')
  else
    # On Linux we assume we're in GCP thus the secret should be injected as an evironment variable. Please check the instance configuration
    value=""
  fi
  echo "${value}"
}

build() {
  echo --- :magnify_glass: detecting platform
  local version
  local do_updater_bundle
  local platform
  local bundles

  platform="$1"
  version="$2"
  do_updater_bundle="$3"

  # we only allow the updater build once we're in CI or see "SRC_APP_UPDATER_BUILD=1"
  bundles="deb,appimage,app,dmg"

  echo "platform is: ${platform}"

  if [[ ${CI:-""} == "true" ]]; then
    local secrets
    echo "--- :aws::gcp::tauri: Retrieving tauri signing secrets"
    secrets=$(secret_value "sourcegraph/tauri-key")
    # if the value is not found in secrets we default to an empty string
    export TAURI_PRIVATE_KEY="${TAURI_PRIVATE_KEY:-"$(echo "${secrets}" | jq -r '.TAURI_PRIVATE_KEY' | base64 -d || echo '')"}"
    export TAURI_KEY_PASSWORD="${TAURI_KEY_PASSWORD:-"$(echo "${secrets}" | jq -r '.TAURI_KEY_PASSWORD' || echo '')"}"
  fi

  if [[ ${do_updater_bundle} == 1 ]]; then
    bundles="$bundles,updater"
  fi

  echo "--- :tauri: Building Application (${version}) with bundles '${bundles}' for platform: ${platform}"
  NODE_ENV=production pnpm run build-app-shell
  pnpm tauri build --bundles ${bundles} --target "${platform}"
}

if [[ ${CI:-""} == "true" ]]; then
  download_artifacts
fi

VERSION=$(./dev/app/app-version.sh)
set_version "${VERSION}"
PLATFORM="$(./dev/app/detect-platform.sh)"

if [[ ${CODESIGNING:-"0"} == 1 && $(uname -s) == "Darwin" ]]; then
  # We want any xcode related tools to be picked up first so inject it here in the path
  xcode_path="$(xcode-select -p)/usr/bin"
  export PATH="${xcode_path}:$PATH"
  # If on a macOS host, Tauri will invoke the base64 command as part of the code signing process.
  # it expects the macOS base64 command, not the gnutils one provided by homebrew, so we prefer
  # that one here:
  export PATH="/usr/bin/:$PATH"

  echo "--- :tauri::mac: Performing code signing"
  binaries=$(find ${BIN_DIR} -type f -name "*${PLATFORM}*")
  # if the paths contain spaces this for loop will fail, but we're pretty sure the binaries in bin don't contain spaces
  for binary in ${binaries}; do
    pre_codesign "${binary}"
  done
fi

CI="${CI:-"false"}"
# note that this script respects the OVERRIDE_PLATFORM env variable
SRC_APP_UPDATER_BUILD="${SRC_APP_UPDATER_BUILD:-"0"}"
build "${PLATFORM}" "${VERSION}" "${SRC_APP_UPDATER_BUILD:-"0"}"

create_app_archive "${VERSION}" "${PLATFORM}"

if [[ ${CI:-""} == "true" ]]; then
  upload_dist
fi
