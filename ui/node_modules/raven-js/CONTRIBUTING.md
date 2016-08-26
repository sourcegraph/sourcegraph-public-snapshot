# Contributing

## Setting up an Environment

To run the test suite and run our code linter, node.js and npm are required. If you don't have node installed, [get it here](http://nodejs.org/download/) first.

Installing all other dependencies is as simple as:

```bash
$ npm install
```

And if you don't have [Grunt](http://gruntjs.com/) already, feel free to install that globally:

```bash
$ npm install -g grunt-cli
```

## Running the Test Suite

The test suite is powered by [Mocha](http://visionmedia.github.com/mocha/) and can both run from the command line or in the browser.

From the command line:

```bash
$ grunt test
```

From your browser:

```bash
$ grunt run:test
```

Then visit: http://localhost:8000/test/

## Compiling Raven.js

The simplest way to compile your own version of Raven.js is with the supplied grunt command:

```bash
$ grunt build
```

By default, this will compile raven.js and all of the included plugins.

If you only want to compile the core raven.js:

```bash
$ grunt build.core
```

Files are compiled into `build/`.

## Contributing Back Code

Please, send over suggestions and bug fixes in the form of pull requests on [GitHub](https://github.com/getsentry/raven-js). Any nontrivial fixes/features should include tests.
Do not include any changes to the `dist/` folder or bump version numbers yourself.

## Documentation

The documentation is written using [reStructuredText](http://en.wikipedia.org/wiki/ReStructuredText), and compiled using [Sphinx](http://sphinx-doc.org/). If you don't have Sphinx installed, you can do it using following command (assuming you have Python already installed in your system):

```bash
$ pip install sphinx
```

Documentation can be then compiled by running:

```bash
$ make docs
```

Afterwards you can view it in your browser by running following command and than pointing your browser to http://127.0.0.1:8000/:

```bash
$ grunt run:docs
```

## Releasing New Version

* Verify TypeScript [language definition file](https://github.com/getsentry/raven-js/blob/master/typescript/raven.d.ts) is up to date
* Bump version numbers in both `package.json` and `bower.json`.
* Bump version across all docs under `docs/`
* Put together [CHANGELOG](https://github.com/getsentry/raven-js/blob/master/CHANGELOG.md)
* `$ grunt dist` This will compile a new version and update it in the `dist/` folder.
* Confirm that build was fine, etc.
* Commit new version, create a tag. Push to GitHub.
* Copy CHANGELOG entry into a new GH Release: https://github.com/getsentry/raven-js/releases
* `$ grunt publish` to recompile all plugins and all permutations and upload to S3.
* `$ npm publish` to push to npm.
* Confirm that the new version exists behind `cdn.ravenjs.com`
* Bump version in the `gh-pages` branch specifically for http://ravenjs.com/.
* Bump marketing pages on getsentry.com, e.g. https://getsentry.com/for/javascript/
* Bump getsentry.com `<script>` tag of raven.js
* Bump `package.json` in Sentry repo https://github.com/getsentry/sentry/blob/master/package.json
* Bump version for Segment integration since they don't: https://github.com/segment-integrations/analytics.js-integration-sentry
* glhf
