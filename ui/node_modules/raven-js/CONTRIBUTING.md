# Contributing

## Reporting Issues and Asking Questions

Before opening an issue, please search the issue tracker to make sure your issue hasn't already been reported.

## Bugs and Improvements

We use the issue tracker to keep track of bugs and improvements to Raven.js itself, plugins, and the documentation. We encourage you to open issues to discuss improvements, architecture, implementation, etc. If a topic has been discussed before, we will ask you to join the previous discussion.

## Getting Help

For support or usage questions like “how do I do X with Raven.js and “my code doesn't work”, please search and ask on the [Sentry forum](https://forum.sentry.io).

## Help Us Help You

On both GitHub and the Sentry forum, it is a good idea to structure your code and question in a way that is easy to read and understand. For example, we encourage you to use syntax highlighting, indentation, and split text in paragraphs.

Additionally, it is helpful if you can let us know:

* The version of Raven.js affected
* The browser and OS affected
* Which Raven.js plugins are enabled, if any
* If you are using [hosted Sentry](https://sentry.io) or on-premises, and if the latter, which version (e.g. 8.7.0)
* If you are using the Raven CDN (http://ravenjs.com)

Lastly, it is strongly encouraged to provide a small project reproducing your issue. You can put your code on [JSFiddle](https://jsfiddle.net/) or, for bigger projects, on GitHub. Make sure all the necessary dependencies are declared in package.json so anyone can run npm install && npm start and reproduce your issue.

## Development

### Setting up an Environment

To run the test suite and run our code linter, node.js and npm are required. If you don't have node installed, [get it here](http://nodejs.org/download/) first.

Installing all other dependencies is as simple as:

```bash
$ npm install
```

And if you don't have [Grunt](http://gruntjs.com/) already, feel free to install that globally:

```bash
$ npm install -g grunt-cli
```

### Running the Test Suite

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

### Compiling Raven.js

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

_This is a checklist for core contributors when releasing a new version._

  * [ ] Verify TypeScript [language definition file](https://github.com/getsentry/raven-js/blob/master/typescript/raven.d.ts) is up to date
  * [ ] Bump version numbers in both `package.json` and `bower.json`.
  * [ ] Bump version across all docs under `docs/`
  * [ ] Put together [CHANGELOG](https://github.com/getsentry/raven-js/blob/master/CHANGELOG.md)
  * [ ] `$ grunt dist` This will compile a new version and update it in the `dist/` folder.
  * [ ] Confirm that build was fine, etc.
  * [ ] Commit new version, create a tag. Push to GitHub.
  * [ ] Copy CHANGELOG entry into a new GH Release: https://github.com/getsentry/raven-js/releases
  * [ ] `$ grunt publish` to recompile all plugins and all permutations and upload to S3.
  * [ ] `$ npm publish` to push to npm.
  * [ ] Confirm that the new version exists behind `cdn.ravenjs.com`
  * [ ] Bump version in the `gh-pages` branch specifically for http://ravenjs.com/.
  * [ ] Bump marketing pages on sentry.io, e.g. https://sentry.io/for/javascript/
  * [ ] Bump sentry.io `<script>` tag of raven.js
  * [ ] Bump `package.json` in Sentry repo https://github.com/getsentry/sentry/blob/master/package.json
  * [ ] Bump version for Segment integration since they don't: https://github.com/segment-integrations/analytics.js-integration-sentry
  * [ ] glhf
