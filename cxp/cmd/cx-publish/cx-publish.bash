#!/bin/bash

set -euf -o pipefail

if [[ -z "${@-}" ]]; then
	echo "Usage: $0 extension-bundle.{js,map}"
	exit 1
fi

GCS_BUCKET=sourcegraph-cx-dev

# To create the GCS bucket:
#
# gsutil mb -c multi_regional gs://$GCS_BUCKET
# echo '[{"origin": ["*"],"responseHeader":["Content-Type"],"method":["GET","HEAD"],"maxAgeSeconds":2592000}]' | gsutil cors set /dev/stdin gs://$GCS_BUCKET

gsutil -h 'Cache-Control: no-transform, public, max-age=5' -q cp -a public-read -Z "$@" gs://$GCS_BUCKET
echo https://storage.googleapis.com/$GCS_BUCKET/$(basename $1)
