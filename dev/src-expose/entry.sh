#!/usr/bin/env sh

# Feed every directory in /app/data to src-expose

codedirs=$(cd /app/data && find . -maxdepth 1 -mindepth 1 | cut -c3- | xargs)

# shellcheck disable=SC2086
exec /usr/local/bin/src-expose $codedirs
