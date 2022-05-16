// @ts-check

const gulp = require('gulp')

const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
} = require('./client/shared/gulpfile')
const { webpack: webWebpack, developmentServer, generate, watchGenerators } = require('./client/web/gulpfile')

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
const development = gulp.series(generate, gulp.parallel(watchGenerators, developmentServer))

module.exports = {
  generate,
  watchGenerate,
  build,
  dev: development,
  schema,
  graphQlSchema,
  watchGraphQlSchema,
  graphQlOperations,
  watchGraphQlOperations,
}
