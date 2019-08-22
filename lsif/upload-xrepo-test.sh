#!/bin/bash -e

urlencode() {
    echo "$1" | curl -Gso /dev/null -w %{url_effective} --data-urlencode @- "" | cut -c 3- | sed -e 's/%0A//'
}

upload() {
    TIMEFORMAT=%R
    echo $1
    stat -f'%z' "$2/data.lsif"
    time curl -s -H 'Content-Type: application/x-ndjson+lsif' \
    "http://localhost:3080/.api/lsif/upload?repository=$(urlencode "$1")&commit=$(urlencode "$(git --git-dir $2/.git rev-parse HEAD)")&upload_token=$(urlencode "0000000000000000000000000000000000000000000000000000000000000000")" \
    --data-binary "@$2/data.lsif"
    echo
}

# Upload everything
upload github.com/sourcegraph/codeintellify ../../codeintellify
# upload github.com/sourcegraph/sourcegraph ..
# upload github.com/reactivex/rxjs ../../rxjs
