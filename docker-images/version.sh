#!/bin/bash

# If a subdirectory has its own version script, that produces the version number.
if [ -x $1/version.sh ]; then
	$1/version.sh
	exit 0
fi

# Otherwise, we just use the shortened Git SHA
git rev-parse --short HEAD
