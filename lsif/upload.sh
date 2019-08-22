#!/usr/bin/env bash

urlencode() {
    echo "$1" | curl -Gso /dev/null -w %{url_effective} --data-urlencode @- "" | cut -c 3- | sed -e 's/%0A//'
}

file="$1"

usage() {
    echo "Sourcegraph LSIF uploader usage:"
    echo ""
    echo "env \\"
    echo "  SRC_ENDPOINT=<https://sourcegraph.example.com> \\"
    echo "  SRC_LSIF_UPLOAD_TOKEN=<secret> \\"
    echo "  REPOSITORY=<github.com/you/your-repo> \\"
    echo "  COMMIT=<40-char-hash> \\"
    echo "  bash upload-lsif.sh <file.lsif>"
}

if [[ -z "$SRC_LSIF_UPLOAD_TOKEN" || -z "$REPOSITORY" || -z "$COMMIT" || -z "$file" ]]; then
    usage
    exit 1
fi

curl \
    -H "Content-Type: application/x-ndjson+lsif" \
    "$SRC_ENDPOINT/.api/lsif/upload?repository=$(urlencode "$REPOSITORY")&commit=$(urlencode "$COMMIT")&upload_token=$(urlencode "$SRC_LSIF_UPLOAD_TOKEN")" \
    --data-binary "@$file"
