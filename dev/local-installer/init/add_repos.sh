#!/bin/bash
set -x
set -e
echo "this script imports non-GitHub repositories into Sourcegraph"

if [ $# -ne 1 ]; then
    echo $0: usage: ./setup.sh repo-list-path
    exit 1
fi

REPO_LIST=$1

for repo in $(cat $REPO_LIST); do
    set +e
    psql -U sg -c "\
INSERT INTO repo (uri, owner, name, description, vcs, http_clone_url, ssh_clone_url, homepage_url, default_branch, language, blocked, deprecated, fork, mirror, private, created_at, updated_at, pushed_at, vcs_synced_at, indexed_revision, freeze_indexed_revision, origin_repo_id, origin_service, origin_api_base_url) \
SELECT '${repo}', '', '${repo}', '', '', '', '', '', 'master', '', false, false, false, false, false, now(), now(), now(), now(), null, false, '', 0, '' \
WHERE NOT EXISTS ( \
    SELECT id FROM repo WHERE uri='${repo}' \
) \
RETURNING id"
    set -e
done;
