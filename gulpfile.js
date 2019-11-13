// @ts-check

const gulp = require('gulp')
const { copyIntegrationAssets, watchIntegrationAssets } = require('./browser/gulpfile')
const { graphQLTypes, schema, watchGraphQLTypes, watchSchema } = require('./shared/gulpfile')
const { webpack: webWebpack, webpackDevServer: webWebpackDevServer } = require('./web/gulpfile')

/**
 * Generates files needed for builds.
 */
const generate = gulp.parallel(schema, graphQLTypes)

/**
 * Builds everything.
 */
const build = gulp.parallel(gulp.series(generate, gulp.parallel(webWebpack, copyIntegrationAssets)))

/**
 * Watches everything and rebuilds on file changes.
 */
const watch = gulp.series(
  generate,
  gulp.parallel(watchSchema, watchGraphQLTypes, webWebpackDevServer, watchIntegrationAssets)
)

module.exports = { generate, build, watch, schema, graphQLTypes }
