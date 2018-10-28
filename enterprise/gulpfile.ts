import log from 'fancy-log'
import gulp from 'gulp'
import httpProxyMiddleware from 'http-proxy-middleware'
// @ts-ignore
import convert from 'koa-connect'
import createWebpackCompiler, { Stats } from 'webpack'
import serve from 'webpack-serve'
import webpackConfig from './webpack.config'

const PHABRICATOR_EXTENSION_FILES = './node_modules/@sourcegraph/phabricator-extension/dist/**'

/**
 * Copies the bundles from the `@sourcegraph/phabricator-extension` package over to the ui/assets
 * folder so they can be served by the webapp.
 * The package is published from https://github.com/sourcegraph/browser-extensions
 */
export function phabricator(): NodeJS.ReadWriteStream {
    return gulp.src(PHABRICATOR_EXTENSION_FILES).pipe(gulp.dest('./ui/assets/extension'))
}

export const watchPhabricator = gulp.series(phabricator, async function watchPhabricator(): Promise<void> {
    await new Promise<never>((_, reject) => {
        gulp.watch(PHABRICATOR_EXTENSION_FILES, phabricator).on('error', reject)
    })
})

const WEBPACK_STATS_OPTIONS = {
    all: false,
    timings: true,
    errors: true,
    warnings: true,
    warningsFilter: warning =>
        // This is intended, so ignore warning
        /node_modules\/monaco-editor\/.*\/editorSimpleWorker.js.*\n.*dependency is an expression/.test(warning),
    colors: true,
} as Stats.ToStringOptions
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

export async function webpackServe(): Promise<void> {
    await serve(
        {},
        {
            config: {
                ...webpackConfig,
                serve: {
                    clipboard: false,
                    content: './ui/assets',
                    port: 3080,
                    hotClient: false,
                    devMiddleware: {
                        publicPath: '/.assets/',
                        stats: WEBPACK_STATS_OPTIONS,
                    },
                    add: (app, middleware) => {
                        // Since we're manipulating the order of middleware added, we need to handle adding these
                        // two internal middleware functions.
                        //
                        // The `as any` cast is necessary because the `middleware.webpack` typings are incorrect
                        // (the related issue https://github.com/webpack-contrib/webpack-serve/issues/238 perhaps
                        // explains why: the webpack-serve docs incorrectly state that resolving
                        // `middleware.webpack()` is not necessary).
                        ;(middleware.webpack() as any).then(() => {
                            middleware.content()

                            // Proxy *must* be the last middleware added.
                            app.use(
                                convert(
                                    // Proxy all requests (that are not for webpack-built assets) to the Sourcegraph
                                    // frontend server, and we make the Sourcegraph appURL equal to the URL of
                                    // webpack-serve. This is how webpack-serve needs to work (because it does a bit
                                    // more magic in injecting scripts that use WebSockets into proxied requests).
                                    httpProxyMiddleware({
                                        target: 'http://localhost:3081',
                                        ws: true,

                                        // Avoid crashing on "read ECONNRESET".
                                        onError: err => console.error(err),
                                        onProxyReqWs: (_proxyReq, _req, socket) =>
                                            socket.on('error', err => console.error('WebSocket proxy error:', err)),
                                    })
                                )
                            )
                        })
                    },
                    compiler: createWebpackCompiler(webpackConfig),
                },
            },
        }
    )
}

export const build = gulp.series(gulp.parallel(phabricator, webpack))

export const watch = gulp.parallel(watchPhabricator, webpackServe)
