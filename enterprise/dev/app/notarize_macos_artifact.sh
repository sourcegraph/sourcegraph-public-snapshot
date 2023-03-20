#!/usr/bin/env bash

# shellcheck disable=SC2064

# INPUT ENVIRONMENT VARIABLES
# - APPLE_APP_STORE_CONNECT_API_KEY_ID - App Store Connect API Key ID, for authenticating with the notarization service. Required.
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_ISSUER - App Store Connect API Key Issuer, for authenticating with the notarization service.. Required.
#   - comes from Secrets in CI
# - APPLE_APP_STORE_CONNECT_API_KEY_FILE (optional) - path to the App Store Connect API key file, for authenticating with the notarization service.
#   - defaults to /mnt/AuthKey_${APPLE_APP_STORE_CONNECT_API_KEY_ID}.p8
#     - the file comes from Secrets in CI
# - artifact (optional) - path to signed artifact to be notarized
#   - can also be passed on the command line
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

# allow for specifying the location of the artifact via the "artifact" env var
# supports testing outside of CI, also
artifact_path="${artifact}"

# app bundles can be stapled; standalone executables cannot
unset staple

while [ ${#} -gt 0 ]; do
  case "${1}" in
    --staple)
      staple="--staple"
      ;;
    --help)
      echo "$(basename "${BASH_SOURCE[0]}") [--staple] [<file path>]" 1>&2
      exit 1
      ;;
    *)
      # also support passing the artifact path on the command line
      artifact_path="${1}"
      ;;
  esac
  shift
done

[ -n "${artifact_path}" ] || {
  echo "missing artifact path on command line or in \"artifact\" env var" 1>&2
  exit 1
}

[ -f "${artifact_path}" ] || {
  echo "invalid artifact path on command line or in \"artifact\" env var" 1>&2
  exit 1
}

artifact_name="$(basename "${artifact_path}")"

workdir=$(dirname "${artifact_path}")

# Paranoid cleanup of the api key file that may be left in workdir because of Docker bind mounts.
# When in buildkite, the whole workdir will get cleaned up; setting that trap will replace this one
trap "rm -rf \"${workdir}/private_keys\"" EXIT

[ -n "${BUILDKITE-}" ] && {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${artifact_path} = /mnt/* ]] || {
    workdir=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
    cp -R "${artifact_path}" "${workdir}" || exit 1
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
  ${staple} \
  --api-key "${APPLE_APP_STORE_CONNECT_API_KEY_ID}" \
  --api-issuer "${APPLE_APP_STORE_CONNECT_API_KEY_ISSUER}" \
  "/sign/${artifact_name}" || exit 1

# put that thing back where it came from or so help me!
[ -n "${BUILDKITE-}" ] && {
  [[ ${artifact_path} = /mnt/* ]] || {
    # when running in buildkite, and the original bundle was not in a host mount dir,
    # copy the notarized artifact back from the temp dir
    rm -rf "${artifact_path}"
    mv "${workdir}/${artifact_name}" "${artifact_path}" || exit 1
  }
}

# goreleaser support: if an output location is defined, copy the signed artifact there
[ -n "${signature}" ] && {
  cp -R "${artifact_path}" "${signature}" || exit 1
}

exit 0
