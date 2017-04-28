#!/bin/bash

if [ -z "$GIT_PARENT_DIRECTORY" ]; then
    echo '$GIT_PARENT_DIRECTORY is not set'
    exit 1
fi

comm -12 <(find "$GIT_PARENT_DIRECTORY" -name 'config' -not -path '*/.data/*' | xargs dirname | sort) <(find "$GIT_PARENT_DIRECTORY" -name 'HEAD' -not -path '*/.data/*' | xargs dirname | sort) > /tmp/git_repos

function normalize() {
    for a in "$@"; do
        x=${a%/.git}
        echo "local/${x#$GIT_PARENT_DIRECTORY/}"
    done
}

for r in $(cat /tmp/git_repos); do
    normalize "$r";
done
