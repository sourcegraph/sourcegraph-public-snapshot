#!/usr/bin/env bash

# INPUT ENVIRONMENT VARIABLES
# - APPLE_APP_STORE_CONNECT_API_KEY_ID - App Store Connect API Key ID, for authenticating with the notarization service. Required.
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ISSUER - App Store Connect API Key Issuer, for authenticating with the notarization service.. Required.
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_FILE (optional) - path to the App Store Connect API key file, for authenticating with the notarization service.
#   - defaults to /mnt/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8
#     - the file comes from Secrets in CI
# - artifact (optional) - path to signed app bundle
#   - defaults to ${PWD}/Sourcegraph App.app
# - signature (optional) - path to destination, if different from artifact
#   - no default, not used if not present

[ -n "${APPLE_APP_STORE_CONNECT_API_KEY_ID}" ] || {
  echo "missing Apple App Store Connect API Key Id" 1>&2
  exit 1
}

[ -n "${APPLE_APP_STORE_CONNECT_API_KEY_ISSUER}" ] || {
  echo "missing Apple App Store Connect API Key Issuer" 1>&2
  exit 1
}

api_key_path="${APPLE_APP_STORE_CONNECT_API_KEY_FILE:-/mnt/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8}"

[ -f "${api_key_path}" ] || {
  echo "missing Apple App Store Connect API Key file" 1>&2
  exit 1
}

# allow for specifying the location of the app bundle via the "artifact" env var
# supports testing outside of CI, also
app_bundle_path="${artifact:-${PWD}/Sourcegraph App.app}"

app_name="$(basename "${app_bundle_path}" .app)"

workdir=$(dirname "${app_bundle_path}")

# Paranoid cleanup of the api key file that may be left in workdir because of Docker bind mounts.
# When in buildkite, the whole workdir will get cleaned up; setting that trap will replace this one
trap "rm -rf \"${workdir}/private_keys\"" EXIT

[ -n "${BUILDKITE-}" ] && {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${app_bundle_path} = /mnt/* ]] || {
    workdir=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
    cp -R "${app_bundle_path}" "${workdir}" || exit 1
    trap "popd 1>/dev/null && rm -rf \"${workdir}\"" EXIT
  }
}

docker run --rm \
  -v "${workdir}:/sign" \
  -v "${api_key_path}":"/sign/private_keys/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8" \
  -w "/sign" \
  sourcegraph/apple-codesign:0.22.0 \
  notary-submit \
  --wait \
  --staple \
  --api-key "${APPLE_APP_STORE_CONNECT_API_KEY_ID}" \
  --api-issuer "${APPLE_APP_STORE_CONNECT_API_KEY_ISSUER}" \
  "/sign/Sourcegraph App.app" || exit 1

# put that thing back where it came from or so help me!
[ -n "${BUILDKITE-}" ] && {
  [[ ${app_bundle_path} = /mnt/* ]] || {
    # when running in buildkite, and the original bundle was not in a host mount dir,
    # copy the signed app back from the temp dir
    rm -rf "${app_bundle_path}"
    mv "${workdir}/${app_name}.app" "${app_bundle_path}" || exit 1
  }
}

# goreleaser support: if an output location is defined, copy the signed app bundle there
[ -n "${signature}" ] && {
  cp -R "${app_bundle_path}" "${signature}" || exit 1
}

exit 0
