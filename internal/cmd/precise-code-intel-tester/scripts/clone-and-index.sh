#!/usr/bin/env bash

set -eu
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."
SCRIPTDIR=$(realpath './internal/cmd/precise-code-intel-tester/scripts')

declare -A REVS=(
  # NOTE: Keep these commits in order of their commit date. For the test harness, we will
  # upload sequentially within each repository so that our heurstic that the oldest upload
  # correlates roughly to older comits (which is true in practice) remains true in the test
  # environment.
  [etcd]='fb77f9b1d56391318823c434f586ffe371750321 1044a8b07c56f3d32a1f3fe91c8ec849a8b17b5e dfb0a405096af39e694a501de5b0a46962b3050e'
  [tidb]='2f9a487ebbd2f1a46b5f2c2262ae8f0ef4c4d42f 43764a59b7dcb846dc1e9754e8f125818c69a96f b4f42abc36d893ec3f443af78fc62705a2e54236'
  [titan]='aef232fbec9089d4468ff06705a3a7f84ee50ea6 33623cc32f8d9f999fd69189d29124d4368c20ab 0ad2e75d529bda74472a1dbb5e488ec095b07fe7'
  [zap]='a6015e13fab9b744d96085308ce4e8f11bad1996 2aa9fa25da83bdfff756c36a91442edc9a84576c'
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
