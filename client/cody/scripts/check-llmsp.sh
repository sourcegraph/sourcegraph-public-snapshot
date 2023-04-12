#!/bin/bash

run() {

	pushd "$(dirname "$(readlink -f "$0")")/../resources/bin" &> /dev/null || return
	trap 'popd > /dev/null' EXIT

	# Get this list by copying output of `ls resources/bin/ripgrep*`
	binaries=$(cat <<EOF
llmsp-v1.0.0-amd64-darwin
llmsp-v1.0.0-arm64-darwin
llmsp-v1.0.0-amd64-linux
llmsp-v1.0.0-arm64-linux
llmsp-v1.0.0-amd64-windows.exe
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
	echo "Need to build llmsp binaries. Try running 'scripts/build-llmsp.sh'? Afterward, update the list of binaries in 'scripts/check-llmsp.sh'"
	exit 1
fi
