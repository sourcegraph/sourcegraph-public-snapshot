#!/usr/bin/env /bin/sh
# This script runs several git commands with the goal to optimize the
# performance of git for large repositories.  With the exception of "repack" and
# "commit-graph write", the commands and their order follow the default
# behavior of "git gc".

set -xe
git pack-refs --all --prune
git reflog expire --all

# With multi-pack-index and geometric packing, repacking should take time
# proportional to the number of new objects.
git repack --write-midx --write-bitmap-index -d --geometric=2
git prune --expire 2.weeks.ago

# --changed-paths provides significant performance gains for getting the history
# of a directory or a file with git log -- <path>
git commit-graph write --reachable --changed-paths
