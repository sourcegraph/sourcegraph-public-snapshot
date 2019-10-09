import log from 'fancy-log'
import gulp from 'gulp'
import createWebpackCompiler, { Stats } from 'webpack'
import WebpackDevServer, { addDevServerEntrypoints } from 'webpack-dev-server'
import { copyIntegrationAssets, watchIntegrationAssets } from '../browser/gulpfile'
import { graphQLTypes, schema, watchGraphQLTypes, watchSchema } from '../shared/gulpfile'
import webpackConfig from './webpack.config'

const WEBPACK_STATS_OPTIONS: Stats.ToStringOptions & { colors?: boolean } = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    colors: true,
}
const logWebpackStats = (stats: Stats): void => {
    log(stats.toString(WEBPACK_STATS_OPTIONS))
}

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
    const options: WebpackDevServer.Configuration & { liveReload?: boolean } = {
        hot: !process.env.NO_HOT,
        inline: !process.env.NO_HOT,
        allowedHosts: ['.host.docker.internal'],
        host: 'localhost',
        port: 3080,
        publicPath: '/.assets/',
        contentBase: './ui/assets',
        stats: WEBPACK_STATS_OPTIONS,
        noInfo: false,
        disableHostCheck: true,
        proxy: {
            '/': {
                target: 'http://localhost:3081',
                // Avoid crashing on "read ECONNRESET".
                onError: err => console.error(err),
                onProxyReqWs: (_proxyReq, _req, socket) =>
                    socket.on('error', err => console.error('WebSocket proxy error:', err)),
            },
        },
    }
    addDevServerEntrypoints(webpackConfig, options)
    const server = new WebpackDevServer(createWebpackCompiler(webpackConfig), options)
    await new Promise<void>((resolve, reject) => {
        server.listen(3080, '0.0.0.0', (err?: Error) => (err ? reject(err) : resolve()))
    })
}

/**
 * Builds everything.
 */
export const build = gulp.parallel(
    gulp.series(gulp.parallel(schema, graphQLTypes), gulp.parallel(webpack, copyIntegrationAssets))
)

/**
 * Watches everything and rebuilds on file changes.
 */
export const watch = gulp.series(
    // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
    gulp.parallel(schema, graphQLTypes),
    gulp.parallel(watchSchema, watchGraphQLTypes, webpackDevServer, watchIntegrationAssets)
)
