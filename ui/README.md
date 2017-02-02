app
=====

Run `yarn run start` in this directory to recompile assets while developing the app.

## Filesystem layout

* **../app/assets/** contains final products of assets only. Do not check in final
  products whose source files are also checked into the repository. (Do check in
  asset files, such as images, that aren't generated from other files also in
  the repository.)
* **{web_modules}/** contains source files used to generate final products for
  assets; the entry point for building these source files is `script/app.js`.
* **CSS Modules**: re-usable styles are located at `web_modules/sourcegraph/components/styles`
  while component-specific styles are localized to the component's directory.

To run all the tests for the UI code:

```
cd sourcegraph/ui
yarn test
```

To run a specific test:

```
cd sourcegraph/ui
node run-tests.js ./web_modules/path/to/The_Test.js
```


## Notes

* raven-js had to be patched manually to support CommonJS. The patch
  consists of running `npm install && grunt dist && rm -rf
  node_modules` in the `node_modules/raven-js` dir. That updates the
  dist/raven.js file to support CommonJS (the dist file in the repo
  didn't include updates made to templates/_footer.js to support
  CommonJS as of git release tag 1.1.16). **NOTE:** You'll need to
  repeat these steps if you ever update raven-js.
* react-icons has a dependency (react-icon-base) which shows a deprecation warning
  for using the className attribute on <svg> elements (which is provided by
  parent components in our application). The patch consists of changing
  `react-icon-base/lib/index.js` to rename `this.props.className` to
  `this.props.class`. **NOTE:** You'll need to repeat this step if you ever update
  react-icon-base.


### Hacking on vscode-languageclient

The vscode-languageclient npm package that we use is pending upstream
approval at
https://github.com/Microsoft/vscode-languageserver-node/pull/148. Until
it is merged upstream, our changes in that PR are published to the
@sourcegraph/vscode-languageclient package.

To make modifications to the package without requiring a
npm-publish-yarn-install cycle, you can use npm/yarn's `link` feature
to set up a symlink to your dev version:

1. In your local clone of the vscode-languageserver-node repository, run `cd client && yarn link`.
2. In the Sourcegraph `./ui` dir, run `yarn link vscode-languageclient`.
3. Ensure that there is a `yarn run watch` process running in `vscode-languageserver-node/client` to produce the `vscode-languageserver-node/client/lib/**` files.

Be sure to push, open a PR for, and npm-publish your changes to the
vscode-languageclient package if you make any.