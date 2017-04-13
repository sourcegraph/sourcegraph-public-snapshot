# Sourcegraph Phabricator integration

## Developing using the chrome extension
Our Chrome extension devloper mode can be used on the aws-test-cluster Phabricator instance we have running. Visit http://http://phabricator.aws.sgdev.org/, and you can use these credentials: `sourcegraph-test`/`Fyu2e4yb7hTcFs8J2E`.

Note: This server is also serving a version of the Phabricator integration. The two versions will interfere with one another. You must set a localStorage flag to prevent loading of the server-side Phabricator JavaScript. Run `window.localStorage['SOURCEGRAPH_DISABLED'] = true;`, and [this code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/init.tsx?_event=ExploreAtCursor#L18:100) prevents the server-side integration from loading.

## How Phabricator integrations work
Phabricator is able to serve up arbitrary JavaScript. Here are the instructions to do so:
* Create the JavaScript file in the rsrc directory, such as `/var/www/phabricator/webroot/rsrc/js/sourcegraph/sgdev-sourcegraph.js`
* I don't think the file name matters, at the top of the JS file, the plugin declares it's name. This looks something like:
* Make sure that your JS will get injected into the page in question. You need to modify some PHP code to require your JavaScript extension. I figured that the base/Controller.php code file gets loaded every page, so put the "require" statement there. Check out `/var/www/phabricator/src/applications/base/controller/PhabricatorController.php`. Check out line 66, `require_celerity_resource('sgdev-sourcegraph');`.
* Run `${PHABRICATOR_ROOT}/bin/celerity map`. This creates an index file where the PHP can find where sgdev-sourcegraph lives.
* Run `sudo service apache2 restart`, to restart Phabricator
* All this is available in the `restart.sh` script on the phabricator.aws.sgdev.org server.

## Deploying a server-side version of the Phabricator integration to aws-test-cluster
Talk to @neelance about getting you access to ubuntu@phabricator.aws.sgdev.org .

Theory: In general, the Chrome extension is a pretty good approximation for the behavior of the integration. However, it is important to test on the actual server. You either want to a) test the version of the code being served from the [aws-test-cluster version of Sourcegraph](http://node.aws.sgdev.org:30000/.assets/scripts/phabricator.bundle.js), OR you want to test JavaScript that resulted from `make phabricator` that is not yet deployed to the aws-test-cluster.

### Deploying a Phabricator integration that pulls its JavaScript from node.aws.sgdev.org:30000
`ssh ubuntu@phabricator.aws.sgdev.org`
`sh setup_hosted_script.sh`

### Deploying a Phabricator integration that has the JavaScript right in it
`scp sgdev.bundle.js ubuntu@phabricator.aws.sgdev.org:~/`, from your computer SCP the sgdev.bundle.js you want to test out to the sgdev server.
On the SGDev Phabricator server, run `sh setup_new_script.sh`, and this will append the right header to the bundle file, and restart the Phabricator server.

## How does the logic for the Chrome extension / Phabricator extension look
The main point of entry for the [Chrome extension](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/chrome/extension/inject.tsx#L27:15-27:32). When you are developing the extension in Chrome, this is the code path taken.

The main points of entry for the [sgdev.bundle.js](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/sgdev/sgdev.tsx) and for [umami.bundle.js](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/umami/umami.tsx#L1:1).
