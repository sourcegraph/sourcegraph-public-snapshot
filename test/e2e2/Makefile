.PHONY: test dep dep-browser-ext deploy vnc selenium run chrome-prod chrome-prod-novnc chrome-local firefox-prod firefox-local chrome-prod-repo-jump-to browser-ext

default: all-prod

test: dep dep-browser-ext
	BROWSER=${BROWSER} ./selenium.sh
	BROWSER=${BROWSER} ./vnc.sh &
	BROWSER=${BROWSER} SOURCEGRAPH_URL=${SOURCEGRAPH_URL} ./run.sh
	BROWSER=${BROWSER} ./selenium.sh kill

dep: .env/bin/python .env/bin/pip
	.env/bin/pip install -r requirements.txt

dep-browser-ext:
	cd ./browser-ext && yarn && yarn run build

.env/bin/python .env/bin/pip:
	virtualenv -p python2 .env || virtualenv -p python .env

deploy:
	cd ./browser-ext && make selenium-chrome
	./deploy.sh

vnc:
	./vnc.sh

selenium:
	./selenium.sh

run:
	./run.sh

#
# Demo recipes
#

all-prod:
	BROWSER=chrome SOURCEGRAPH_URL="https://sourcegraph.com" $(MAKE) test || true
	BROWSER=firefox SOURCEGRAPH_URL="https://sourcegraph.com" $(MAKE) test || true

chrome-prod:
	BROWSER=chrome SOURCEGRAPH_URL="https://sourcegraph.com" $(MAKE) test

browser-ext:
	BROWSER=chrome OPT="--filter=test_browser_extension" $(MAKE) test

chrome-prod-novnc:
	BROWSER=chrome SOURCEGRAPH_URL="https://sourcegraph.com" NOVNC=1 $(MAKE) test

chrome-local:
	BROWSER=chrome SOURCEGRAPH_URL="http://localhost:3080" $(MAKE) test

firefox-prod:
	BROWSER=firefox SOURCEGRAPH_URL="https://sourcegraph.com" $(MAKE) test

firefox-local:
	BROWSER=firefox SOURCEGRAPH_URL="http://localhost:3080" $(MAKE) test

chrome-prod-repo-jump-to:
	BROWSER=chrome SOURCEGRAPH_URL="https://sourcegraph.com" OPT="--filter=test_repo_jump_to" $(MAKE) test
