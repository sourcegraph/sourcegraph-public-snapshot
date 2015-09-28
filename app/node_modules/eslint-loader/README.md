# eslint-loader [![Build Status](http://img.shields.io/travis/MoOx/eslint-loader.svg)](https://travis-ci.org/MoOx/eslint-loader)

> eslint loader for webpack

## Install

```console
$ npm install eslint-loader
```

## Usage

In your webpack configuration

```javascript
module.exports = {
  // ...
  module: {
    loaders: [
      {test: /\.js$/, loader: "eslint-loader", exclude: /node_modules/}
    ]
  }
  // ...
}
```

When using with transpiling loaders (like `babel-loader`), make sure they are in correct order
(bottom to top). Otherwise files will be check after being processed by `babel-loader`

```javascript
module.exports = {
  // ...
  module: {
    loaders: [
      {test: /\.js$/, loader: "babel-loader", exclude: /node_modules/},
      {test: /\.js$/, loader: "eslint-loader", exclude: /node_modules/}
    ]
  }
  // ...
}
```

To be safe, you can use `preLoaders` section to check source files, not modified
by other loaders (like `babel-loader`)

```js
module.exports = {
  // ...
  module: {
    preLoaders: [
      {test: /\.js$/, loader: "eslint-loader", exclude: /node_modules/}
    ]
  }
  // ...
}
```

### Options

You can pass [eslint options](http://eslint.org/docs/user-guide/command-line-interface) directly by

- Adding a query string to the loader for this loader usabe only

```js
{
  module: {
    preLoaders: [
      {
        test: /\.js$/,
        loader: "eslint-loader?{rules:[{semi:0}]}",
        exclude: /node_modules/,
      },
    ],
  },
}
```

- Adding an `eslint` entry in your webpack config for global options:

```js
module.exports = {
  eslint: {
    configFile: 'path/.eslintrc'
  }
}
```

**Note that you can use both method in order to benefit from global & specific options**

#### `formatter` (default: eslint stylish formatter)

Loader accepts a function that will have one argument: an array of eslint messages (object).
The function must return the output as a string.
You can use official eslint formatters.

```js
module.exports = {
  entry: "...",
  module: {
    // ...
  }
  eslint: {
    // several examples !

    // default value
    formatter: require("eslint/lib/formatters/stylish"),

    // community formatter
    formatter: require("eslint-friendly-formatter"),

    // custom formatter
    formatter: function(results) {
      // `results` format is available here
      // http://eslint.org/docs/developer-guide/nodejs-api.html#executeonfiles()
      
      // you should return a string
      // DO NOT USE console.*() directly !
      return "OUTPUT"
    }
  }
}
```

#### Errors and Warning

**By default the loader will auto adjust error reporting depending
on eslint errors/warnings counts.**
You can still force this behavior by using `emitError` **or** `emitWarning` options:

##### `emitError` (default: `false`)

Loader will always return errors if this option is set to `true`.

```js
module.exports = {
  entry: "...",
  module: {
    // ...
  }
  eslint: {
    emitError: true
  }
}
```

##### `emitWarning` (default: `false`)

Loader will always return warnings if option is set to `true`.

#### `quiet` (default: `false`)

Loader will process and report errors only and ignore warnings if this option is set to true

```js
module.exports = {
  entry: "...",
  module: {
    // ...
  }
  eslint: {
    quiet: true
  }
}
```

##### `failOnWarning` (default: `false`)

Loader will cause the module build to fail if there are any eslint warnings.

```js
module.exports = {
  entry: "...",
  module: {
    // ...
  }
  eslint: {
    failOnWarning: true
  }
}
```

##### `failOnError` (default: `false`)

Loader will cause the module build to fail if there are any eslint errors.

```js
module.exports = {
  entry: "...",
  module: {
    // ...
  }
  eslint: {
    failOnError: true
  }
}
```

## Gotchas

### NoErrorsPlugin

`NoErrorsPlugin` prevents Webpack from outputting anything into a bundle. So even ESLint warnings
will fail the build. No matter what error settings are used for `eslint-loader`.

So if you want to see ESLint warnings in console during development using `WebpackDevServer`
remove `NoErrorsPlugin` from webpack config.

## [Changelog](CHANGELOG.md)

## [License](LICENSE)
