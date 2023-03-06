#!/bin/bash

VERSION="v13.0.0-4"

run() {
	pushd "$(dirname "$(readlink -f "$0")")/../resources/bin" || return
	trap 'popd' EXIT

	for url in $(curl https://api.github.com/repos/microsoft/ripgrep-prebuilt/releases/tags/$VERSION 2>/dev/null | jq -r '.assets[] | .browser_download_url'); do

		b=$(basename $url)
		ext=${b##*.}

		if [ "$ext" = "gz" ]; then
			stripped=${b%.tar.gz}

			echo "$url -> $stripped"
			wget -qO- "$url" | tar xvz -C ./ && mv ./rg "./$stripped"

		elif [ "$ext" = "zip" ]; then
			stripped=${b%.zip}
		else
			echo 'ERROR: unrecognized extension' $ext
		fi

	done
}

run
