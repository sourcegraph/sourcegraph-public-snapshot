#!/usr/bin/env bash
#
# Checks that no enterprise/ directory exists at repository root if the remote URL contains
# "sourcegraph/sourcegraph".

# Constants
DISALLOW_PATH=":/enterprise"
CHECK_REMOTE_SUBSTR="sourcegraph/sourcegraph"

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

function usage() {
    cat <<'EOF'
Usage: ./check-no-enterprise.sh

The following environment variables must be set:
    $remote_url
    $local_sha
    $local_ref
EOF
}

function failCheck() {
    cat <<EOF
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!! DANGER DANGER DANGER DANGER DANGER DANGER DANGER DANGER DANGER DANGER    !!
!!  ACHTUNG ACHTUNG ACHTUNG ACHTUNG ACHTUNG ACHTUNG ACHTUNG ACHTUNG ACHTUNG !!
!! PELIGRO PELIGRO PELIGRO PELIGRO PELIGRO PELIGRO PELIGRO PELIGRO PELIGRO  !!
!!   危险 危险 危险 危险 危险 危险 危险 危险 危险 危险 危险 危险 危险 危险  !!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!
L!=> DANGER: There is a path "${DISALLOW_PATH}" you are attempting to push
             to the remote "${remote_url}" or committing to
             a branch tracking that remote.

     Either change the upstream remote or modify the commits to remove the
     "${DISALLOW_PATH}" path.

     The output of the following commands was non-empty:

        git log ${local_sha} -- ${DISALLOW_PATH}
        git log ${local_ref} -- ${DISALLOW_PATH}
!                                                                            !
!!                                                                          !!
!!!!                                                                      !!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
EOF
    exit 1
}

case "$remote_url" in
    *"$CHECK_REMOTE_SUBSTR"*)
        # If pushing to *"$CHECK_REMOTE_SUBSTR"*, continue with check
        ;;
    *)
        # If not pushing there, don't check
        exit 0
esac

if [ "$local_sha" = "0000000000000000000000000000000000000000" ]; then
    exit 0
fi

if [ ! -z "$(git log ${local_ref} -- ${DISALLOW_PATH})" ]; then
    failCheck
fi

if [ ! -z "$(git log ${local_sha} -- ${DISALLOW_PATH})" ]; then
    failCheck
fi
