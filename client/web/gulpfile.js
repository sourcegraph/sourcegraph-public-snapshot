const path = require('path')

require('ts-node').register({
  transpileOnly: true,
  // Use config with "module": "commonjs" because not all modules involved in tasks are esnext modules.
  project: path.resolve(__dirname, './dev/tsconfig.json'),
})

const compression = require('compression')
const gulp = require('gulp')
const { createProxyMiddleware } = require('http-proxy-middleware')
const signale = require('signale')
const createWebpackCompiler = require('webpack')
const WebpackDevServer = require('webpack-dev-server')
// The `DevServerPlugin` should be exposed after the `webpack-dev-server@4` goes out of the beta stage.
// eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
const DevServerPlugin = require('webpack-dev-server/lib/utils/DevServerPlugin')

const {
  graphQlSchema,
  graphQlOperations,
  schema,
  watchGraphQlSchema,
  watchGraphQlOperations,
  watchSchema,
  cssModulesTypings,
  watchCSSModulesTypings,
} = require('../shared/gulpfile')

const { build: buildEsbuild } = require('./dev/esbuild/build')
const { esbuildDevelopmentServer } = require('./dev/esbuild/server')
const {
  ENVIRONMENT_CONFIG,
  HTTPS_WEB_SERVER_URL,
  WEBPACK_STATS_OPTIONS,
  DEV_SERVER_LISTEN_ADDR,
  DEV_SERVER_PROXY_TARGET_ADDR,
  shouldCompressResponse,
  printSuccessBanner,
} = require('./dev/utils')
const webpackConfig = require('./webpack.config')

const { DEV_WEB_BUILDER, SOURCEGRAPH_HTTPS_DOMAIN, SOURCEGRAPH_HTTPS_PORT } = ENVIRONMENT_CONFIG

/**
 * @param {import('webpack').Stats} stats
 */
const logWebpackStats = stats => {
  signale.info(stats.toString(WEBPACK_STATS_OPTIONS))
}

function createWebApplicationCompiler() {
  signale.info('Building web application with the environment config', ENVIRONMENT_CONFIG)

  return createWebpackCompiler(webpackConfig)
}

async function webpack() {
  const compiler = createWebApplicationCompiler()
  /** @type {import('webpack').Stats} */
  const stats = await new Promise((resolve, reject) => {
    compiler.run((error, stats) => (error ? reject(error) : resolve(stats)))
  })
  logWebpackStats(stats)
  if (stats.hasErrors()) {
    throw Object.assign(new Error('Failed to compile'), { showStack: false })
  }
}

const webBuild = DEV_WEB_BUILDER === 'webpack' ? webpack : buildEsbuild

/**
 * Watch files and update the webpack bundle on disk without starting a dev server.
 */
async function watchWebpack() {
  const compiler = createWebApplicationCompiler()
  compiler.hooks.watchRun.tap('Notify', () => signale.info('Webpack compiling...'))
  await new Promise(() => {
    compiler.watch({ aggregateTimeout: 300 }, (error, stats) => {
      logWebpackStats(stats)
      if (error || stats.hasErrors()) {
        signale.error('Webpack compilation error')
      } else {
        signale.info('Webpack compilation done')
      }
    })
  })
}

