// @ts-check

const gulp = require('gulp')
const { graphQLTypes, schema, watchGraphQLTypes, watchSchema } = require('./shared/gulpfile')
const { webpack: webWebpack, webpackDevServer: webWebpackDevServer } = require('./web/gulpfile')

/**
 * Generates files needed for builds.
 */
const generate = gulp.parallel(schema, graphQLTypes)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webWebpack)

/**
 * Watches everything and rebuilds on file changes.
 */
const watch = gulp.series(generate, gulp.parallel(watchSchema, watchGraphQLTypes, webWebpackDevServer))

module.exports = { generate, build, watch, schema, graphQLTypes }
