#!/usr/bin/env bash

# shellcheck disable=SC2064

# designed to be run as a post-action on an Archive action
# checks to make sure the app is universal, downloads the dependencies,
# creates the app template with those dependencies,
# and finally uploads a tar gzip archive to the "sourcegraph_app_macos_dependencies" GCS bucket

filename="$(basename "${BASH_SOURCE[0]}")"

log() {
  # https://developer.apple.com/documentation/xcode/running-custom-scripts-during-a-build#Log-errors-and-warnings-from-your-script
  # [filename]:[linenumber]: error | warning | note : [message]
  # `log error|warning|note <message>`
  printf '%s: %s: %s - %s\n' "${filename}" "${1}" "$(date +'%Y-%m-%d %H:%M:%S')" "${2}"
}
debug() {
  log "note" "${1}"
}
info() {
  log "note" "${1}"
}
warn() {
  log "warning" "${1}"
}
error() {
  log "error" "${1}"
}

git_version=2.39.2
src_version=4.5.0
ctags_version=6.0.0

info "starting app bundle template build"

app_name="Sourcegraph App.app"
app_bundle_path="${ARCHIVE_PRODUCTS_PATH}/Applications/${app_name}"

[[ -d "${app_bundle_path}" ]] || {
    error "missing app bundle"
    exit 1
}

[[ $(file "${app_bundle_path}/Contents/MacOS/Sourcegraph App" | grep -c Mach-O) -lt 3 ]] && {
    error "App is not universal"
    exit 1
}

tempdir=$(mktemp -d || mktemp -d -t temp_XXXXXXXX)
pushd "${tempdir}" || {
    error "unable to work in temp dir"
    exit 1
}

trap "popd; rm -rf \"${tempdir}\"" EXIT

cp -R "${app_bundle_path}" . || {
    error "unable to copy app template to temp dir"
    exit 1
}

[ -d "${PWD}/git-${git_version}/bin" ] || {
  curl -fsSLO "https://storage.googleapis.com/sourcegraph_app_macos_dependencies/git-universal-${git_version}.tar.gz"
  tar -xvzf "git-universal-${git_version}.tar.gz"
}
rm -rf "${app_name}/Contents/Resources/git"
cp -R git-${git_version} "${app_name}/Contents/Resources/git" || {
    error "unable to add git to the app bundle"
    exit 1
}

[ -f "${PWD}/src" ] || {
  curl -fsSLO "https://storage.googleapis.com/sourcegraph_app_macos_dependencies/src-universal-${src_version}.tar.gz"
  tar -xvzf "src-universal-${src_version}.tar.gz"
}
rm -rf "${app_name}/Contents/Resources/src"
cp src "${app_name}/Contents/Resources/src" || {
    error "unable to add src-cli to the app bundle"
    exit 1
}


[ -f "${PWD}/universal-ctags" ] || {
  curl -fsSLO "https://storage.googleapis.com/sourcegraph_app_macos_dependencies/universal-ctags-universal-${ctags_version}.tar.gz"
  tar -xvzf "universal-ctags-universal-${ctags_version}.tar.gz"
}
rm -rf "${app_name}/Contents/Resources/universal-ctags"
cp universal-ctags "${app_name}/Contents/Resources/universal-ctags" || {
    error "unable to add universal-ctags to the app bundle"
    exit 1
}

tar -cvzf "${app_name}-template.tar.gz" "${app_name}" || {
    error "unable to archive the app bundle"
    exit 1
}

# version the app bundle template
unset app_bundle_template_version
# extract the date and time from the ARCHIVE_PRODUCTS_PATH environment variable
pattern='^.*/Archives/([0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9])/.* ([0-9][0-9][.][0-9][0-9])[.]xcarchive/Products$'
[[ ${ARCHIVE_PRODUCTS_PATH} =~ ${pattern} ]] && {
    app_bundle_template_version="${BASH_REMATCH[1]}_${BASH_REMATCH[2]}"
}
app_bundle_archive_versioned=${app_name}-template.tar.gz

[ -z "${app_bundle_template_version}" ] || app_bundle_archive_versioned=${app_name}-template-${app_bundle_template_version}.tar.gz

"${HOME}/google-cloud-sdk/bin/gsutil" cp "${app_name}-template.tar.gz" "gs://sourcegraph_app_macos_dependencies/${app_bundle_archive_versioned}" || {
    cp "${app_name}-template.tar.gz" "${HOME}/Downloads/${app_bundle_archive_versioned}"
    error "unable to upload the app bundle template to GCS"
    [ -f "${HOME}/Downloads/${app_bundle_archive_versioned}" ] && error "the app bundle template has been placed in \"${HOME}/Downloads/${app_bundle_archive_versioned}\" for you to upload manually"
    exit 1
}
[ -z "${app_bundle_template_version}" ] || {
    printf '%s' "${app_bundle_template_version}" > template-version.txt
    "${HOME}/google-cloud-sdk/bin/gsutil" cp template-version.txt "gs://sourcegraph_app_macos_dependencies/template-version.txt"
}

info "uploaded the app bundle template to gs://sourcegraph_app_macos_dependencies/${app_bundle_archive_versioned}"
