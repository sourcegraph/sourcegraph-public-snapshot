#!/usr/bin/env bash

# shellcheck disable=SC2064

# INPUT ENVIRONMENT VARIABLES
# - APPLE_DEV_ID_APPLICATION_CERT (optional) - path to the Apple Developer ID Application certificate file
#   - defaults to /mnt/Apple-Developer-ID-Application.p12
#     - the file comes from Secrets in CI
# - APPLE_DEV_ID_APPLICATION_PASSWORD - password for the cert file. Required.
#   - comes from Secrets in CI
# - artifact (optional) - path to app bundle
# - signature (optional) - path to destination, if different from artifact
#   - no default, not used if not present

# make sure the cert and password is in place
# use the APPLE_DEV_ID_APPLICATION_CERT env var to permit testing outside of CI
# other env variables and file path names come from buildkite, via Google Secrets
application_cert_path=${APPLE_DEV_ID_APPLICATION_CERT:-/mnt/Apple-Developer-ID-Application.p12}
[ -s "${application_cert_path}" ] || {
  echo "missing code signing certificate file" 1>&2
  exit 1
}
[ -n "${APPLE_DEV_ID_APPLICATION_PASSWORD}" ] || {
  echo "missing code signing certificate password" 1>&2
  exit 1
}

# allow for specifying the location of the artifact via the "artifact" env var
# supports testing outside of CI, also
artifact_path="${artifact}"

unset entitlements_path

while [ ${#} -gt 0 ]; do
  case "${1}" in
    --entitlements)
      [ ${#} -ge 2 ] || {
        echo "missing entitlements path" 1>&2
        exit 1
      }
      [ -f "${2}" ] || {
        echo "invalid entitlements path: ${2}" 1>&2
        exit 1
      }
      entitlements_path="${2}"
      shift
      ;;
    --help)
      echo "$(basename "${BASH_SOURCE[0]}") [<file path>]" 1>&2
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

[[ -d "${artifact_path}" || -f "${artifact_path}" ]] || {
  echo "invalid artifact path on command line or in \"artifact\" env var" 1>&2
  exit 1
}

artifact_file="$(basename "${artifact_path}")"

workdir=$(dirname "${artifact_path}")

[ -n "${BUILDKITE-}" ] && {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${app_bundle_path} = /mnt/* ]] || {
    workdir=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
    cp -R "${artifact_path}" "${workdir}"
    trap "popd 1>/dev/null && rm -rf \"${workdir}\"" EXIT
  }
}

entitlements_volume=()
xml_entitlements=()
[ -z "${entitlements_path}" ] || {
  entitlements_volume+=(-v "${entitlements_path}:/entitle/apply.entitlements")
  xml_entitlements+=(--entitlements-xml-path "/entitle/apply.entitlements")
}

docker run --rm "${entitlements_volume[@]}" \
  -v "${application_cert_path}:/sign/cert.p12" \
  -v "${workdir}/${artifact_file}:/sign/${artifact_file}" \
  sourcegraph/apple-codesign:0.22.0 \
  sign "${xml_entitlements[@]}" \
  --p12-file "/sign/cert.p12" \
  --p12-password "${APPLE_DEV_ID_APPLICATION_PASSWORD}" \
  --code-signature-flags runtime \
  "/sign/${artifact_file}" || exit 1

# if not modifying in place, copy to the output location
[ -z "${BUILDKITE-}" ] || {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${app_bundle_path} = /mnt/* ]] || {
    cp -R "${workdir}/${artifact_file}" "${artifact_path}"
  }
}

exit 0
