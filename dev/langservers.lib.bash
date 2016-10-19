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
# To use your own langserver-{type,java}script, just symlink
# $TYPESCRIPT_DIR to it.
if type yarn > /dev/null 2>&1 && ([[ -z "${LANGSERVER_TYPESCRIPT-}" ]] || [[ -z "${LANGSERVER_JAVASCRIPT-}" ]]); then
	TYPESCRIPT_DIR="${HOME}/.sourcegraph/lang/langserver-typescript"
	if [[ ! -d "$TYPESCRIPT_DIR" ]]; then
		mkdir -p "$TYPESCRIPT_DIR"/..
		git clone --quiet https://github.com/sourcegraph/langserver-typescript.git "$TYPESCRIPT_DIR"
	fi
	if [[ ! -d "$TYPESCRIPT_DIR"/node_modules ]]; then
		(cd "$TYPESCRIPT_DIR" && yarn install)
	fi
	if [[ ! -d "$TYPESCRIPT_DIR"/build ]]; then
		(cd "$TYPESCRIPT_DIR" && node_modules/.bin/tsc)
	fi
	export LANGSERVER_TYPESCRIPT=${LANGSERVER_TYPESCRIPT-"$TYPESCRIPT_DIR"/bin/langserver-jsts}
	export LANGSERVER_JAVASCRIPT=${LANGSERVER_JAVASCRIPT-"$TYPESCRIPT_DIR"/bin/langserver-jsts}
fi
