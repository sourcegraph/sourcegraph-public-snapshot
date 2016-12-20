LS_ROOT="${HOME}/.sourcegraph/lang"

# Prepare and set the LANGSERVER_<lang> env vars for development lang
# servers in ~/.sourcegraph/lang and the builtin Go lang server.
detect_dev_langservers() {
	# Go (builtin)
	export LANGSERVER_GO=${LANGSERVER_GO-:builtin:}
	export LANGSERVER_GO_BG=${LANGSERVER_GO_BG-:builtin:}

	CSS_LS_DIR="${LS_ROOT}/css-langserver"
	if [[ -d "$CSS_LS_DIR" ]]; then
		export LANGSERVER_CSS=${LANGSERVER_CSS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
		export LANGSERVER_LESS=${LANGSERVER_LESS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
		export LANGSERVER_SCSS=${LANGSERVER_SCSS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
	else
		echo '# To add css/less/scss language support, run `dev/install-langserver.sh css-langserver`'
	fi

	# JavaScript/TypeScript
	JSTS_LS_DIR="${LS_ROOT}/javascript-typescript-langserver"
	if [[ -d "$JSTS_LS_DIR" ]]; then
		export LANGSERVER_TYPESCRIPT=${LANGSERVER_TYPESCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
		export LANGSERVER_JAVASCRIPT=${LANGSERVER_JAVASCRIPT-"$JSTS_LS_DIR"/bin/javascript-typescript-langserver}
	else
		echo '# To add JavaScript/TypeScript language support, run `dev/install-langserver.sh javascript-typescript-langserver` or symlink '"$JSTS_LS_DIR"' to a local clone of javascript-typescript-langserver.'
	fi

	# Python
	PY_LS_DIR="${LS_ROOT}/python-langserver"
	if [[ -d "$PY_LS_DIR" ]]; then
		export LANGSERVER_PYTHON=${LANGSERVER_PYTHON-"$PY_LS_DIR"/bin/python-langserver}
	else
		echo '# To add Python language support, run `dev/install-langserver.sh python-langserver` or symlink '"$PY_LS_DIR"' to a local clone of python-langserver.'
	fi

	# PHP
	if [[ -n "${LANGSERVER_PHP:+1}" ]]; then
		# no-op, the user has already set the server
		true
	elif [[ $(hash docker 2>/dev/null && docker images -q felixfbecker/php-language-server) ]]; then
		export LANGSERVER_PHP=$(dirname "${BASH_SOURCE[0]}")/php-langserver
	else
		echo '# To add PHP language support, run `docker pull felixfbecker/php-language-server`'
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
		css-langserver)
			(cd "$LS_DIR/langserver" && yarn && node_modules/.bin/tsc)
			;;
		javascript-typescript-langserver)
			(cd "$LS_DIR" && yarn && node_modules/.bin/tsc)
			;;
		python-langserver)
			(cd "$LS_DIR" && pip3 install -r requirements.txt)
			;;
		*)
			echo '# Do not know how to install '"$LS_NAME"'. See dev/langservers.lib.bash for a list of known language servers that can be installed using this method.'
			exit 1
			;;
	esac
}
