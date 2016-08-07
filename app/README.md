app
=====

Run `npm start` in this directory to recompile assets while developing the app.

## Filesystem layout

* **assets/** contains final products of assets only. Do not check in final
  products whose source files are also checked into the repository. (Do check in
  asset files, such as images, that aren't generated from other files also in
  the repository.)
* **{web_modules}/** contains source files used to generate final products for
  assets; the entry point for building these source files is `script/app.js`.
* **CSS Modules**: re-usable styles are located at `web_modules/sourcegraph/components/styles`
  while component-specific styles are localized to the component's directory.

To run all the tests for the `app/script` code:

```
cd sourcegraph/app
npm test
```

To run a specific test:

```
cd sourcegraph/app
env NODE_ENV=test ./node_modules/.bin/mocha-webpack ./web_modules/path/to/The_Test.js
```


## Notes

* raven-js had to be patched manually to support CommonJS. The patch
  consists of running `npm install && grunt dist && rm -rf
  node_modules` in the `node_modules/raven-js` dir. That updates the
  dist/raven.js file to support CommonJS (the dist file in the repo
  didn't include updates made to templates/_footer.js to support
  CommonJS as of git release tag 1.1.16). **NOTE:** You'll need to
  repeat these steps if you ever update raven-js.
* amplitude-js had to be patched manually for loading via ES6 modules. The patch
  consists of updating the package.json "main" script to reference `amplitude.js`,
  not `src/index.js`. **NOTE:** You'll need to repeat this step if you ever
  update amplitude-js.
* react-icons has a dependency (react-icon-base) which shows a deprecation warning
  for using the className attribute on <svg> elements (which is provided by
  parent components in our application). The patch consists of changing
  `react-icon-base/lib/index.js` to rename `this.props.className` to
  `this.props.class`. **NOTE:** You'll need to repeat this step if you ever update
  react-icon-base.
