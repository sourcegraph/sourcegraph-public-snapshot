# babel-preset-es2015-webpack

[![NPM version](http://img.shields.io/npm/v/babel-preset-es2015-webpack.svg?style=flat-square)](https://www.npmjs.org/package/babel-preset-es2015-webpack)
[![Travis build status](http://img.shields.io/travis/gajus/babel-preset-es2015-webpack/master.svg?style=flat-square)](https://travis-ci.org/gajus/babel-preset-es2015-webpack)
[![js-canonical-style](https://img.shields.io/badge/code%20style-canonical-blue.svg?style=flat-square)](https://github.com/gajus/canonical)

Babel preset for all es2015 plugins except [`babel-plugin-transform-es2015-modules-commonjs`](https://github.com/babel/babel/tree/master/packages/babel-plugin-transform-es2015-modules-commonjs).

This preset is used to enable ES2015 code compilation down to ES5. webpack 2 natively supports ES6 [`import`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/import) and [`export`](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Statements/export) statements. webpack 2 leverages the [static structure of the ES6 modules](http://exploringjs.com/es6/ch_modules.html#static-module-structure) to perform tree shaking.

For an introduction to tree shaking and webpack 2 see [Tree-shaking with webpack 2 and Babel 6](http://www.2ality.com/2015/12/webpack-tree-shaking.html).

## Install

```sh
npm install babel-preset-es2015-webpack --save-dev
```

## Usage

Add to `.babelrc`:

```json
{
    "presets": [
        "es2015-webpack"
    ]
}
```
