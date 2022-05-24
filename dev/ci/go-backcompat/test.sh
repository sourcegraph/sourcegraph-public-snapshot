#!/usr/bin/env bash

# This script runs the go-build.sh in a clone of the previous minor release as part
# of the continuous backwards compatibility regression tests.

cd "$(dirname "${BASH_SOURCE[0]}")/../../../"
set -eu

MIGRATION_STAGING=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$MIGRATION_STAGING"
}
trap cleanup EXIT

# `disable_test ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in the given Go test file. This will return 1 if there was a matching test and
# return 0 otherwise.
function disable_test() {
  sed -i_bak "s/func ${2}/func _${2}/g" "${1}"

  local ret=1
  if diff "${1}" "${1}_bak" >/dev/null; then
    ret=0 # no diff
  fi

  rm "${1}_bak"
  return ${ret}
}

# `disable_test_file ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in the given Go test file. If there is no matching test, an unknown test message is
# displayed and the script is halted with exit code 1.
function disable_test_file() {
  if disable_test "${1}" "${2}"; then
    echo "Unknown test in ${1}: ${2}"
    exit 1
  fi
}

# `disable_test_dir ${path} ${prefix}` rewrites `func ${prefix}` to `func _${prefix}`
# in all Go test files under the given path.
function disable_test_dir() {
  local num_changed=0

  while read -r path; do
    if ! disable_test "${path}" "${2}"; then
      num_changed=$((num_changed + 1))
    fi
  done < <(find "${1}" -name '*_test.go' -type f)

  if [ ${num_changed} -eq 0 ]; then
    echo "Unknown test in ${1}: ${2}"
  fi
}

# `disable_test_path ${path} ${prefix}`
function disable_test_path() {
  echo "Disabling test '${2}*' in ${1}"

  if [ -d "${1}" ]; then
    disable_test_dir "${1}" "${2}"
  elif [ -f "${1}" ]; then
    disable_test_file "${1}" "${2}"
  fi
}

current_head=$(git rev-parse HEAD)
latest_minor_release_tag="v${MINIMUM_UPGRADEABLE_VERSION}"
flakefile="./dev/ci/go-backcompat/flakefiles/${latest_minor_release_tag}.json"

# Early exit
if git diff --quiet "${latest_minor_release_tag}".."${current_head}" migrations; then
  echo "--- No schema changes"
  echo "No schema changes since last minor release"
  exit 0
fi

echo "--- Running backwards compatibility tests"
echo "current_head                = ${current_head}"
echo "latest_minor_release_tag    = ${latest_minor_release_tag}"
echo ""
echo "Running Go tests to test database schema backwards compatibility:"
echo "- database schemas are defined at HEAD / ${current_head}, and"
echo "- unit tests are defined at the last minor release ${latest_minor_release_tag}."
echo ""
echo "New failures of these tests indicate that new changes to the"
echo "database schema do not allow for continued proper operation of"
echo "Sourcegraph instances deployed at the previous release."
echo ""

PROTECTED_FILES=(
  ./dev/ci/go-test.sh
  ./dev/ci/go-backcompat
  ./dev/ci/asdf-install.sh
)

# Rewrite the current migrations into a temporary folder that we can force
# apply over old code.
go run ./dev/ci/go-backcompat/reorganize.go "${MIGRATION_STAGING}"

# Check out the previous code then immediately restore whatever
# the current version of the protected files are.
git checkout "${latest_minor_release_tag}"
git checkout "${current_head}" -- "${PROTECTED_FILES[@]}"

# Remove the languages submodules, because they mess these tests up
rm -rf ./docker-images/syntax-highlighter/crates/sg-syntax/languages/

for schema in frontend codeintel codeinsights; do
  # Force apply newer schema definitions
  rm -rf "./migrations/${schema}"
  mv "${MIGRATION_STAGING}/${schema}" "./migrations/${schema}"
done

# If migration files have been renamed or deleted between these commits
# (which historically we've done in response to reverted migrations), we
# might end up with a combination of files from both commits that ruin
# some of the assumptions we make (unique prefix ID being one major one).
# We delete this directory first prior to the checkout so that we don't
# have any current state in the migrations directory to mess us up in this
# way.

if [ -f "${flakefile}" ]; then
  echo ""
  echo "Disabling tests listed in flakefile ${flakefile}"

  for pair in $(jq -r '.[] | "\(.path):\(.prefix)"' <"${flakefile}"); do
    IFS=' ' read -ra parts <<<"${pair/:/ }"
    disable_test_path "${parts[0]}" "${parts[1]}"
  done
fi

# Re-run asdf to ensure we have the correct set of utilities to
# run the currently checked out version of the Go unit tests.
echo "--- asdf install checked out tools"
./dev/ci/asdf-install.sh
go version

echo "--- run tests"
if ! ./dev/ci/go-test.sh "$@"; then
  annotation=$(
    cat <<EOF
This commit contains database schema definitions that caused an unexpected
failure of one or more unit tests at tagged commit \`${latest_minor_release_tag}\`.
Rewrite these schema changes to be backwards compatible. For help,
see [the migrations guide](https://docs.sourcegraph.com/dev/background-information/sql/migrations).

If this backwards incompatibility is intentional or if the test is flaky,
an exception for this test can be added to the following flakefile:

\`\`\`
${flakefile}
\`\`\`

EOF
  )
  mkdir -p ./annotations/
  echo "$annotation" | tee './annotations/go-backcompat.md'
  exit 1
fi
