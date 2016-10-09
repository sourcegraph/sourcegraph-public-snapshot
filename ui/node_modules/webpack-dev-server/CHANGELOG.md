# 2.x (unreleased)

# 2.1.0-beta.7

- Add support for http/2 when using the `https` option. Falls back on http/1.1.

## 2.1.0-beta.6

- Start with tests! There's still a lot more to test, but at least there are _some_ tests now ([#623](https://github.com/webpack/webpack-dev-server/issues/623)).
- Add optional callback for the `close` API ([`1cf6549`](https://github.com/webpack/webpack-dev-server/commit/1cf6549415b078c80e027edbf6279a183fbcb631)).
- Fix `historyApiFallback` to fallback correctly to `contentBase` ([#617](https://github.com/webpack/webpack-dev-server/pull/617)).
- When using the `bypass` feature in a proxy, it was not possible to use in-memory webpack assets ([#613](https://github.com/webpack/webpack-dev-server/pull/613)).
- Simplify code for delivering assets ([#618](https://github.com/webpack/webpack-dev-server/issues/618)).

## 2.1.0-beta.5

- Add proxy config hot reloading - needs some additional configuration ([#605](https://github.com/webpack/webpack-dev-server/pull/605)).
- Fix `--progress` not working ([#609](https://github.com/webpack/webpack-dev-server/issues/609)).
- Fix `[WDS] Hot Module Replacement enabled` appearing even if the `clientLogLevel` was set to a non-`info` value ([#607](https://github.com/webpack/webpack-dev-server/pull/607)).
- Don't rely on a CDN for providing the SockJS script in iframe modus ([#493](https://github.com/webpack/webpack-dev-server/pull/493)).
- Explain what `--inline` does in help section for the CLI ([#596](https://github.com/webpack/webpack-dev-server/pull/596)).

## 2.1.0-beta.4

- Fix `contentBase` option in webpack config being ignored when using the CLI ([#597](https://github.com/webpack/webpack-dev-server/issues/597), [#599](https://github.com/webpack/webpack-dev-server/pull/599)).
- Fix SockJS providing an old SocKJS-client file, causing compatibility error ([#474](https://github.com/webpack/webpack-dev-server/issues/474)).
- Fix websocket connection issues when using https with a relative script path ([#592](https://github.com/webpack/webpack-dev-server/issues/592)).
- Fix hostname resolving issues ([#594](https://github.com/webpack/webpack-dev-server/pull/594)).
- Improve reliability of `--open` parameter ([#593](https://github.com/webpack/webpack-dev-server/issues/593)).

## 2.1.0-beta.3

- **Breaking change:** removed overriding `output.path` to `"/"` in the webpack config when using the CLI ([#337](https://github.com/webpack/webpack-dev-server/issues/337)). Note that `output.path` needs to be an absolute path!
- **Breaking change:** removed `contentBase` as a proxy feature (deprecated since 1.x).
- Limit websocket retries when the server can't be reached ([#589](https://github.com/webpack/webpack-dev-server/issues/589)).
- Improve detection for getting the server URL in the client ([#496](https://github.com/webpack/webpack-dev-server/issues/496)).
- Add `clientLogLevel` (`--client-log-level` for CLI) option. It controls the log messages shown in the browser. Available levels are `error`, `warning`, `info` or `none` ([#579](https://github.com/webpack/webpack-dev-server/issues/579)).
- Allow using no content base with the `--no-content-base` flag (previously it always defaulted to the working directory).
- Use stronger certs for the `https` modus, to prevent browsers from complaining about it ([#572](https://github.com/webpack/webpack-dev-server/issues/572)).

## 2.0 to 2.1.0-beta.2

- **Breaking change**: Only compatible with webpack v2.
- Add compatibility for web workers ([#298](https://github.com/webpack/webpack-dev-server/issues/298)).
- `--inline` is enabled by default now.
- Convert to `yargs` to handle commandline options.
- Allow a `Promise` instead of a config object in the CLI ([#419](https://github.com/webpack/webpack-dev-server/issues/419)).
- Add `--hot-only` flag, a shortcut that adds `webpack/hot/only-dev-server` in `entry` in the webpack config ([#439](https://github.com/webpack/webpack-dev-server/issues/439)).

For the 1.x changelog, see the [webpack-1 branch](https://github.com/webpack/webpack-dev-server/blob/webpack-1/CHANGELOG.md).
