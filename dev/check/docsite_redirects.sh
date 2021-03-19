#!/bin/bash

set -eu -o pipefail

REDIRECTS_FILE=../../doc/_resources/assets/redirects

# Prints out the HTTP status code for all of the absolute urls in the redirects file.
dest_urls=$(awk '{print $2;}' <"$REDIRECTS_FILE" | sort | uniq | grep '^https:')
for URL in $dest_urls; do
  echo -n "$URL "
  curl -s -o /dev/null -w "%{http_code}\n" "$URL"
done
