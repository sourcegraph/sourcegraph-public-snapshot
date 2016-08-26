bin := ./node_modules/.bin

all: browserify

lint:
	@ $(bin)/eslint lib/*.js test/*-test.js

browserify: lint
	@ $(bin)/browserify test/*-test.js > test/index.js

test: browserify
	@ if [ "$(SAUCE_USERNAME)" != "" ] && [ "$(SAUCE_ACCESS_KEY)" != "" ]; \
	then \
		$(bin)/easy-sauce; \
	else \
		$(bin)/mocha-phantomjs test/index.html; \
	fi

.PHONY: all lint test
