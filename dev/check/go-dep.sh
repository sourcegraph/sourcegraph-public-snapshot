#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

dep=dep-$(go env GOOS)-$(go env GOARCH)

curl -L -o /tmp/$dep https://github.com/golang/dep/releases/download/v0.5.0/$dep
chmod +x /tmp/$dep

/tmp/$dep check
