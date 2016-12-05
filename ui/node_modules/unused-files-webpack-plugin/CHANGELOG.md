# Change Log

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

<a name="3.0.0"></a>
# [3.0.0](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.5...v3.0.0) (2016-09-22)


### Bug Fixes

* **CHANGELOG.md:** deprecated v2.0.5 ([b7dfd07](https://github.com/tomchentw/unused-files-webpack-plugin/commit/b7dfd07))
* **index.js:** if bail is set, callback with error ([14b510a](https://github.com/tomchentw/unused-files-webpack-plugin/commit/14b510a))


### BREAKING CHANGES

* index.js: error reporting behaviour changed when bail is set

Before: if bail is set, there's still an error from UnusedFilesWebpackPlugin

After: if bail is set, no error will be emitted by UnusedFilesWebpackPlugin



<a name="2.0.5"></a>
## (Deprecated) [2.0.5](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.4...v2.0.5) (2016-09-20)

*Deprecated* due to https://github.com/tomchentw/unused-files-webpack-plugin/pull/11#issuecomment-248652421

### Bug Fixes

* **index.js:** if bail is set, callback with error ([#11](https://github.com/tomchentw/unused-files-webpack-plugin/issues/11)) ([23b85ad](https://github.com/tomchentw/unused-files-webpack-plugin/commit/23b85ad)), closes [#10](https://github.com/tomchentw/unused-files-webpack-plugin/issues/10)



<a name="2.0.4"></a>
## [2.0.4](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.3...v2.0.4) (2016-07-14)



<a name="2.0.3"></a>
## [2.0.3](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.2...v2.0.3) (2016-05-30)



<a name="2.0.2"></a>
## [2.0.2](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.1...v2.0.2) (2016-01-25)




<a name="2.0.1"></a>
## [2.0.1](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v2.0.0...v2.0.1) (2016-01-22)




<a name="2.0.0"></a>
# [2.0.0](https://github.com/tomchentw/unused-files-webpack-plugin/compare/v1.3.0...v2.0.0) (2016-01-22)


### Features

* **src:** rewrite in ES2015 format ([9a61f21](https://github.com/tomchentw/unused-files-webpack-plugin/commit/9a61f21))


### BREAKING CHANGES

* src: Removes commonjs module support.

Before:

```js
// webpack.config.js
var UnusedFilesWebpackPlugin = require("unused-files-webpack-plugin");
```

After:

In ES2015 module format:

```js
import UnusedFilesWebpackPlugin from "unused-files-webpack-plugin";
// it's the same as
import { default as UnusedFilesWebpackPlugin } from "unused-files-webpack-plugin";
// You could access from named export as well.
import { UnusedFilesWebpackPlugin } from "unused-files-webpack-plugin";
```

If you still use commonjs:

```js
var UnusedFilesWebpackPlugin = require("unused-files-webpack-plugin").default;
// or named export
var UnusedFilesWebpackPlugin = require("unused-files-webpack-plugin").UnusedFilesWebpackPlugin;
// with destructive assignment
var { UnusedFilesWebpackPlugin } = require("unused-files-webpack-plugin");
```



<a name"1.3.0"></a>
## 1.3.0 (2015-09-07)


#### Features

* **options:** add failOnUnused to enable generating error ([7b7620d8](https://github.com/tomchentw/unused-files-webpack-plugin/commit/7b7620d8), closes [#3](https://github.com/tomchentw/unused-files-webpack-plugin/issues/3))


<a name"1.2.0"></a>
## 1.2.0 (2015-07-11)


#### Features

* **globOptions:** makes ignore option overidable ([6b630944](https://github.com/tomchentw/unused-files-webpack-plugin/commit/6b630944), closes [#1](https://github.com/tomchentw/unused-files-webpack-plugin/issues/1))


#### Breaking Changes

* globOptions.ignore is now overridable

    If you choose to override globOptions with new ignore option,
    make sure you'll include `node_modules/**/*` for the new ignore.

 ([6b630944](https://github.com/tomchentw/unused-files-webpack-plugin/commit/6b630944))


## 1.1.0 (2015-05-22)


#### Bug Fixes

* **UnusedFilesWebpackPlugin:**
  * use objectAssign for default values ([f8a2b6f2](https://github.com/tomchentw/unused-files-webpack-plugin/commit/f8a2b6f28825ee6e3898c9f4b60f3e6a22d55bcb))
  * include emitted assets in fileDepsMap ([896e8e23](https://github.com/tomchentw/unused-files-webpack-plugin/commit/896e8e233557de43618ad700b40ed773db73f691))


## 1.0.0 (2015-05-22)


#### Features

* **UnusedFilesWebpackPlugin:** use glob to select files ([f8e081e8](https://github.com/tomchentw/unused-files-webpack-plugin/commit/f8e081e835344820c419dc37162c8028af7ba3f9))
