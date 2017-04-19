#!/bin/bash

if [ -z "$GIT_PARENT_DIRECTORY" ]; then
    echo '$GIT_PARENT_DIRECTORY is not set'
    exit 1
fi

comm -12 <(find "$GIT_PARENT_DIRECTORY" -name 'config' | xargs dirname | sort) <(find "$GIT_PARENT_DIRECTORY" -name 'HEAD' | xargs dirname | sort) > /tmp/git_repos

function normalize() {
    for a in "$@"; do
        x=${a%/.git}
        echo "local/${x#$GIT_PARENT_DIRECTORY/}"
    done
}
export -f normalize

cat /tmp/git_repos | xargs bash -c 'normalize "${@:0}"'
