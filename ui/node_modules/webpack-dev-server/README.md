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

## Inspiration

This project is heavily inspired by [peerigon/nof5](https://github.com/peerigon/nof5).

## Contributing

The client scripts are built with `npm run-script prepublish`.

Run the relevant [examples](https://github.com/webpack/webpack-dev-server/tree/master/examples) to see if all functionality still works. When introducing new functionality, also add an example. This helps the maintainers to understand it and check if it still works.

When making a PR, keep these goals in mind:

- The communication library (`SockJS`) should not be exposed to the user.
- A user should not try to implement stuff that accesses the webpack filesystem, because this lead to bugs (the middleware does it while blocking requests until the compilation has finished, the blocking is important).
- It should be a development only tool (compiling in production is bad, one should precompile and deliver the compiled assets).
- There are hooks to add your own features, so we should not add less-common features.
- Processing options and stats display is delegated to webpack, so webpack-dev-server/middleware should not do much with it. This also helps us to keep up-to-date with webpack updates.
- The workflow should be to start webpack-dev-server as a separate process, next to the "normal" server and to request the script from this server or to proxy from dev-server to "normal" server (because webpack blocks the event queue too much while compiling which can affect "normal" server).

## License

Copyright 2012-2016 Tobias Koppers

[MIT](http://www.opensource.org/licenses/mit-license.php)

[npm]: https://img.shields.io/npm/v/webpack-dev-server.svg
[npm-url]: https://npmjs.com/package/webpack-dev-server

[deps]: https://david-dm.org/webpack/webpack-dev-server.svg
[deps-url]: https://david-dm.org/webpack/webpack-dev-server

[test]: http://img.shields.io/travis/webpack/webpack-dev-server.svg
[test-url]: https://travis-ci.org/webpack/webpack-dev-server
