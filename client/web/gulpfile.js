const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './dev/tsconfig.json'),
})

const gulp = require('gulp')
const { createProxyMiddleware } = require('http-proxy-middleware')

const {
  graphQlOperations,
  schema,
  watchGraphQlOperations,
  watchSchema,
  cssModulesTypings,
  watchCSSModulesTypings,
} = require('../shared/gulpfile')

const { build: buildEsbuild } = require('./dev/esbuild/build')
const { esbuildDevelopmentServer } = require('./dev/esbuild/server')
const { DEV_SERVER_LISTEN_ADDR, DEV_SERVER_PROXY_TARGET_ADDR } = require('./dev/utils')

const webBuild = buildEsbuild

const esbuildDevelopmentProxy = () =>
  esbuildDevelopmentServer(DEV_SERVER_LISTEN_ADDR, app => {
    app.use(
      '/',
      createProxyMiddleware({
        target: {
          protocol: 'http:',
          host: DEV_SERVER_PROXY_TARGET_ADDR.host,
          port: DEV_SERVER_PROXY_TARGET_ADDR.port,
        },
        logLevel: 'error',
      })
    )
  })

const developmentServer = esbuildDevelopmentProxy

// Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
const generate = gulp.parallel(schema, graphQlOperations, cssModulesTypings)

// Watches code generation only, rebuilds on file changes
const watchGenerators = gulp.parallel(watchSchema, watchGraphQlOperations, watchCSSModulesTypings)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webBuild)

/**
 * Starts a development server without initial code generation, watches everything and rebuilds on file changes.
 */
const developmentWithoutInitialCodeGen = gulp.parallel(watchGenerators, developmentServer)

/**
 * Runs code generation first, then starts a development server, watches everything and rebuilds on file changes.
 */
const development = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  generate,
  developmentWithoutInitialCodeGen
)

/**
 * Watches everything, rebuilds on file changes and writes the bundle to disk.
 * Useful to running integration tests.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  generate,
  // invoked with WATCH=1
  gulp.parallel(watchGenerators, webBuild)
)

module.exports = {
  build,
  watch,
  dev: development,
  unsafeDev: developmentWithoutInitialCodeGen,
  webBuild,
  developmentServer,
  generate,
  watchGenerators,
}
