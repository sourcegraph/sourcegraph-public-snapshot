# Install these:
#     pip install squeeze
#     pip install Pygments==dev
# Install nodejs and npm
# Install these:
#     npm install uglify-js@latest  # or use the latest from the github repo.
#     npm install docco
# Install coffee:
#     git clone https://github.com/jashkenas/coffee-script.git
#     cd coffee-script
#     sudo bin/cake install
SQUEEZE=squeeze
UGLIFYJS=uglifyjs
DOCCO=docco

.PHONY: all clean docs

packed: string_score.min.js string_score.uglify.js

all: packed docs string_score.js

string_score.min.js: string_score.js
	@echo "Minifying (YUICompressor) string_score.js into string_score.min.js"
	@$(SQUEEZE) yuicompressor --type=js string_score.js > string_score.min.js

string_score.uglify.js: string_score.js
	@echo "Minifying (UglifyJS) string_score.js into string_score.uglify.js"
	@$(UGLIFYJS) -nc string_score.js > string_score.uglify.js

string_score.js: coffee/string_score.coffee
	coffee -b -c coffee/string_score.coffee

docs: coffee/string_score.coffee
	$(DOCCO) coffee/string_score.coffee

clean:
	@-rm -rf string_score.*.js
	@-rm -rf score*js
	@-rm -rf docs/
	@-rm -f coffee/score.js

