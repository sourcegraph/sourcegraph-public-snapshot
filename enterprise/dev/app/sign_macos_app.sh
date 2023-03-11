#!/usr/bin/env bash

# INPUT ENVIRONMENT VARIABLES
# - APPLE_DEV_ID_APPLICATION_CERT (optional) - path to the Apple Developer ID Application certificate file
#   - defaults to /mnt/Apple-Developer-ID-Application.p12
#     - the file comes from Secrets in CI
# - APPLE_DEV_ID_APPLICATION_PASSWORD - password for the cert file. Required.
#   - comes from Secrets in CI
# - artifact (optional) - path to app bundle
#   - defaults to ${PWD}/Sourcegraph App.app
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

app_bundle_path="${artifact:-${PWD}/Sourcegraph App.app}"

app_name="$(basename "${app_bundle_path}" .app)"

workdir=$(dirname "${app_bundle_path}")

[ -n "${BUILDKITE-}" ] && {
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  [[ ${app_bundle_path} = /mnt/* ]] || {
    workdir=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
    cp -R "${app_bundle_path}" "${workdir}"
    trap "popd 1>/dev/null && rm -rf \"${workdir}\"" EXIT
  }
}

# sign the app bundle
# going to skip entitlements for now; I don't think we need them

# need to sign the individual binaries individually
# ran into a problem where it failed to sign in place when the permissions on the file were 555
# so open up the permissions and then close them down again
# assume that since these are all executables, they should all end up as 555
files_to_sign=()
while IFS= read -r f; do
  [ "$(file "${workdir}/${app_name}.app/${f}" | grep -c Mach-O)" -gt 0 ] && files_to_sign+=("${f}")
done < <(cd "${workdir}/${app_name}.app" && find . -type f)
for f in "${files_to_sign[@]}"; do
  chmod 777 "${workdir}/${app_name}.app/${f}"
  docker run --rm \
    -v "${application_cert_path}:/certs/cert.p12" \
    -v "${workdir}/${app_name}.app:/sign" \
    -w "/sign" \
    sourcegraph/apple-codesign:0.22.0 \
    sign \
    --p12-file "/certs/cert.p12" \
    --p12-password "${APPLE_DEV_ID_APPLICATION_PASSWORD}" \
    --code-signature-flags runtime \
    "${f}" || exit 1
done

# now sign the whole thing
docker run --rm \
  -v "${application_cert_path}:/certs/cert.p12" \
  -v "${workdir}:/sign" \
  -w "/sign" \
  sourcegraph/apple-codesign:0.22.0 \
  sign \
  --p12-file "/certs/cert.p12" \
  --p12-password "${APPLE_DEV_ID_APPLICATION_PASSWORD}" \
  --code-signature-flags runtime \
  "/sign/${app_name}.app" || exit 1

# close down permissions on the executables after signing
for f in "${files_to_sign[@]}"; do
  chmod 555 "${workdir}/${app_name}.app/${f}"
done

# put that thing back where it came from or so help me!
[ -n "${BUILDKITE-}" ] && {
  [[ ${app_bundle_path} = /mnt/* ]] || {
    # when running in buildkite, and the original bundle was not in a host mount dir,
    # copy the signed app back from the temp dir
    rm -rf "${app_bundle_path}"
    mv "${workdir}/${app_name}.app" "${app_bundle_path}" || exit 1
  }
}

# buildkite support: if an output location is defined, copy the signed app bundle there
[ -n "${signature}" ] && {
  cp -R "${app_bundle_path}" "${signature}" || exit 1
}

exit 0
