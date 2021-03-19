#!/bin/bash

set -eu -o pipefail

# Prints out the HTTP status code for all of the absolute urls in ./redirects
dest_urls=$(awk '{print $2;}' < redirects | sort | uniq | grep '^https:')
for URL in $dest_urls; do
    echo -n "$URL "; curl -s -o /dev/null -w "%{http_code}\n" "$URL"
done
