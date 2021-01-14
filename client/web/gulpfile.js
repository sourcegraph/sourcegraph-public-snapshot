// @ts-check

const path = require('path')
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

const watchWebapp = () => gulp.watch(['src/**', '../shared/src/**'], { delay: 0, cwd: __dirname }, webapp)

/**
 * Builds everything.
 */
const build = gulp.series(gulp.parallel(schema, graphQlOperations, graphQlSchema), gulp.parallel(webapp))

/**
 * Watches everything, rebuilds on file changes and writes the bundle to disk.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are built to avoid first-time-run errors.
  gulp.parallel(schema, graphQlOperations, graphQlSchema),
  gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations, watchWebapp)
)

module.exports = {
  build,
  watch,
}
