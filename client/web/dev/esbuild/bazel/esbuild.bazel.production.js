// This file is only used by Bazel production builds.

process.env.NODE_ENV = 'production'

const { esbuildBuildOptions } = require('../config.js')
const { ENVIRONMENT_CONFIG } = require('../../utils/environment-config.js')

module.exports = {
  ...esbuildBuildOptions(ENVIRONMENT_CONFIG),

  // Unset configuration properties that are provided by Bazel.
  entryPoints: undefined,
  bundle: undefined,
  outdir: undefined,
  sourcemap: undefined,
  splitting: undefined,
  external: undefined,
}
