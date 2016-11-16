# webpack-dev-server

[![npm][npm]][npm-url]
[![deps][deps]][deps-url]
[![test][test]][test-url]

Use [webpack](http://webpack.github.io) with a development server that provides live reloading. This should be used for **development only**.

It uses [webpack-dev-middleware](https://github.com/webpack/webpack-dev-middleware) under the hood, which provides fast in-memory access to the webpack assets.

## Installation

```
npm install webpack-dev-server --save-dev
```

## Getting started

The easiest way to use it is with the CLI. In the directory where your `webpack.config.js` is, run:

```
node_modules/.bin/webpack-dev-server
```

This will start a server, listening on connections from `localhost` on port `8080`.

Now, when you change something in your assets, it should live-reload the files.

See [**the documentation**](http://webpack.github.io/docs/webpack-dev-server.html) for more use cases and options.

## Contributing

We appreciate all help! Check out [CONTRIBUTING.md](CONTRIBUTING.md) for more information on how to help.

## Maintainer

This project is maintained by [**Kees Kluskens**](https://github.com/spacek33z/).

## Inspiration

This project is heavily inspired by [peerigon/nof5](https://github.com/peerigon/nof5).


[npm]: https://img.shields.io/npm/v/webpack-dev-server.svg
[npm-url]: https://npmjs.com/package/webpack-dev-server

[deps]: https://david-dm.org/webpack/webpack-dev-server.svg
[deps-url]: https://david-dm.org/webpack/webpack-dev-server

[test]: http://img.shields.io/travis/webpack/webpack-dev-server.svg
[test-url]: https://travis-ci.org/webpack/webpack-dev-server
