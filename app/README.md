app
=====

Run `npm start` in this directory to recompile assets while developing the app.

## Filesystem layout

* **assets/** contains final products of assets only. Do not check in final
  products whose source files are also checked into the repository. (Do check in
  asset files, such as images, that aren't generated from other files also in
  the repository.)
* **{style,script}/** contains source files used to generate final products for
  assets.

## Testing deprecated app/script code

The `app/script` code is deprecated and is being migrated into a better
architecture at `app/web_modules/sourcegraph`. _Do not write new code in
app/script!_

To run all the tests for the `app/script` code:

```
cd sourcegraph/app
npm test
```

To run a specific test:

```
cd sourcegraph/app
NODE_PATH=web_modules ./node_modules/.bin/mocha --compilers js:babel-register --require babel-polyfill ./script/path/to/The_Test.js
```


## Notes

* raven-js had to be patched manually to support CommonJS. The patch
  consists of running `npm install && grunt dist && rm -rf
  node_modules` in the `node_modules/raven-js` dir. That updates the
  dist/raven.js file to support CommonJS (the dist file in the repo
  didn't include updates made to templates/_footer.js to support
  CommonJS as of git release tag 1.1.16). **NOTE:** You'll need to
  repeat these steps if you ever update raven-js.
