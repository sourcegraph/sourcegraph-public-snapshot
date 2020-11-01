// @ts-check

const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './dev/tsconfig.json'),
})

const gulp = require('gulp')

const { spawn } = require('child_process')

const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
} = require('../shared/gulpfile')

// TODO(sqs): differentiate enterprise build
const isEnterpriseBuild = !!process.env.ENTERPRISE

const webapp = gulp.parallel(
  () =>
    spawn(path.join(__dirname, '..', '..', 'esbuild-js.sh'), [], {
      stdio: 'inherit',
    }),
  () =>
    spawn(path.join(__dirname, '..', '..', 'esbuild-css.sh'), [], {
      stdio: 'inherit',
    })
)

const watchWebapp = () => gulp.watch('src/**', { delay: 0, cwd: __dirname }, webapp)

// Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
const codeGen = gulp.parallel(schema, graphQlOperations, graphQlSchema)

// Watches code generation only, rebuilds on file changes
const watchCodeGen = gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations)

/**
 * Builds everything.
 */
const build = gulp.series(codeGen, webapp)

/**
 * Starts a development server without initial code generation, watches everything and rebuilds on file changes.
 */
const developmentWithoutInitialCodeGen = gulp.parallel(watchCodeGen, watchWebapp)

/**
 * Runs code generation first, then starts a development server, watches everything and rebuilds on file changes.
 */
const development = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  codeGen,
  developmentWithoutInitialCodeGen
)

/**
 * Watches everything, rebuilds on file changes and writes the bundle to disk.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  codeGen,
  gulp.parallel(watchCodeGen, watchWebapp)
)

module.exports = {
  build,
  watch,
  dev: development,
  unsafeDev: developmentWithoutInitialCodeGen,
  webapp,
  watchWebapp,
}
