#!/usr/bin/env bash

set -eu
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
SCRIPTDIR=$(realpath './dev/codeintel-qa/scripts')

declare -A REVS=(
  # This repository has not been changed
  [zap]='a6015e13fab9b744d96085308ce4e8f11bad1996 2aa9fa25da83bdfff756c36a91442edc9a84576c'

  # Each commit here is tagged as sg-test-1, sg-test-2, and sg-test-3, respectively. See CHANGES.md in the root of the
  # repository's master branch to see a history of changes and which revisions were targeted. We specifically use replace
  # directives in the project root's go.mod file to target sourcegraph-testing/zap, which has no changes of its own. This
  # simulates how common forking works in the Go ecosystem (see our own use of zoekt).
  #
  # To ensure that the last commit in the list for each repository is visible at tip, the master branch's last commit is
  # a merge commit between the true upstream tip and sg-test-3.
  [etcd]='4397ceb9c11be0b3e9ee0111230235c868ba581d bc588b7a2e9af4f903396cdcf66f56190b9e254f ad7848014a051dbe3fcd6a4cff2c7befdd16d5a8'
  [tidb]='8eaaa098b4e938b18485f7b1fa7d8e720b04c699 b5f100a179e20d5539e629bd0919d05774cb7c6a 9aab49176993f9dc0ed2fcb9ef7e5125518e8b98'
  [titan]='fb38de395ba67f49978b218e099de1c45122fb50 415ffd5a3ba7a92a07cd96c7d9f4b734f61248f7 f8307e394c512b4263fc0cd67ccf9fd46f1ad9a5'
)

KEYS=()
VALS=()
IDXS=()
for k in "${!REVS[@]}"; do
  i=0
  for v in ${REVS[$k]}; do
    KEYS+=("${k}")
    VALS+=("${v}")
    IDXS+=("${i}")

    ((i = i + 1))
  done
done

./dev/ci/parallel_run.sh "${SCRIPTDIR}/clone.sh" {} ::: "${!REVS[@]}"
./dev/ci/parallel_run.sh "${SCRIPTDIR}/go-index.sh" {} {} ::: "${KEYS[@]}" :::+ "${IDXS[@]}" :::+ "${VALS[@]}"
