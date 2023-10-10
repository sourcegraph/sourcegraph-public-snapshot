// This file is only used by Bazel builds.

const { esbuildBuildOptions } = require('./esbuild.js')

module.exports = {
  ...esbuildBuildOptions(process.env.NODE_ENV === 'development' ? 'dev' : 'prod'),

  // Unset configuration properties that are provided by Bazel.
  entryPoints: undefined,
  bundle: undefined,
  outdir: undefined,
  sourcemap: undefined,
}
