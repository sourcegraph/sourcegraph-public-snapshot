#!/bin/sh

OUT=$(test/backtrace-test-raise 2>&1)
REPO_LINT=$(
	git diff origin/main -- foobar.md |
	# some comment
	grep ^+ |
	# more sed
	sed 's/#readme//')
echo "$OUT"
echo "$OUT" | grep 'in main backtrace-test-raise.cc:4'
if [[ x != '0' ]]; then
echo 'foo'
fi
