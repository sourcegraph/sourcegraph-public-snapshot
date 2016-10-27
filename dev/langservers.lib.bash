LS_ROOT="${HOME}/.sourcegraph/lang"

# Prepare and set the LANGSERVER_<lang> env vars for development lang
# servers in ~/.sourcegraph/lang and the builtin Go lang server.
detect_dev_langservers() {
	# Go (builtin)
	export LANGSERVER_GO=${LANGSERVER_GO-:builtin:}

	# JavaScript/TypeScript
	JSTS_LS_DIR="${LS_ROOT}/javascript-typescript-langserver"
	if [[ -d "$JSTS_LS_DIR" ]]; then
		echo '# Using javascript-typescript-langserver in '"$JSTS_LS_DIR"' (run `dev/install-langserver.sh javascript-typescript-langserver` to update)'
		export LANGSERVER_TYPESCRIPT=${LANGSERVER_TYPESCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
		export LANGSERVER_JAVASCRIPT=${LANGSERVER_JAVASCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
	else
		echo '# To add JavaScript/TypeScript language support, run `dev/install-langserver.sh javascript-typescript-langserver` or symlink '"$JSTS_LS_DIR"' to a local clone of javascript-typescript-langserver.'
	fi
}

install_langserver() {
	set -x
	LS_NAME=$1
	LS_DIR="${LS_ROOT}/${LS_NAME}"
	if [[ ! -d "$LS_DIR" ]]; then
		mkdir -p "$LS_DIR"/..
		git clone --quiet https://github.com/sourcegraph/"$LS_NAME".git "$LS_DIR"
	else
		(cd "$LS_DIR" && git pull)
	fi

	case "$LS_NAME" in
		javascript-typescript-langserver)
			(cd "$LS_DIR" && yarn install && node_modules/.bin/tsc)
			;;
		*)
			echo '# Do not know how to install '"$LS_NAME"'. See dev/langservers.lib.bash for a list of known language servers that can be installed using this method.'
			exit 1
			;;
	esac
}
