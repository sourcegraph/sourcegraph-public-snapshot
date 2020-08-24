#!/usr/bin/env bash

set -euxo pipefail

cd /home/sourcegraph

TEMP_DIR=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

GIT_CLONE_DIR="${TEMP_DIR}/aports"
# For git 2.28, see https://gitlab.alpinelinux.org/alpine/aports/-/commit/bc9abfdcd429f8fa22575a0826625a2427e8b109
GIT_REVISION=${GIT_REVISION:-"bc9abfdcd429f8fa22575a0826625a2427e8b109"}

git config --global user.name "Sourcegraph"
git config --global user.email "support@sourcegraph.com"
git clone git://git.alpinelinux.org/aports "${GIT_CLONE_DIR}"
cd "${GIT_CLONE_DIR}"
git reset --hard "${GIT_REVISION}"

TARGET_DIR="/target"
BUILD_OUTPUT_DIR="${TEMP_DIR}/GIT_OUTPUT"
DIST_FILES_DIR="/var/cache/distfiles"

sudo mkdir -p "${DIST_FILES_DIR}"
sudo chmod a+w "${DIST_FILES_DIR}"
sudo echo -ne '\n' | abuild-keygen -a -i
sudo mkdir "${BUILD_OUTPUT_DIR}"
sudo chmod 0777 "${BUILD_OUTPUT_DIR}"
sudo mkdir "${TARGET_DIR}"
sudo chmod 0777 "${TARGET_DIR}"

cd main/git

echo "--- building git binaries to output directory: ${OUTPUT_DIR}"
abuild -c -r -P "${OUTPUT_DIR}"
mv "${BUILD_OUTPUT_DIR}/main/x86_64/" "${TARGET_DIR}"

exa --long --tree "${TARGET_DIR}"
