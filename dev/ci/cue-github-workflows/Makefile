root="../../.."
file="dev/ci/cue-github-workflows/lint.cue"
output=".github/workflows/lint.yaml"

default: all

cuefmt:
	cd ${root} && cue fmt ${file}

build:
	cd ${root} && cue eval ${file} -e workflow --out yaml > ${output}

prettier:
	cd ${root} && node_modules/.bin/prettier ${output} --write --config prettier.config.js

all: build cuefmt prettier

.PHONY: all cuefmt build
