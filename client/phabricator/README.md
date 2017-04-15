# Sourcegraph Phabricator integration

## Sourcegraph development stack
This stack is a simulated Umami environment.

* [aws-test-cluster Sourcegraph instance](http://node.aws.sgdev.org:30000/)
To access this instance using kubectl, the infrastructure repo has an on-prem folder, and the aws-test-cluster folder has the relevant configs.
* [aws-test-cluster Phabricator instance](http://phabricator.aws.sgdev.org/)
Richard controls this. In the root direcotry of this box (ubuntu user), you'll find all the scripts in the [scripts directory.](./scripts).
To log in, use these credentials: `sourcegraph-test`/`Fyu2e4yb7hTcFs8J2E`.
* [aws-test-cluster gitolite instance](git@gitolite.aws.sgdev.org:<REPO_NAME>)

## Developing using the chrome extension
Note: http://phabricator.aws.sgdev.org/ is usually serving a version of the Phabricator integration. The two versions will interfere with one another. You must set a localStorage flag to prevent loading of the server-side Phabricator JavaScript. Run `window.localStorage['SOURCEGRAPH_DISABLED'] = true;`, and [this code](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/init.tsx?_event=ExploreAtCursor#L18:100) prevents the server-side integration from loading.

Other than that, load the Chrome extension in developer mode, and it should run on http://phabricator.aws.sgdev.org/. Try it on a [diffusion view](http://phabricator.aws.sgdev.org/source/nmux/browse/master/mux.go) and a [differential view](http://phabricator.aws.sgdev.org/D3).

## Pushing changes to production
In the [browser-ext](../browser-ext/) directory, run `make phabricator`. This will automatically generate 4 files - the sgdev.bundle.js\[.map\] files, and the umami.bundle.js\[.map\] files. The sgdev bundle files are used by the aws-test-cluster Phabricator extension, and the umami ones are used at umami. These files are copied in to the [ui/assets/scripts](../../ui/assets/scripts) folder. Once these assets are commited, they will be available from the Sourcegraph frontend pod @ <SOURCEGRAPH_URL>/.assets/scripts/phabricator.bundle.js.

Note: Our build process will automatically deploy the changes to Sourcegraph.com. However, to deploy changes on node.aws.sgdev.org or umami, you will need to update the frontend pods on those instances.

## How Phabricator integrations work
Phabricator is able to serve up arbitrary JavaScript. Here is how it works on the aws-test-cluster phabricator host:
* Create the JavaScript file in the rsrc directory, such as `/var/www/phabricator/webroot/rsrc/js/sourcegraph/sgdev-sourcegraph.js`
* I don't think the file name matters, at the top of the JS file, the plugin declares it's name. This looks like: [scripts/base.js](./scripts/base.js).
* Make sure that your JS will get injected into the page in question. You need to modify some PHP code to require your JavaScript extension. I figured that the base/Controller.php code file gets loaded every page, so put the "require" statement there. Look at `/var/www/phabricator/src/applications/base/controller/PhabricatorController.php` on the phabricator server; on line 66, you'll see `require_celerity_resource('sgdev-sourcegraph');`.
* Run `${PHABRICATOR_ROOT}/bin/celerity map`. This creates an index file where the PHP can find where sgdev-sourcegraph lives.
* Run `sudo service apache2 restart`, to restart Phabricator
* All this is available in the `restart.sh` script on the phabricator.aws.sgdev.org serverm also in the scripts/ directory.

## Deploying a server-side version of the Phabricator integration to aws-test-cluster
Talk to @neelance about getting you access to ubuntu@phabricator.aws.sgdev.org .

Theory: In general, the Chrome extension is a pretty good approximation for the behavior of the integration. However, it is important to test on the actual server. You either want to a) test the version of the code being served from the [aws-test-cluster version of Sourcegraph](http://node.aws.sgdev.org:30000/.assets/scripts/phabricator.bundle.js), OR you want to test JavaScript that resulted from `make phabricator` that is not yet deployed to the aws-test-cluster.

### Deploying a Phabricator integration that pulls its JavaScript from node.aws.sgdev.org:30000
`ssh ubuntu@phabricator.aws.sgdev.org`
`sh setup_hosted_script.sh`

### Deploying a Phabricator integration that has the JavaScript right in it
`scp sgdev.bundle.js ubuntu@phabricator.aws.sgdev.org:~/`, from your computer SCP the sgdev.bundle.js you want to test out to the sgdev server.
On the SGDev Phabricator server, run `sh setup_script_from_bundle.sh`, and this will append the right header to the bundle file, and restart the Phabricator server.

## How does the logic for the Chrome extension / Phabricator extension look
The main point of entry for the [Chrome extension](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/chrome/extension/inject.tsx#L27:15-27:32). When you are developing the extension in Chrome, this is the code path taken.

The main points of entry for the [sgdev.bundle.js](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/sgdev/sgdev.tsx) and for [umami.bundle.js](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@f9c7474eb5dbd4477146ac4907c3098539075417/-/blob/client/browser-ext/phabricator/umami/umami.tsx#L1:1).

## Questions

### Why are there different bundles for sgdev and umami?
Phabricator requires a mapping from repo nickname to repo URI. Also, our plugin requires an address for where to find the Sourcegraph instance. This is the only difference between the bundles.

### Long term plans
Ideally, we could get the repo nickname -> uri mapping using the Phabricator API, either directly from the front-end, or Sourcegraph would do this. I think either could probably work, and there is a [conduit API](http://phabricator.aws.sgdev.org/conduit/method/diffusion.repository.search/) that performs this operation.
