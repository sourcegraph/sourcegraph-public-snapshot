const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './tsconfig.json'),
})

const gulp = require('gulp')
const signale = require('signale')
const createWebpackCompiler = require('webpack')

const {
  graphQlOperations,
  schema,
  watchGraphQlOperations,
  watchSchema,
  cssModulesTypings,
  watchCSSModulesTypings,
} = require('../shared/gulpfile')

const buildScripts = require('./scripts/build')
const createWebpackConfig = require('./webpack.config')

const WEBPACK_STATS_OPTIONS = {
  all: false,
  timings: true,
  errors: true,
  warnings: true,
  colors: true,
}

/**
 * @param {import('webpack').Stats} stats
 */
const logWebpackStats = stats => {
  signale.log(stats.toString(WEBPACK_STATS_OPTIONS))
}

async function webpack() {
  const webpackConfig = await createWebpackConfig()
  const compiler = createWebpackCompiler(webpackConfig)
  /** @type {import('webpack').Stats} */
  const stats = await new Promise((resolve, reject) => {
    compiler.run((error, stats) => (error ? reject(error) : resolve(stats)))
  })
  logWebpackStats(stats)
  if (stats.hasErrors()) {
    throw Object.assign(new Error('Failed to compile'), { showStack: false })
  }
}

async function watchWebpack() {
  const webpackConfig = await createWebpackConfig()
  const compiler = createWebpackCompiler(webpackConfig)
  compiler.hooks.watchRun.tap('Notify', () => signale.log('Webpack compiling...'))
  await new Promise(() => {
    compiler.watch({ aggregateTimeout: 300 }, (error, stats) => {
      logWebpackStats(stats)
      if (error || stats.hasErrors()) {
        signale.error('Webpack compilation error')
      } else {
        signale.log('Webpack compilation done')
      }
    })
  })
}

async function esbuild() {
  await buildScripts.build()
}

// Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
const generate = gulp.parallel(schema, graphQlOperations, cssModulesTypings)

// Watches code generation only, rebuilds on file changes
const watchGenerators = gulp.parallel(watchSchema, watchGraphQlOperations, watchCSSModulesTypings)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webpack)

/**
 * Watches everything, rebuilds on file changes and writes the bundle to disk.
 * Useful to running integration tests.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  generate,
  gulp.parallel(watchGenerators, watchWebpack)
)

module.exports = { build, watch, webpack, watchWebpack, esbuild }
