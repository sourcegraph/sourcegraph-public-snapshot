# 1.8.3 (2016-09-26)

The last few releases had a few bugs in them that shouldn't have reached stable. We're working hard on adding enough tests to prevent these kind of failures.

- Fix for `output.path` check with node < `0.12` ([`38ff513`](https://github.com/webpack/webpack-dev-middleware/commit/38ff513cf6a5e5a909134db0fc6614a1831fbdd1)).
- Fix for `output.path` check on Windows with node < `5.0` ([`38ff513`](https://github.com/webpack/webpack-dev-middleware/commit/38ff513cf6a5e5a909134db0fc6614a1831fbdd1)).
- Fix exception when using the deprecated `watchDelay` option ([#131](https://github.com/webpack/webpack-dev-middleware/pull/131)).

# 1.8.2 (2016-09-23)

- Add a nice error message when `output.path` in the webpack config is relative.
- Expose stats when using the advanced API ([#127](https://github.com/webpack/webpack-dev-middleware/pull/127)).
- Make log & warn functions replaceable ([#124](https://github.com/webpack/webpack-dev-middleware/pull/124)).
- Fix protocol-relative urls with `publicPath` ([#129](https://github.com/webpack/webpack-dev-middleware/pull/129)).

# 1.8.1 (2016-09-18)

- Fix exception thrown when using webpack array config ([#125](https://github.com/webpack/webpack-dev-middleware/issues/125)).

# 1.8.0 (2016-09-17)

- Add tests for everything ([#121](https://github.com/webpack/webpack-dev-middleware/issues/121))!
- Use valid license for npm ([`af6243b`](https://github.com/webpack/webpack-dev-middleware/commit/af6243b4f4cf7da00923c8cddef20c4cfb5d145c)).
- Instantiate the in-memory filesystem only once, meaning it can be shared when using multiple middleware instances with one webpack config ([#120](https://github.com/webpack/webpack-dev-middleware/pull/120)).
- When sending requests with a `Range` header, the status code would not always be correct ([`cedc09f`](https://github.com/webpack/webpack-dev-middleware/commit/cedc09f714fa1e8ef35cbe373466c6d56db0ac4f)).
- Middleware would not work with a proxy ([#80](https://github.com/webpack/webpack-dev-middleware/pull/80)).
- Add `index` option (default: `index.html`) to customize the file used as index for a directory ([#123](https://github.com/webpack/webpack-dev-middleware/pull/123)).

# 1.7.0 (2016-09-10)

- Add `reporter` option for custom logs ([#91](https://github.com/webpack/webpack-dev-middleware/pull/91)).
- Add `serverSideRender` option to help support server-side rendering ([#118](https://github.com/webpack/webpack-dev-middleware/pull/118)).
- Set `statusCode` to 200 for served files ([#94](https://github.com/webpack/webpack-dev-middleware/pull/94)).
- Fix `waitUntilValid` callback always being called immediately when using MultiCompiler ([#100](https://github.com/webpack/webpack-dev-middleware/pull/100)).
- Fix range support not working ([#116](https://github.com/webpack/webpack-dev-middleware/pull/116)).
- Only use middleware for GET requests ([#96](https://github.com/webpack/webpack-dev-middleware/pull/96)).
