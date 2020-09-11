// @ts-check

const gulp = require('gulp')
const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
} = require('./shared/gulpfile')
const { webpack: webWebpack, webpackDevServer: webWebpackDevServer } = require('./web/gulpfile')

/**
 * Generates files needed for builds.
 */
const generate = gulp.parallel(schema, graphQlSchema, graphQlOperations)

/**
 * Generates files needed for builds whenever files change.
 */
const watchGenerate = gulp.series(generate, gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations))

/**
 * Builds everything.
 */
const build = gulp.series(generate, webWebpack)

/**
 * Watches everything and rebuilds on file changes.
 */
const watch = gulp.series(generate, gulp.parallel(watchGenerate, webWebpackDevServer))

module.exports = {
  generate,
  watchGenerate,
  build,
  watch,
  schema,
  graphQlSchema,
  watchGraphQlSchema,
  graphQlOperations,
  watchGraphQlOperations,
}
