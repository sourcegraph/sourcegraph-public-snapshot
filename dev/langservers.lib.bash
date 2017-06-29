LS_ROOT="${HOME}/.sourcegraph/lang"

# Prepare and set the LANGSERVER_<lang> env vars for development lang
# servers in ~/.sourcegraph/lang and the builtin Go lang server.
detect_dev_langservers() {
	# Go. We can assume server.sh has installed the binary and running the
	# xlang-go server.
	export LANGSERVER_GO=${LANGSERVER_GO-"tcp://localhost:4389"}
	export LANGSERVER_GO_BG=${LANGSERVER_GO_BG-"tcp://localhost:4389"}

	CSS_LS_DIR="${LS_ROOT}/css-langserver"
	if [[ -d "$CSS_LS_DIR" ]]; then
		export LANGSERVER_CSS=${LANGSERVER_CSS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
		export LANGSERVER_LESS=${LANGSERVER_LESS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
		export LANGSERVER_SCSS=${LANGSERVER_SCSS-"$CSS_LS_DIR"/bin/css-langserver-stdio}
	else
		echo '# To add css/less/scss language support, run `dev/install-langserver.sh css-langserver`'
	fi

	# JavaScript/TypeScript
	JSTS_LS_DIR=$(dirname "${BASH_SOURCE[0]}")/../xlang/javascript-typescript/buildserver
	if [[ -d "$JSTS_LS_DIR/lib" ]]; then
		export LANGSERVER_TYPESCRIPT=${LANGSERVER_TYPESCRIPT-"$JSTS_LS_DIR"/lib/language-server-stdio.js}
		export LANGSERVER_TYPESCRIPT_ARGS_JSON='["--strict"]'
		export LANGSERVER_TYPESCRIPT_BG=${LANGSERVER_TYPESCRIPT}
		export LANGSERVER_TYPESCRIPT_BG_ARGS_JSON='["--strict"]'
		export LANGSERVER_JAVASCRIPT=${LANGSERVER_JAVASCRIPT-"$JSTS_LS_DIR"/lib/language-server-stdio.js}
		export LANGSERVER_JAVASCRIPT_ARGS_JSON='["--strict"]'
		export LANGSERVER_JAVASCRIPT_BG=${LANGSERVER_JAVASCRIPT}
		export LANGSERVER_JAVASCRIPT_BG_ARGS_JSON='["--strict"]'
	else
		echo '# To add JavaScript/TypeScript language support, run `dev/install-langserver.sh javascript-typescript`'
	fi

	# Python
	PY_LS_DIR="${LS_ROOT}/python-langserver"
	if [[ -d "$PY_LS_DIR" ]]; then
		export LANGSERVER_PYTHON=${LANGSERVER_PYTHON-$(dirname "${BASH_SOURCE[0]}")/python-langserver}
	else
		echo '# To add Python language support, run `dev/install-langserver.sh python-langserver` or symlink '"$PY_LS_DIR"' to a local clone of python-langserver.'
	fi

	# Java
	JAVA_LS_DIR="${LS_ROOT}/java-langserver"
	if [[ -d "$JAVA_LS_DIR" ]]; then
		export LANGSERVER_JAVA=${LANGSERVER_JAVA-"$JAVA_LS_DIR"/bin/java-langserver}
	else
		echo '# To add Java language support, run `dev/install-langserver.sh java-langserver` or symlink '"$JAVA_LS_DIR"' to a local clone of java-langserver.'
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
	
	# Swift
	SWIFT_LS_DIR="${LS_ROOT}/swift-langserver"
	if [[ -d "$SWIFT_LS_DIR" ]]; then
		export LANGSERVER_SWIFT=${LANGSERVER_SWIFT-$(dirname "${BASH_SOURCE[0]}")/swift-langserver}
	else
		echo '# To add Swift language support, run `dev/install-langserver.sh swift-langserver` or symlink '"$SWIFT_LS_DIR"' to a local clone of swift-langserver.'
	fi
}

install_langserver() (
	set -x
	local LS_NAME=$1
	local LS_DIR="${LS_ROOT}/${LS_NAME}"
	clone_repo() {
		if [[ ! -d "$LS_DIR" ]]; then
			mkdir -p "$LS_DIR"/..
			git clone --quiet git@github.com:sourcegraph/"$LS_NAME".git "$LS_DIR"
		else
			(cd "$LS_DIR" && git pull)
		fi
	}

	case "$LS_NAME" in
		css-langserver)
			clone_repo
			(cd "$LS_DIR/langserver" && yarn && node_modules/.bin/tsc)
			;;
		javascript-typescript)
			(cd $(dirname "${BASH_SOURCE[0]}")/../xlang/javascript-typescript/buildserver && yarn && yarn run build)
			;;
		java-langserver)
			clone_repo
			(cd "$LS_DIR" && mvn clean compile assembly:single)
			;;
		python-langserver)
			clone_repo
			(
			    cd "$LS_DIR"
			    test -d venv || python3 -m venv venv
			    venv/bin/pip install -r requirements.txt
			)
			;;
		swift-langserver)
			clone_repo
			(cd "$LS_DIR" && brew install sourcekitten && go install ./)
			;;
		*)
			echo '# Do not know how to install '"$LS_NAME"'. See dev/langservers.lib.bash for a list of known language servers that can be installed using this method.'
			exit 1
			;;
	esac
)
