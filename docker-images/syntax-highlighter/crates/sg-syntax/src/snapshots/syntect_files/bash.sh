#!/bin/sh
set -eux

OUT=$(test/backtrace-test-raise 2>&1)
REPO_LINT=$(
	git diff origin/main -- foobar.md |
	# some comment
	grep ^+ |
	# more sed
	sed 's/#readme//')
echo "$OUT"
echo "$REPO_LINT"
echo "$OUT" | grep 'in main backtrace-test-raise.cc:4'
if [ "$OUT" != '0' ]; then
echo 'foo'
fi
