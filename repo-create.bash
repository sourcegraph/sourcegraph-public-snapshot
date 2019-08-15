#!/bin/bash

set -eux
unset CDPATH

export HUB_CONFIG=$HOME/.config/hub-sd9
export HUB_PROTOCOL=https
export GITHUB_TOKEN=$(grep oauth_token "$HUB_CONFIG" | sed 's/\s*oauth_token: //')

donefile=$HOME/tmp/lodash-repos.done.txt
skipfile=$HOME/tmp/lodash-repos.skip.txt

upstream_repo_full_name="$1"
upstream_repo_name_with_owner=${upstream_repo_full_name/github.com\//}
repo_name=${upstream_repo_full_name/github.com\/[^\/]*\//}
fork_repo_name_with_owner="sd9/${repo_name}"

grep "$fork_repo_name_with_owner" "$donefile" && exit
grep "$fork_repo_name_with_owner" "$skipfile" && exit

upstream_repo_size=$(curl -sS https://api.github.com/repos/${upstream_repo_name_with_owner} | jq '.size')
MAX_SIZE_KB=10000
[ "$upstream_repo_size" -gt $MAX_SIZE_KB ] && echo "github.com/${fork_repo_name_with_owner}" >> "$skipfile" && exit

tmpdir=$(mktemp -d)
clonedir="${tmpdir}/${repo_name}"
cd $tmpdir
git clone --bare --single-branch "https://${upstream_repo_full_name}.git" "$clonedir"
cd $clonedir
git remote rm origin
curl -sS -XPOST -u "x-oauth-token:$GITHUB_TOKEN" -H 'Content-Type: application/json; charset=utf-8' -d '{"name":"'${repo_name}'", "private": true}' https://api.github.com/user/repos
git remote add origin "https://x-oauth-token:${GITHUB_TOKEN}@github.com/${fork_repo_name_with_owner}.git"
git push origin
rm -rf "$tmpdir"
echo "github.com/${fork_repo_name_with_owner}" >> "$donefile"
