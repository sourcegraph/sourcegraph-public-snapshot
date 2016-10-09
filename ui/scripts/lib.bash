set -euf -o pipefail

function cleanupSourceFiles {
	rm -f "$VENDOR_DIR"/src/tsconfig.json
	find "$VENDOR_DIR" -name '*.ts' -not -name '*.d.ts' -delete
}

# In case we are killed, clean up by removing .ts files that we would
# remove anyway. If we don't do this, people might accidentally commit
# them (which would introduce a large commit and be confusing). Only
# the .d.ts and .js files should be committed.
trap cleanupSourceFiles EXIT


# gsed is required on OS X (brew install gnu-sed)
case $(uname) in
	Darwin*) sedi='gsed -i';;
	*) sedi='sed -i' ;;
esac


function fetchAndClean {
	# Use a bare repo so we don't have to worry about checking for a dirty
	# working tree.
	if [ -d "$CLONE_DIR" ]; then
		echo -n Updating git repository in "$CLONE_DIR"...
		git --git-dir="$CLONE_DIR" fetch --quiet
		echo OK
	else
		echo -n Cloning to "$CLONE_DIR"...
		git clone --quiet --bare --single-branch \
			"$CLONE_URL" "$CLONE_DIR"
		echo OK
	fi

	rm -rf "$VENDOR_DIR"/src/*
	mkdir -p "$VENDOR_DIR"
}
