import { ChildProcess, spawn } from 'child_process'
import gulp from 'gulp'
import { phabricator, watchPhabricator } from './client/browser/gulpfile'
import { graphQLTypes, schema, watchGraphQLTypes, watchSchema } from './shared/gulpfile'
import { webpack, webpackDevServer } from './web/gulpfile'

/**
 * Typechecks the TypeScript code.
 */
export function typescript(): ChildProcess {
    return spawn('yarn', ['-s', 'run', 'tsc', '-p', 'tsconfig.json', '--pretty'], {
        stdio: 'inherit',
        shell: true,
    })
}

/**
 * Builds everything.
 */
export const build = gulp.parallel(
    gulp.series(gulp.parallel(schema, graphQLTypes), typescript, gulp.parallel(webpack, phabricator))
)

export { schema, graphQLTypes, webpackDevServer, webpack }

/**
 * Watches everything and rebuilds on file changes.
 */
export const watch = gulp.series(
    // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
    gulp.parallel(schema, graphQLTypes),
    gulp.parallel(watchSchema, watchGraphQLTypes, webpackDevServer, watchPhabricator)
)
