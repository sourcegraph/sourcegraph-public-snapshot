#!/usr/bin/env bash

set -eu pipefail

TOTAL=$(bazel query --noshow_progress 'tests(//...) except (kind("_diff_test", //...) + tests(//.aspect/...:*) + tests(//doc/...:*))' | wc -l)
MARKED=$(bazel query --noshow_progress 'attr(tags, "owner_.*", tests(//...))' | wc -l)

echo $((MARKED * 100 / TOTAL))
