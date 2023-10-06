// @ts-check
const path = require('path')

const semver = require('semver')
const logger = require('signale')

/** @type {import('@babel/core').ConfigFunction} */
module.exports = api => {
  const isTest = api.env('test')
  api.cache.forever()

  /**
   * Whether to instrument files with istanbul for code coverage.
   * This is needed for e2e test coverage.
   */
  const instrument = Boolean(process.env.COVERAGE_INSTRUMENT && JSON.parse(process.env.COVERAGE_INSTRUMENT))
  if (instrument) {
    logger.info('Instrumenting code for coverage tracking')
  }

  /**
   * Do no use babel-preset-env for mocha tests transpilation in Bazel.
   * This is temporary workaround to allow us to use modern language featurs in `drive.page.evaluate` calls.
   */
  const disablePresetEnv = Boolean(process.env.DISABLE_PRESET_ENV && JSON.parse(process.env.DISABLE_PRESET_ENV))

  return {
    presets: [
      // Can't put this in plugins because it needs to run as the last plugin.
      ...(instrument ? [{ plugins: [['babel-plugin-istanbul', { cwd: path.resolve(__dirname) }]] }] : []),
      ...(disablePresetEnv
        ? []
        : [
            [
              '@babel/preset-env',
              {
                // Node (used for testing) doesn't support modules, so compile to CommonJS for testing.
                modules: process.env.BABEL_MODULE ?? (isTest ? 'commonjs' : false),
                bugfixes: true,
                useBuiltIns: 'entry',
                include: [
                  // Polyfill URL because Chrome and Firefox are not spec-compliant
                  // Hostnames of URIs with custom schemes (e.g. git) are not parsed out
                  'web.url',
                  // URLSearchParams.prototype.keys() is not iterable in Firefox
                  'web.url-search-params',
                  // Commonly needed by extensions (used by vscode-jsonrpc)
                  'web.immediate',
                  // Always define Symbol.observable before libraries are loaded, ensuring interopability between different libraries.
                  'esnext.symbol.observable',
                ],
                // See https://github.com/zloirock/core-js#babelpreset-env
                corejs: semver.minVersion(require('./package.json').dependencies['core-js']),
              },
            ],
          ]),
      '@babel/preset-typescript',
      [
        '@babel/preset-react',
        {
          runtime: 'automatic',
        },
      ],
    ],
    plugins: [
      '@babel/plugin-transform-runtime',
      ['@babel/plugin-transform-typescript', { isTSX: true }],
      'babel-plugin-lodash',
    ],
  }
}
