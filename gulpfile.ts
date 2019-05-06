import gulp from 'gulp'
import { phabricator, watchPhabricator } from './browser/gulpfile'
import { graphQLTypes, schema, watchGraphQLTypes, watchSchema } from './shared/gulpfile'
import { webpack as webWebpack, webpackDevServer as webWebpackDevServer } from './web/gulpfile'

/**
 * Generates files needed for builds.
 */
export const generate = gulp.parallel(schema, graphQLTypes)

/**
 * Builds everything.
 */
export const build = gulp.parallel(gulp.series(generate, gulp.parallel(webWebpack, phabricator)))

export { schema, graphQLTypes }

/**
 * Watches everything and rebuilds on file changes.
 */
export const watch = gulp.series(
    generate,
    gulp.parallel(watchSchema, watchGraphQLTypes, webWebpackDevServer, watchPhabricator)
)
