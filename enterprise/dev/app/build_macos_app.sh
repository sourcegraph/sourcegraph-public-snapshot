#!/usr/bin/env bash

# shellcheck disable=SC2064

# INPUT ENVIRONMENT VARIABLES
# - VERSION - required in order to find the binary on GCS and is stored in the app for use when detecting new versions
#   - defaults to 0.0.0
# - artifact (optional) - path to binary file
#   - if not supplied, downloads from GCS to ${PWD}/sourcegraph
# - signature (optional) - path to destination app bundle
#   - defaults to ${PWD}/${app_name}
# - app_name (optional) - the name of the app bundle
#   - defaults to "Sourcegraph App.app"

# VERSION should come from the environment
VERSION=${VERSION:-0.0.0}

# index off of the directory of this shell script to find other resources
exedir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

# grab the working directory as a reference
workdir="${PWD}"

# set up a temp dir to work in

if [ -n "${BUILDKITE-}" ]; then
  # In Buildkite, we're running in a Docker container, so `docker run -v` needs to refer to a
  # directory on our Docker host, not in our container. Use the /mnt/tmp directory, which is shared
  # between `dind` (the Docker-in-Docker host) and our container.
  tempdir=$(mktemp -d --tmpdir=/mnt/tmp -t sourcegraph.XXXXXXXX)
else
  tempdir=$(mktemp -d 2>/dev/null || mktemp -d -t sourcegraph.XXXXXXXX 2>/dev/null)
fi

# do all work in the temp dir
pushd "${tempdir}" 1>/dev/null || exit 1

#make sure to exit the temp dir and unlink it when done
trap "popd 1>/dev/null && rm -rf \"${tempdir}\"" EXIT

# preserve the ability to run as part of the goreleaser process
# goreleaser puts the path to the file in the "artifact" env var
# shellcheck disable=SC2154
binary_file_path=${artifact}

# grab the binary file from GCS if not running in goreleaser
[ -n "${binary_file_path}" ] || {
  gsutil cp "gs://sourcegraph-app-releases/${VERSION}/sourcegraph_${VERSION}_darwin_all.zip" . || exit 1
  unzip "sourcegraph_${VERSION}_darwin_all.zip" || exit 1
  binary_file_path="${PWD}/sourcegraph"
}

[ -f "${binary_file_path}" ] || {
  echo "missing the binary file to put in the app" 1>&2
  exit 1
}

app_name="$(basename "${app_name:-Sourcegraph App}" .app).app"

# grab the app bundle template
# which contains the binary resources:
# - executable wrapper made by Platypus (for now)
# - src-cli
# - universal-ctags
# - git
# - icons
# include the ability to use a template without downloading it all the time for testing
# shellcheck disable=SC2154
if [ -n "${app_template_path}" ]; then
  cp -R "${app_template_path}" "${app_name}" || exit 1
else
  template_file="${app_name}-template.tar.gz"
  template_version=$(gsutil cat "gs://sourcegraph_app_macos_dependencies/template-version.txt" 2>/dev/null)
  [ -z "${template_version}" ] || template_file="${app_name}-template-${template_version}.tar.gz"
  gsutil cp "gs://sourcegraph_app_macos_dependencies/${template_file}" . || exit 1
  tar -xzf "${template_file}" || exit 1
fi

# copy in the launcher shell script
# the destination name needs to match what the app bundle executable expects
# which used to be controlled by Platypus, but is now a Xcode project
cp "${exedir}/macos_app/app_bundle/sourcegraph_launcher.sh" "${app_name}/Contents/Resources/sourcegraph_launcher.sh" || exit 1
chmod 555 "${app_name}/Contents/Resources/sourcegraph_launcher.sh" || exit 1

# copy in the sourcegraph binary
# the destination name is specified by the launcher shell script
cp "${binary_file_path}" "${app_name}/Contents/Resources/sourcegraph" || exit 1
chmod 555 "${app_name}/Contents/Resources/sourcegraph" || exit 1

# record the version where it can be used for checking for new versions
# would probably be better to store it in Info.plist, but that might need `defaults`, requiring macOS
printf '%s' "${VERSION}" >"${app_name}/Contents/Resources/version.txt"

# put the app in a place where it can be picked up
# preserve the ability to run as part of the goreleaser process
# by using the "signature" name template
destination_file_path="${signature:-${workdir}/${app_name}}"

[ -d "${destination_file_path}" ] && rm -rf "${destination_file_path}"
mv "${app_name}" "${destination_file_path}" || exit 1

exit 0
