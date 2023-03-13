#!/bin/bash

run() {

	pushd "$(dirname "$(readlink -f "$0")")/../resources/bin" &> /dev/null || return
	trap 'popd > /dev/null' EXIT

	# Get this list by copying output of `ls resources/bin/ripgrep*`
	binaries=$(cat <<EOF
ripgrep-v13.0.0-4-aarch64-apple-darwin
ripgrep-v13.0.0-4-aarch64-unknown-linux-gnu
ripgrep-v13.0.0-4-aarch64-unknown-linux-musl
ripgrep-v13.0.0-4-arm-unknown-linux-gnueabihf
ripgrep-v13.0.0-4-i686-unknown-linux-musl
ripgrep-v13.0.0-4-powerpc64le-unknown-linux-gnu
ripgrep-v13.0.0-4-s390x-unknown-linux-gnu
ripgrep-v13.0.0-4-x86_64-apple-darwin
ripgrep-v13.0.0-4-x86_64-unknown-linux-musl
EOF
			)

	for binary in $binaries; do
		if ls "$binary"; then
			continue
		fi
		return 1
	done
}

if run; then
	exit 0
else
	echo 'Need to download ripgrep binaries. Try running `scripts/download-rg.sh`? Afterward, update the list of binaries in `scripts/check-rg.sh`'
	exit 1
fi
