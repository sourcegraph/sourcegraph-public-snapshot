# This file prepares and sets the LANGSERVER_<lang> env vars for the
# development lang servers.
#
# If a lang server should be available to all dev servers, then add
# installation and configuration steps here. Any existing
# LANGSERVER_<lang> env var MUST take precedence (if set, do not
# overwrite it).

export LANGSERVER_GO=${LANGSERVER_GO-:builtin:}

# TypeScript and JavaScript
#
# Install and use the TypeScript/JavaScript lang server (which are the
# same server) if yarn is installed (which is a good indicator of
# whether they care about TypeScript/JavaScript at all).
#
# To use your own javascript-typescript-langserver, just symlink
# $JSTS_LS_DIR to it.
if type yarn > /dev/null 2>&1 && ([[ -z "${LANGSERVER_TYPESCRIPT-}" ]] || [[ -z "${LANGSERVER_JAVASCRIPT-}" ]]); then
	JSTS_LS_DIR="${HOME}/.sourcegraph/lang/javascript-typescript-langserver"
	if [[ ! -d "$JSTS_LS_DIR" ]]; then
		mkdir -p "$JSTS_LS_DIR"/..
		git clone --quiet https://github.com/sourcegraph/javascript-typescript-langserver.git "$JSTS_LS_DIR"
	fi
	if [[ ! -d "$JSTS_LS_DIR"/node_modules ]]; then
		(cd "$JSTS_LS_DIR" && yarn install)
	fi
	if [[ ! -d "$JSTS_LS_DIR"/build ]]; then
		(cd "$JSTS_LS_DIR" && node_modules/.bin/tsc)
	fi
	export LANGSERVER_TYPESCRIPT=${LANGSERVER_TYPESCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
	export LANGSERVER_JAVASCRIPT=${LANGSERVER_JAVASCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
fi
