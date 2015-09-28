# Mostly lifted from https://andreypopp.com/posts/2013-05-16-makefile-recipes-for-node-js.html
# Thanks @andreypopp

BIN = ./node_modules/.bin
SRC = $(wildcard lib/*.js)
LIB = $(SRC:lib/%.js=dist/%.js)
MIN = $(SRC:lib/%.js=dist/%.min.js)

.PHONY: test dev

build: $(LIB) $(MIN)

# Allows usage of `make install`, `make link`
install link:
	@npm $@

# FIXME
dist/%.min.js: $(BIN)
	@$(BIN)/uglifyjs dist/react-draggable.js \
	  --output dist/react-draggable.min.js \
	  --source-map dist/react-draggable.min.map \
	  --source-map-url react-draggable.min.map \
	  --in-source-map dist/react-draggable.map \
	  --compress warnings=false

dist/%.js: $(BIN)
	@$(BIN)/webpack --devtool source-map

test: $(BIN)
	@$(BIN)/karma start --browsers Firefox --single-run

dev: $(BIN)
	script/build-watch

node_modules/.bin: install

define release
	VERSION=`node -pe "require('./package.json').version"` && \
	NEXT_VERSION=`node -pe "require('semver').inc(\"$$VERSION\", '$(1)')"` && \
	node -e "\
		['./package.json', './bower.json'].forEach(function(fileName) {\
			var j = require(fileName);\
			j.version = \"$$NEXT_VERSION\";\
			var s = JSON.stringify(j, null, 2);\
			require('fs').writeFileSync(fileName, s);\
		});" && \
	git add package.json bower.json CHANGELOG.md && \
	git add -f dist/ && \
	git commit -m "release v$$NEXT_VERSION" && \
	git tag "v$$NEXT_VERSION" -m "release v$$NEXT_VERSION"
endef

release-patch: test build
	@$(call release,patch)

release-minor: test build
	@$(call release,minor)

release-major: test build
	@$(call release,major)

publish:
	git push --tags origin HEAD:master
	npm publish
