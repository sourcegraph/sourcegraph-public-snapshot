.PHONY: dep test deploy

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

deploy:
	docker build -t us.gcr.io/sourcegraph-dev/e2e2 .
	gcloud docker -- push us.gcr.io/sourcegraph-dev/e2e2
	@echo "e2e client image uploaded. Kill the existing selenium clients to finish deploying."
