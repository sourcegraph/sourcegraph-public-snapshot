REGIONS=${shell ./scripts/gen_variables.py --list-regions}

.PHONY: default all list ${REGIONS}

default:
	@echo Please read README.md

all: ${REGIONS}

list:
	@echo ${REGIONS}

validate:
	bash -c "packer validate --var-file=<(./scripts/gen_variables.py) baseimage.json"

${REGIONS}:
	bash -c "packer build --var-file=<(./scripts/gen_variables.py $@) baseimage.json"
