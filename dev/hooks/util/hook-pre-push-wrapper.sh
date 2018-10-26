#!/usr/bin/env bash
#
# A script that receives a Husky git pre-push hook to invoke another script ($check_script) with the
# following environment variables set with information received from the pre-push hook.
#
# Husky stores the stdin and command-line args from the pre-push hook in the $HUSKY_GIT_STDIN and
# $HUSKY_GIT_PARAMS env vars. This script uses those and requires them to be set.
#
# From the git pre-push docs (https://git-scm.com/docs/githooks#_pre_push):
#
# > This hook is called by git-push[1] and can be used to prevent a push from taking place. The hook
# > is called with two parameters which provide the name and location of the destination remote, if
# > a named remote is not being used both values will be the same.
# >
# > Information about what is to be pushed is provided on the hook’s standard input with lines of
# > the form:
# >
# > <local ref> SP <local sha1> SP <remote ref> SP <remote sha1> LF
# >
# > For instance, if the command git push origin master:foreign were run the hook would receive a
# > line like the following:
# >
# > refs/heads/master 67890 refs/heads/foreign 12345
# >
# > although the full, 40-character SHA-1s would be supplied. If the foreign ref does not yet exist
# > the <remote SHA-1> will be 40 0. If a ref is to be deleted, the <local ref> will be supplied as
# > (delete) and the <local SHA-1> will be 40 0. If the local commit was specified by something
# > other than a name which could be expanded (such as HEAD~, or a SHA-1) it will be supplied as it
# > was originally given.
# >
# > If this hook exits with a non-zero status, git push will abort without pushing
# > anything. Information about why the push is rejected may be sent to the user by writing to
# > standard error.

set -euo pipefail

function usage() {
    if [ "$#" -gt 0 ]; then
        echo "Failure reason: $1" 1>&2
    fi

    cat <<'EOF'
Usage: ./hook-pre-push-wrapper.sh $check_script

The following environment variables must be set (https://github.com/typicode/husky/blob/master/DOCS.md#access-git-params-and-stdin):
       $HUSKY_GIT_STDIN should be `$remote $remote_url`
       $HUSKY_GIT_PARAMS should contain lines of the form `$local_ref $local_sha $remote_ref $remote_sha`
$check_script should be the path to a script that uses the following environment variables:
       $remote
       $remote_url
       $local_ref
       $local_sha
       $remote_ref
       $remote_sha
EOF

    cat <<EOF
ARGS: $*
HUSKY_GIT_PARAMS: $HUSKY_GIT_PARAMS
HUSKY_GIT_STDIN:
        $HUSKY_GIT_STDIN
EOF
    exit 1
}

if [ "$#" -ne 1 ]; then
    usage 'Missing $check_script'
fi
check_script="$1"

if [ ! -v HUSKY_GIT_STDIN ] || [ ! -v HUSKY_GIT_PARAMS ]; then
    usage 'Malformed $HUSKY_GIT_PARAMS or $HUSKY_GIT_STDIN'
fi


read -ra git_params <<< "$HUSKY_GIT_PARAMS"
if [ "${#git_params[@]}" -ne 2 ]; then usage 'Malformed $HUSKY_GIT_PARAMS'; fi

remote="${git_params[0]}"
remote_url="${git_params[1]}"
if [ -z "$remote" ] || [ -z "$remote_url" ]; then usage 'Malformed $HUSKY_GIT_PARAMS'; fi

while read -ra line
do
    if [ "${#line[@]}" -eq 0 ]; then
        continue
    fi
    if [ "${#line[@]}" -ne 4 ]; then
        usage 'Malformed $HUSKY_GIT_STDIN';
    fi
    local_ref="${line[0]}"
    local_sha="${line[1]}"
    remote_ref="${line[2]}"
    remote_sha="${line[3]}"
    if [ -z "$local_ref" ] || [ -z "$local_sha" ] || [ -z "$remote_ref" ] || [ -z "$remote_sha" ]; then
        usage 'Malformed $HUSKY_GIT_STDIN'
    fi

    remote="$remote" \
          remote_url="$remote_url" \
          local_ref="$local_ref" \
          local_sha="$local_sha" \
          remote_ref="$remote_ref" \
          remote_sha="$remote_sha" \
          $check_script

done <<< "$HUSKY_GIT_STDIN"
