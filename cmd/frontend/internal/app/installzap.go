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
# Or uninstalled via:
#
# 	sh <(curl -sSf https://sourcegraph.com/install/zap) --uninstall
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

	local _option=${1:-}
	if [[ "$_option" == "--uninstall" ]]; then
		do_uninstall
		return
	fi

	if command -v zap > /dev/null 2>&1; then
		log "zap is already installed at $(command -v zap)"
		read -p "installer: Replace the existing installation? [n] " -n 1 -r
		echo ""
		if [[ $REPLY =~ ^[Yy]$ ]]
		then
			do_uninstall
		else
			err "Not replacing existing installation. Aborting."
		fi
	fi

	if [[ "$_os" == "darwin" ]]; then
		log "Downloading the latest zap binary..."
		curl "https://storage.googleapis.com/sourcegraph-zap/updates/bin/zap-main-${_os}-${_arch}-latest" -Sf --connect-timeout 30 --progress > /tmp/zap
		must cp /tmp/zap /usr/local/bin/zap
		must chmod +x /usr/local/bin/zap
		must rm /tmp/zap
		log "Successfully installed Zap to /usr/local/bin/zap"
	elif [[ "$_os" == "linux" ]]; then
		log "Downloading the latest zap binary..."
		curl "https://storage.googleapis.com/sourcegraph-zap/updates/bin/zap-main-${_os}-${_arch}-latest" -Sf --connect-timeout 30 --progress > /tmp/zap
		# Linux requires sudo to write into /usr/local/bin
		sudo_prompt "You will now be prompted for your sudo password (so we can install the binary to /usr/local/bin)"
		must sudo cp /tmp/zap /usr/local/bin/zap
		must sudo chmod +x /usr/local/bin/zap
		must rm /tmp/zap
		log "Successfully installed Zap to /usr/local/bin/zap"
	else
		# should never get here
		err "Unsupported OS installation type $_os"
	fi

	# Install the server as a background daemon
	sudo_prompt "You will now be prompted for your sudo password (so we can install zap server as a daemon)"
	must sudo zap server install

	# Ensure the zap server is running (if it isn't the user will have a bad
	# time).
	local _timeout=$(($(date +%s) + 5)) # 5s timeout
	while sleep 0.1; do
		zap server status 2>&1 | grep 'ZapServer running' &> /dev/null
		if [ $? == 0 ]; then
			break # sucess
		elif [[ $(date +%s) -gt $_timeout ]]; then
			err "unexpected: zap server did not start"
		fi
	done

	echo ""
	echo "Success! Next steps:"
	echo ""
	echo " - check for updates: use 'zap version'"
	echo " - manage the server: use 'zap server [restart|start|stop]'"
	echo " - ⚡️  watch a repository: use 'zap on'"
	echo ""
}

do_uninstall() {
	must get_os
	local _os="$RETVAL";

	must get_arch
	local _arch="$RETVAL";

	# Uninstall the background daemon, if a zap binary exists.
	if command -v zap > /dev/null 2>&1; then
		sudo_prompt "You will now be prompted for your sudo password (so we can uninstall the background daemon)"
		sudo zap server uninstall
	fi

	# In case the zap binary didn't exist, perform best-effort cleanup.
	if [[ "$_os" == "darwin" ]]; then
		sudo_prompt "You will now be prompted for your sudo password (so we can remove the binary from /usr/local/bin)"
		launchctl remove ZapServer
		if pgrep -x "zap" > /dev/null; then
			killall zap
		fi
	else
		# should never get here
		err "Unsupported OS installation type $_os"
	fi

	# Remove the zap binary.
	if command -v zap > /dev/null 2>&1; then
		if [[ "$_os" == "linux" ]]; then
			# Linux requires sudo to write into /usr/local/bin
			sudo_prompt "You will now be prompted for your sudo password (so we can remove the binary from /usr/local/bin)"
			must sudo rm $(command -v zap)
		else
			must rm $(command -v zap)
		fi
	fi

	# Remove zap system state.
	if [[ "$_os" == "darwin" ]]; then
		must rm -f /tmp/zap-local-server
		sudo_prompt "You will now be prompted for your sudo password (so we can remove /var/log/zap.log)"
		must sudo rm -f /var/log/zap.log /tmp/zap-update-state 
	elif [[ "$_os" == "linux" ]]; then
		must rm -f /tmp/zap-local-server
		sudo_prompt "You will now be prompted for your sudo password (so we can remove /var/log/zap.log)"
		must sudo rm -f /var/log/zap.log /tmp/zap-update-state 
	else
		# should never get here
		err "Unsupported OS installation type $_os"
	fi

	# Ensure the zap binary was removed (this also handles weird cases
	# where a user has two zap binaries on their PATH and we only removed
	# the first).
	if command -v zap > /dev/null 2>&1; then
		err "unexpected: zap binary still located at $(command -v zap)"
	fi

	# Ensure the zap server process is no longer running, just to be on the
	# safe side.
	local _timeout=$(($(date +%s) + 5)) # 5s timeout
	while sleep 0.1; do
		if ! pgrep -x "zap" > /dev/null; then
			break # sucess
		elif [[ $(date +%s) -gt $_timeout ]]; then
			err "unexpected: zap server is still running after it was stopped"
		fi
	done
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

# sudo_prompt informs the user if we're going to prompt them for their sudo
# password.
sudo_prompt() {
	if [ $(sudo -n uptime 2>&1|grep "load"|wc -l) -le 0 ]; then
		log "$1"
	fi
}

main "$@" || exit 1
`

func serveInstallZap(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "text/x-sh")
	w.WriteHeader(http.StatusOK)
	_, err := bytes.NewBufferString(installScript).WriteTo(w)
	return err
}
