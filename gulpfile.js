// @ts-check

const gulp = require('gulp')
const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
} = require('./client/shared/gulpfile')
const { webpack: webWebpack, webpackDevServer: webWebpackDevServer } = require('./client/web/gulpfile')

/**
 * Generates files needed for builds.
 */
const generate = gulp.parallel(schema, graphQlSchema, graphQlOperations)

/**
 * Starts all watchers on schema files.
 */
const watchGenerators = gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations)

/**
 * Generates files needed for builds whenever files change.
 */
const watchGenerate = gulp.series(generate, watchGenerators)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webWebpack)

/**
 * Watches everything and rebuilds on file changes.
 */
const dev = gulp.series(generate, gulp.parallel(watchGenerators, webWebpackDevServer))

module.exports = {
  generate,
  watchGenerate,
  build,
  dev,
  schema,
  graphQlSchema,
  watchGraphQlSchema,
  graphQlOperations,
  watchGraphQlOperations,
}
