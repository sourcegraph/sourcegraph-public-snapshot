#!/usr/bin/env bash

set -eu pipefail

echo "Targets without a denoted owner:" >&2
bazel query --noshow_progress 'tests(//...) except (kind("_diff_test", //...) + tests(//.aspect/...:*) + tests(//doc/...:*)) except attr(tags, "owner_.*", tests(//...))' >&2

TOTAL=$(bazel query --noshow_progress 'tests(//...) except (kind("_diff_test", //...) + tests(//.aspect/...:*) + tests(//doc/...:*))' | wc -l)
MARKED=$(bazel query --noshow_progress 'attr(tags, "owner_.*", tests(//...))' | wc -l)

echo $((MARKED * 100 / TOTAL))
