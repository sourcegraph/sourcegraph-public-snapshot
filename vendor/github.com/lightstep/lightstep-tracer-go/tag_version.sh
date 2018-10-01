#!/bin/sh

VERSION=$(cat ./VERSION)

cat > version.go <<EOF
package lightstep

const TracerVersionValue = "$VERSION"
EOF
