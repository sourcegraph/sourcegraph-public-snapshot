.PHONY: dep test

test: dep
	@if [ "Darwin" = "$$(uname)" ]; then \
		./run_darwin.sh; \
	else \
		./run_linux.sh; \
	fi;

test-30: dep
	@if [ "Darwin" = "$$(uname)" ]; then \
		ITERS=30 ./run_darwin.sh; \
	else \
		ITERS=30 ./run_linux.sh; \
	fi;

dep: .env/bin/python .env/bin/pip
	.env/bin/pip install -r requirements.txt

.env/bin/python:
	virtualenv -p python2 .env

.env/bin/pip:
	virtualenv -p python2 .env
