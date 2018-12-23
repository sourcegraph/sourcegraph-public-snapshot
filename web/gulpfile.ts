import log from 'fancy-log'
import gulp from 'gulp'
import createWebpackCompiler, { Stats } from 'webpack'
import WebpackDevServer from 'webpack-dev-server'
import { phabricator, watchPhabricator } from '../client/browser/gulpfile'
import { graphQLTypes, schema, watchGraphQLTypes, watchSchema } from '../shared/gulpfile'
import webpackConfig from './webpack.config'

const WEBPACK_STATS_OPTIONS: Stats.ToStringOptions & { colors?: boolean } = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    warningsFilter: warning =>
        // This is intended, so ignore warning
        /node_modules\/monaco-editor\/.*\/editorSimpleWorker.js.*\n.*dependency is an expression/.test(warning),
    colors: true,
}
const logWebpackStats = (stats: Stats) => log(stats.toString(WEBPACK_STATS_OPTIONS))

export async function webpack(): Promise<void> {
    const compiler = createWebpackCompiler(webpackConfig)
    const stats = await new Promise<Stats>((resolve, reject) => {
        compiler.run((err, stats) => (err ? reject(err) : resolve(stats)))
    })
    logWebpackStats(stats)
    if (stats.hasErrors()) {
        throw Object.assign(new Error('Failed to compile'), { showStack: false })
    }
}

export async function webpackDevServer(): Promise<void> {
    const compiler = createWebpackCompiler(webpackConfig)
    const server = new WebpackDevServer(compiler as any, {
        allowedHosts: ['.host.docker.internal'],
        publicPath: '/.assets/',
        contentBase: './ui/assets',
        stats: WEBPACK_STATS_OPTIONS,
        noInfo: false,
        proxy: {
            '/': {
                target: 'http://localhost:3081',
                ws: true,
                // Avoid crashing on "read ECONNRESET".
                onError: err => console.error(err),
                onProxyReqWs: (_proxyReq, _req, socket) =>
                    socket.on('error', err => console.error('WebSocket proxy error:', err)),
            },
        },
    })
    return new Promise<void>((resolve, reject) => {
        server.listen(3080, '127.0.0.1', (err?: Error) => {
            if (err) {
                reject(err)
            } else {
                resolve()
            }
        })
    })
}

/**
 * Builds everything.
 */
export const build = gulp.parallel(
    gulp.series(gulp.parallel(schema, graphQLTypes), gulp.parallel(webpack, phabricator))
)

/**
 * Watches everything and rebuilds on file changes.
 */
export const watch = gulp.series(
    // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
    gulp.parallel(schema, graphQLTypes),
    gulp.parallel(watchSchema, watchGraphQLTypes, webpackDevServer, watchPhabricator)
)
