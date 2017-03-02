package app

import (
	"bytes"
	"net/http"
)

var installScript = `#!/bin/sh

# This installation script can be curled from the internet to install zap like:
#
#  sh <(curl -sSf https://sourcegraph.com/install/zap)
#
# It simply:
#
#  - Detects the platform.
#  - Downloads a Zap binary from a Google Cloud Storage bucket.
#  - Installs it into the system directory.
#  - Optionally installs the system service for you.
#

set -u

main() {
	must get_os
	local _os="$RETVAL";

	must get_arch
	local _arch="$RETVAL";

	if [[ "$_os" == "windows" ]]; then
		err "Windows is not currently supported, sorry."
	fi
	if [[ "$_arch" == "386" ]]; then
		err "32-bit binary downloads are not currently available, sorry."
	fi

	if command -v zap > /dev/null 2>&1; then
		log "zap is already installed at $(command -v zap)"
		read -p "installer: Replace the existing installation? [n] " -n 1 -r
		echo ""
		if [[ $REPLY =~ ^[Yy]$ ]]
		then
			if [[ "$_os" == "linux" ]]; then
				# Linux requires sudo to write into /usr/local/bin
				sudo rm $(command -v zap)
			else
				rm $(command -v zap)
			fi
		else
			err "Not replacing existing installation. Aborting."
		fi
	fi

	if [[ "$_os" == "darwin" || "$_os" == "linux" ]]; then
		log "Downloading the latest zap binary..."
		curl "https://storage.googleapis.com/sourcegraph-zap/updates/bin/zap-main-${_os}-${_arch}-latest" -Sf --progress > /tmp/zap

		# Inform the user if we're going to prompt them for a password.
		if [ $(sudo -n uptime 2>&1|grep "load"|wc -l) -lt 0 ]; then
			log "You will now be prompted for your password (so we can install the binary into /usr/local/bin)"
		fi
		must sudo cp /tmp/zap /usr/local/bin/zap
		must sudo chmod +x /usr/local/bin/zap
		must rm /tmp/zap
		log "Successfully installed Zap to /usr/local/bin/zap"
	else
		# should never get here
		err "Unsupported OS installation type $_os"
	fi

	# Install the server as a background daemon
	must sudo zap server install

	echo ""
	echo "Success! Next steps:"
	echo ""
	echo " - check for updates: use 'zap version'"
	echo " - manage the server: use 'zap server [restart|start|stop]'"
	echo " - ⚡️  watch a repository: use 'zap init'"
	echo ""
}

get_os() {
	if [[ "$OSTYPE" == "linux-gnu" ]]; then
		local _os="linux"
	elif [[ "$OSTYPE" == "darwin"* || "$OSTYPE" == "bamp" ]]; then
		local _os="darwin"
	elif [[ "$OSTYPE" == "cygwin" || "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
		local _os="windows"
	else
		err "Unsupported OS type $OSTYPE"
		local _os=""
	fi
	RETVAL="$_os"
}

get_arch() {
	# TODO(slimsag): windows support here
	if [[ $(getconf LONG_BIT) = "64" ]]; then
		local _arch="amd64"
	elif [[ $(getconf LONG_BIT) = "32" ]]; then
		local _arch="386"
	else
		err "Unsupported CPU type $(getconf LONG_BIT)"
		local _arch=""
	fi
	RETVAL="$_arch"
}

log() {
	echo "installer: $1" >&2
}

# err echos an err to stderr and exits the program.
err() {
	echo "error: $1" >&2
	exit 1
}

# must ensures the given command does not fail. If it does, the program exits.
must() {
	"$@"
	if [[ $? != 0 ]]; then err "command failed: $*"; fi
}

main "$@" || exit 1
`

func serveInstallZap(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/x-sh")
	w.WriteHeader(http.StatusOK)
	_, err := bytes.NewBufferString(installScript).WriteTo(w)
	return err
}
