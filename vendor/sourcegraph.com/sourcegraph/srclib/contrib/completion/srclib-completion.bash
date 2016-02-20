#!/bin/bash

_src() {
	args=("${COMP_WORDS[@]:1:$COMP_CWORD}")

	local IFS=$'\n'
	COMPREPLY=($(GO_FLAGS_COMPLETION=1 ${COMP_WORDS[0]} __complete -- "${args[@]}"))
	return 1
}

complete -F _srclib srclib
