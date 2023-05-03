#!/bin/sh

if [ "${TARGETARCH}" = "arm64" ]; then
	cat <<- FOE >> .cargo/config.toml
	[env]
	CFLAGS="-mno-outline-atomics"
	FOE
fi;