async function webpackDevelopmentServer() {
  /** @type {import('webpack-dev-server').ProxyConfigMap } */
  const proxyConfig = {
    '/': {
      target: `http://${DEV_SERVER_PROXY_TARGET_ADDR.host}:${DEV_SERVER_PROXY_TARGET_ADDR.port}`,
      // Avoid crashing on "read ECONNRESET".
      onError: () => undefined,
      // Don't log proxy errors, these usually just contain
      // ECONNRESET errors caused by the browser cancelling
      // requests. This should not be needed to actually debug something.
      logLevel: 'silent',
      onProxyReqWs: (_proxyRequest, _request, socket) =>
        socket.on('error', error => console.error('WebSocket proxy error:', error)),
    },
  }

  /** @type {import('webpack-dev-server').Configuration} */
  const options = {
    // react-refresh plugin triggers page reload if needed.
    liveReload: false,
    hot: true,
    host: DEV_SERVER_LISTEN_ADDR.host,
    port: DEV_SERVER_LISTEN_ADDR.port,
    // Disable default DevServer compression. We need more fine grained compression to support streaming search.
    compress: false,
    onBeforeSetupMiddleware: developmentServer => {
      // Re-enable gzip compression using our own `compression` filter.
      developmentServer.app.use(compression({ filter: shouldCompressResponse }))
    },
    client: {
      overlay: false,
      webSocketTransport: 'ws',
      logging: 'verbose',
      webSocketURL: {
        hostname: SOURCEGRAPH_HTTPS_DOMAIN,
        port: SOURCEGRAPH_HTTPS_PORT,
        protocol: 'wss',
      },
    },
    static: {
      directory: './ui/assets',
      publicPath: '/.assets/',
    },
    proxy: proxyConfig,
    webSocketServer: 'ws',
  }

  // Based on the update: https://github.com/webpack/webpack-dev-server/pull/2844
  if (!webpackConfig.plugins.find(plugin => plugin.constructor === DevServerPlugin)) {
    // eslint-disable-next-line @typescript-eslint/no-unsafe-call
    webpackConfig.plugins.push(new DevServerPlugin(options))
  }

  const compiler = createWebApplicationCompiler()
  let compilationDoneOnce = false
  compiler.hooks.done.tap('Print external URL', stats => {
    stats = stats.toJson()
    if (stats.errors !== undefined && stats.errors.length > 0) {
      // show errors
      return
    }
    if (compilationDoneOnce) {
      return
    }
    compilationDoneOnce = true

    printSuccessBanner(['✱ Sourcegraph is really ready now!', `Click here: ${HTTPS_WEB_SERVER_URL}`])
  })

  const server = new WebpackDevServer(options, compiler)
  signale.await('Waiting for Webpack to compile assets')
  await server.start()
}

const esbuildDevelopmentProxy = () =>
  esbuildDevelopmentServer(DEV_SERVER_LISTEN_ADDR, app => {
    app.use(
      '/',
      createProxyMiddleware({
        target: {
          protocol: 'http:',
          host: DEV_SERVER_PROXY_TARGET_ADDR.host,
          port: DEV_SERVER_PROXY_TARGET_ADDR.port,
        },
        logLevel: 'error',
      })
    )
  })

const developmentServer = DEV_WEB_BUILDER === 'webpack' ? webpackDevelopmentServer : esbuildDevelopmentProxy

// Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
const generate = gulp.parallel(schema, graphQlSchema, graphQlOperations, cssModulesTypings)

// Watches code generation only, rebuilds on file changes
const watchGenerators = gulp.parallel(watchSchema, watchGraphQlSchema, watchGraphQlOperations, watchCSSModulesTypings)

/**
 * Builds everything.
 */
const build = gulp.series(generate, webBuild)

/**
 * Starts a development server without initial code generation, watches everything and rebuilds on file changes.
 */
const developmentWithoutInitialCodeGen = gulp.parallel(watchGenerators, developmentServer)

/**
 * Runs code generation first, then starts a development server, watches everything and rebuilds on file changes.
 */
const development = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  generate,
  developmentWithoutInitialCodeGen
)

/**
 * Watches everything, rebuilds on file changes and writes the bundle to disk.
 * Useful to running integration tests.
 */
const watch = gulp.series(
  // Ensure the typings that TypeScript depends on are build to avoid first-time-run errors
  generate,
  gulp.parallel(watchGenerators, watchWebpack)
)

module.exports = {
  build,
  watch,
  dev: development,
  unsafeDev: developmentWithoutInitialCodeGen,
  webpackDevServer: webpackDevelopmentServer,
  webpack,
  watchWebpack,
  webBuild,
  developmentServer,
  generate,
  watchGenerators,
}
