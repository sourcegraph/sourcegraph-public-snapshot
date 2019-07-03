#!/usr/bin/env bash

file="$1"

usage() {
    echo "Sourcegraph LSIF uploader usage:"
    echo ""
    echo "env \\"
    echo "  SRC_ENDPOINT=<https://sourcegraph.example.com> \\"
    echo "  SRC_ACCESS_TOKEN=<secret> \\"
    echo "  REPOSITORY=<github.com/you/your-repo> \\"
    echo "  COMMIT=<40-char-hash> \\"
    echo "  bash upload-lsif.sh <file.lsif>"
}

if [[ -z "$SRC_ACCESS_TOKEN" || -z "$REPOSITORY" || -z "$COMMIT" || -z "$file" ]]; then
  usage
  exit 1
fi

curl \
  -F "data=@$file" \
  -H "Authorization: token $SRC_ACCESS_TOKEN" \
  "$SRC_ENDPOINT/upload?repository=$REPOSITORY&commit=$COMMIT"
