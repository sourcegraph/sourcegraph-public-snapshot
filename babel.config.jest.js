// @ts-check
const path = require('path')

// A minimal babel config only for jest transformations.
// All typescript and react transformations are done by previous
// bazel build rules, so we only need to do jest transformations here.

const logger = require('signale')

// TODO(bazel): drop when non-bazel removed.
if (!(process.env.BAZEL_BINDIR || process.env.BAZEL_TEST)) {
  throw new Error(__filename + ' is only for use with Bazel')
}

/** @type {import('@babel/core').ConfigFunction} */
module.exports = api => {
  api.cache.forever()

  /**
   * Whether to instrument files with istanbul for code coverage.
   * This is needed for e2e test coverage.
   */
  const instrument = Boolean(process.env.COVERAGE_INSTRUMENT && JSON.parse(process.env.COVERAGE_INSTRUMENT))
  if (instrument) {
    logger.info('Instrumenting code for coverage tracking')
  }

  return {
    presets: [
      // Can't put this in plugins because it needs to run as the last plugin.
      ...(instrument ? [{ plugins: [['babel-plugin-istanbul', { cwd: path.resolve(__dirname) }]] }] : []),
      '@babel/preset-env',
    ],
  }
}
