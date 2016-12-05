<a name="4.0.1"></a>
## [4.0.1](https://github.com/vvakame/typescript-formatter/compare/4.0.0...v4.0.1) (2016-11-16)

* **tsfmt:** add typescript `>=2.2.0-dev` to peerDependencies ([#68](https://github.com/vvakame/typescript-formatter/pull/68)) thanks @myitcv !

<a name="4.0.0"></a>
# [4.0.0](https://github.com/vvakame/typescript-formatter/compare/3.1.0...v4.0.0) (2016-10-27)

Now, typescript-formatter supports `typescript@^2.0.6`.
If you want to use with older version typescript, you can use `typescript-formatter@^3.0.0`.

### Features

* **tsfmt:** support TypeScript 2.0.6 ([26db3de](https://github.com/vvakame/typescript-formatter/commit/26db3de))



<a name="3.1.0"></a>
# [3.1.0](https://github.com/vvakame/typescript-formatter/compare/3.0.1...v3.1.0) (2016-10-09)


### Features

* **tsfmt:** move final newline character logic to editorconfig part ([2df1f7a](https://github.com/vvakame/typescript-formatter/commit/2df1f7a))

thanks @jrieken !

<a name="3.0.1"></a>
## [3.0.1](https://github.com/vvakame/typescript-formatter/compare/3.0.0...v3.0.1) (2016-09-23)

[TypeScript 2.0.3 released](https://blogs.msdn.microsoft.com/typescript/2016/09/22/announcing-typescript-2-0/)! yay!

### Features

* **example:** update example code ([3b365be](https://github.com/vvakame/typescript-formatter/commit/3b365be))



<a name="3.0.0"></a>
# [3.0.0](https://github.com/vvakame/typescript-formatter/compare/2.3.0...v3.0.0) (2016-08-19)


### Features

* **tsfmt:** support comments in tsconfig.json & tsfmt.json & tslint.json ([5a4fdfd](https://github.com/vvakame/typescript-formatter/commit/5a4fdfd))
* **tsfmt:** support include, exclude properties [@tsconfig](https://github.com/tsconfig).json when using --replace options [#48](https://github.com/vvakame/typescript-formatter/issues/48) ([d8e71f5](https://github.com/vvakame/typescript-formatter/commit/d8e71f5))
* **tsfmt:** update peerDependencies, remove tsc ^1.0.0 ([35c1d62](https://github.com/vvakame/typescript-formatter/commit/35c1d62))



<a name="2.3.0"></a>
# [2.3.0](https://github.com/vvakame/typescript-formatter/compare/2.2.1...v2.3.0) (2016-07-16)


### Features

* **tsfmt:** support TypeScript 2.0.0 and next ([38dc72e](https://github.com/vvakame/typescript-formatter/commit/38dc72e))



<a name="2.2.1"></a>
## [2.2.1](https://github.com/vvakame/typescript-formatter/compare/2.2.0...v2.2.1) (2016-06-30)

### Features

* **tsfmt:** Add 'next' support for TypeScript 2.0.0-dev. ([35a371c](https://github.com/vvakame/typescript-formatter/commit/35a371c))



<a name="2.2.0"></a>
# [2.2.0](https://github.com/vvakame/typescript-formatter/compare/2.1.0...v2.2.0) (2016-05-14)


### Bug Fixes

* **tsfmt:** check rules.indent[1] is "tabs" fromt tslint fixes [#42](https://github.com/vvakame/typescript-formatter/issues/42) ([450c467](https://github.com/vvakame/typescript-formatter/commit/450c467)), closes [#42](https://github.com/vvakame/typescript-formatter/issues/42)



<a name="2.1.0"></a>
# [2.1.0](https://github.com/vvakame/typescript-formatter/compare/2.0.0...v2.1.0) (2016-02-25)


### Bug Fixes

* **ci:** fix Travis CI failed ([68a9c7c](https://github.com/vvakame/typescript-formatter/commit/68a9c7c))

### Features

* **tsfmt:** support typescript@1.8.2. add `insertSpaceAfterOpeningAndBeforeClosingTemplateStringBraces`. ([611fee0](https://github.com/vvakame/typescript-formatter/commit/611fee0))



<a name="2.0.0"></a>
# [2.0.0](https://github.com/vvakame/typescript-formatter/compare/1.2.0...v2.0.0) (2016-02-06)


### Features

* **tsfmt:** remove es6-promise from dependencies. tsfmt supports after latest LTS of node.js ([19a7f44](https://github.com/vvakame/typescript-formatter/commit/19a7f44))
* **tsfmt:** remove typescript from dependencies and add to peerDependencies refs #30 ([b8a58c6](https://github.com/vvakame/typescript-formatter/commit/b8a58c6))
* **tsfmt:** update dependencies. support TypeScript 1.7.5 ([bb9cd81](https://github.com/vvakame/typescript-formatter/commit/bb9cd81))



<a name="1.2.0"></a>
# [1.2.0](https://github.com/vvakame/typescript-formatter/compare/1.1.0...v1.2.0) (2015-12-01)


### Features

* **tsfmt:** update dependencies. support TypeScript 1.7.3 ([abd22cf](https://github.com/vvakame/typescript-formatter/commit/abd22cf))



<a name="1.1.0"></a>
# [1.1.0](https://github.com/vvakame/typescript-formatter/compare/1.0.0...v1.1.0) (2015-10-14)


### Bug Fixes

* **tsfmt:** replace line delimiter to formatOptions.NewLineCharacter fixes #26 ([8d84ddb](https://github.com/vvakame/typescript-formatter/commit/8d84ddb)), closes [#26](https://github.com/vvakame/typescript-formatter/issues/26)

### Features

* **example:** update example, support typescript-formatter@1.0.0 ([dd283b3](https://github.com/vvakame/typescript-formatter/commit/dd283b3))
* **tsfmt:** add support for filesGlob. thanks @ximenean #25 ([bf9f704](https://github.com/vvakame/typescript-formatter/commit/bf9f704))
* **tsfmt:** support newline char settings from tsconfig.json ([d618ee8](https://github.com/vvakame/typescript-formatter/commit/d618ee8))

<a name="1.0.0"></a>
# [1.0.0](https://github.com/vvakame/typescript-formatter/compare/0.4.3...v1.0.0) (2015-09-22)


### Features

* **ci:** use `sudo: false` and switch to node.js v4 ([29b0f45](https://github.com/vvakame/typescript-formatter/commit/29b0f45))
* **tsfmt:** add baseDir options closes #23 ([b69c4b6](https://github.com/vvakame/typescript-formatter/commit/b69c4b6)), closes [#23](https://github.com/vvakame/typescript-formatter/issues/23)
* **tsfmt:** add tsconfig.json support. thanks @robertknight #22 ([cb52bd4](https://github.com/vvakame/typescript-formatter/commit/cb52bd4))
* **tsfmt:** change tsc version specied. strict to loose. ([ea4401c](https://github.com/vvakame/typescript-formatter/commit/ea4401c))
* **tsfmt:** fix many issue by @myitcv #24 ([d0f2719](https://github.com/vvakame/typescript-formatter/commit/d0f2719)), closes [#24](https://github.com/vvakame/typescript-formatter/issues/24)
* **tsfmt:** pass Options object to providers ([c425bac](https://github.com/vvakame/typescript-formatter/commit/c425bac))
* **tsfmt:** refactor to es6 style ([2941857](https://github.com/vvakame/typescript-formatter/commit/2941857))
* **tsfmt:** update dependencies, switch to typescript@1.6.2, change build process (tsconfig. ([d8f5670](https://github.com/vvakame/typescript-formatter/commit/d8f5670))



<a name="0.4.3"></a>
## 0.4.3 (2015-08-04)


### Features

* **tsfmt:** pass specified file name to typescript language service. support tsx files. ([b9196e9](https://github.com/vvakame/typescript-formatter/commit/b9196e9))



<a name="0.4.2"></a>
## 0.4.2 (2015-07-26)


### Bug Fixes

* **tsfmt:** remove trailing white chars and add linefeed ([3843e40](https://github.com/vvakame/typescript-formatter/commit/3843e40))



<a name"0.4.0"></a>
## 0.4.0 (2015-06-28)


#### Features

* **tsfmt:** support --verify option ([8dd0f8ee](https://github.com/vvakame/typescript-formatter/commit/8dd0f8ee), closes [#15](https://github.com/vvakame/typescript-formatter/issues/15))


<a name"0.3.2"></a>
### 0.3.2 (2015-05-08)


#### Features

* **tsfmt:** change --stdio option to do not required fileName ([32055514](https://github.com/vvakame/typescript-formatter/commit/32055514))


<a name"0.3.1"></a>
### 0.3.1 (2015-05-06)


#### Features

* **tsfmt:** support typescript@1.5.0-beta ([a5f7f19a](https://github.com/vvakame/typescript-formatter/commit/a5f7f19a))


<a name="0.3.0"></a>
## 0.3.0 (2015-03-22)


#### Features

* **tsfmt:** support --stdin option refs #9 ([e322fc74](git@github.com:vvakame/typescript-formatter/commit/e322fc74eb4b62f908a8a7c0f8c0c736bd933631))


<a name="0.2.2"></a>
### 0.2.2 (2015-02-24)


#### Bug Fixes

* **tsfmt:** fix .d.ts file generation refs #7 ([f5520ec6](git@github.com:vvakame/typescript-formatter/commit/f5520ec65c2a034c40884e07276abc4a9a210ca9))


<a name="0.2.1"></a>
### 0.2.1 (2015-02-18)


#### Features

* **tsfmt:** add grunt-dts-bundle and generate typescript-formatter.d.ts ([c846cf37](git@github.com:vvakame/typescript-formatter/commit/c846cf3762982b9bb23bc6b617155488c125d2ad))


<a name="0.2.0"></a>
## 0.2.0 (2015-02-14)

TypeScript 1.4.1 support!

#### Bug Fixes

* **deps:**
  * bump editorconfig version ([68140595](git@github.com:vvakame/typescript-formatter/commit/681405952ed68071cd97d5358bc0fb153f76d841))
  * remove jquery.d.ts dependency ([ae3b52c6](git@github.com:vvakame/typescript-formatter/commit/ae3b52c6faa69bec862f370fc6dd8e86e429a92d))


#### Features

* **deps:**
  * add grunt-conventional-chagelog ([bbe79771](git@github.com:vvakame/typescript-formatter/commit/bbe797712227c0ce6a70bf2e7baf95e41f939126))
  * remove grunt-espower and add espower-loader, refactor project ([4f213464](git@github.com:vvakame/typescript-formatter/commit/4f21346472cca229c089dd91abd65667c03c6c66))
* **grunt:** remove TypeScript compiler specified ([b241945a](git@github.com:vvakame/typescript-formatter/commit/b241945a13e77ca1db25fdb35d1dd4e9ba3dff27))
* **tsfmt:** add typescript package to dependencies and remove typescript-toolbox submodule ([48d176e9](git@github.com:vvakame/typescript-formatter/commit/48d176e967e67ec41aef2402f299fd99330cde33))
