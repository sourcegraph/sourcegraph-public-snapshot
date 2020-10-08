// @ts-check

const log = require('fancy-log')
const gulp = require('gulp')
const createWebpackCompiler = require('webpack')
const WebpackDevServer = require('webpack-dev-server')
const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
} = require('../shared/gulpfile')
const webpackConfig = require('./webpack.config')

const WEBPACK_STATS_OPTIONS = {
  all: false,
  timings: true,
  errors: true,
  warnings: true,
  colors: true,
}

/**
 * @param {import('webpack').Stats} stats
 */
const logWebpackStats = stats => {
  log(stats.toString(WEBPACK_STATS_OPTIONS))
}

async function webpack() {
  const compiler = createWebpackCompiler(webpackConfig)
  /** @type {import('webpack').Stats} */
  const stats = await new Promise((resolve, reject) => {
    compiler.run((error, stats) => (error ? reject(error) : resolve(stats)))
  })
  logWebpackStats(stats)
  if (stats.hasErrors()) {
    throw Object.assign(new Error('Failed to compile'), { showStack: false })
  }
}

async function webpackDevelopmentServer() {
  const sockHost = process.env.SOURCEGRAPH_HTTPS_DOMAIN || 'sourcegraph.test'
  const sockPort = Number(process.env.SOURCEGRAPH_HTTPS_PORT || 3443)

  /** @type {import('webpack-dev-server').Configuration & { liveReload?: boolean }} */
  const options = {
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
        onError: error => console.error(error),
        onProxyReqWs: (_proxyRequest, _request, socket) =>
          socket.on('error', error => console.error('WebSocket proxy error:', error)),
      },
    },
    sockHost,
    sockPort,
  }
  WebpackDevServer.addDevServerEntrypoints(webpackConfig, options)
  const server = new WebpackDevServer(createWebpackCompiler(webpackConfig), options)
  await new Promise((resolve, reject) => {
    server.listen(3080, '0.0.0.0', error => (error ? reject(error) : resolve()))
  })
}

/**
 * Builds everything.
 */
const build = gulp.series(gulp.parallel(schema, graphQlOperations, graphQlSchema), gulp.parallel(webpack))

/**
 * Watches everything and rebuilds on file changes.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  gulp.parallel(schema, graphQlSchema),
  gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations, webpackDevelopmentServer)
)

module.exports = { build, watch, webpackDevServer: webpackDevelopmentServer, webpack }
