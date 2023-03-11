#!/usr/bin/env bash

### code sign the macos binary

# INPUT ENVIRONMENT VARIABLES
# - VERSION - required in order to find the binary on GCS; otherwsise optional
#   - defaults to 0.0.0
# - APPLE_DEV_ID_APPLICATION_CERT (optional) - path to the Apple Developer ID Application certificate file
#   - defaults to /mnt/Apple-Developer-ID-Application.p12
#     - the file comes from Secrets in CI
# - APPLE_DEV_ID_APPLICATION_PASSWORD - password for the cert file. Required.
#   - comes from Secrets in CI
# - artifact (optional) - path to binary file
#   - if not supplied, downloads from GCS to ${PWD}/sourcegraph
# - signature (optional) - path to destination, if different from artifact
#   - no default, not used if not present

# make sure the cert and password is in place
# use the APPLE_DEV_ID_APPLICATION_CERT env var to permit testing outside of CI
application_cert_path=${APPLE_DEV_ID_APPLICATION_CERT:-/mnt/Apple-Developer-ID-Application.p12}
[ -s "${application_cert_path}" ] || {
  echo "missing code signing certificate file" 1>&2
  exit 1
}
[ -n "${APPLE_DEV_ID_APPLICATION_PASSWORD}" ] || {
  echo "missing code signing certificate password" 1>&2
  exit 1
}

VERSION=${VERSION:-0.0.0}

# preserve the ability to run as part of the goreleaser process
# goreleaser puts the path to the file in the "artifact" env var
binary_file_path=${artifact}

# but allow grabbing the binary file from GCS
[ -n "${binary_file_path}" ] || {
  gsutil cp "gs://sourcegraph-app-releases/${VERSION}/sourcegraph_${VERSION}_darwin_all.zip" . || exit 1
  unzip -o "sourcegraph_${VERSION}_darwin_all.zip" || exit 1
  binary_file_path="${PWD}/sourcegraph"
}

[ -f "${binary_file_path}" ] || {
  echo "no binary file to sign" 1>&2
  exit 1
}

binary_file_name=$(basename "${binary_file_path}")

# write the signed binary to the location specified by goreleaser, if it's specified
# otherwise, back to where we got it from
output_file_path="${signature:-${binary_file_path}}"

[ -n "${BUILDKITE-}" ] && {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${binary_file_path} = /mnt/* ]] || {
    TMPDIR=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
    cp "${binary_file_path}" "${TMPDIR}/${binary_file_name}" || exit 1
    binary_file_path="${TMPDIR}/${binary_file_name}"
    trap '[ -d "${TMPDIR}" ] && rm -rf "${TMPDIR}"' EXIT
  }
}

docker run --rm \
  -v "${application_cert_path}:/sign/cert.p12" \
  -v "${binary_file_path}:/sign/${binary_file_name}" \
  sourcegraph/apple-codesign:0.22.0 \
  sign \
  --p12-file "/sign/cert.p12" \
  --p12-password "${APPLE_DEV_ID_APPLICATION_PASSWORD}" \
  --code-signature-flags runtime \
  "/sign/${binary_file_name}" || exit 1

# if not modifying in place, copy to the output location
[[ "${binary_file_path}" = "${output_file_path}" ]] || {
  cp "${binary_file_path}" "${output_file_path}" || exit 1
}

# if we got the file from GCS, put it back
[ -n "${artifact}" ] || {
  zip -u9 "sourcegraph_${VERSION}_darwin_all.zip" sourcegraph || exit 1
  gsutil cp "sourcegraph_${VERSION}_darwin_all.zip" "gs://sourcegraph-app-releases/${VERSION}/sourcegraph_${VERSION}_darwin_all.zip" || exit 1
}

exit 0
