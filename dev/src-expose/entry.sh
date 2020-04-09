#!/usr/bin/env sh

# Feed every directory in /app/data to src-expose
codedirs=$(cd /app/data && ls -d * | xargs)
exec /usr/local/bin/src-expose $codedirs
