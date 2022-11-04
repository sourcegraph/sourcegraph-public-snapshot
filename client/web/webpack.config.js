// @ts-check

const path = require('path')

const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin')
const SentryWebpackPlugin = require('@sentry/webpack-plugin')
const CompressionPlugin = require('compression-webpack-plugin')
const CssMinimizerWebpackPlugin = require('css-minimizer-webpack-plugin')
const mapValues = require('lodash/mapValues')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const webpack = require('webpack')
const { WebpackManifestPlugin } = require('webpack-manifest-plugin')

const {
  ROOT_PATH,
  STATIC_ASSETS_PATH,
  getBabelLoader,
  getCacheConfig,
  getMonacoWebpackPlugin,
  getCSSLoaders,
  getTerserPlugin,
  getProvidePlugin,
  getCSSModulesLoader,
  getMonacoCSSRule,
  getMonacoTTFRule,
  getBasicCSSLoader,
  getStatoscopePlugin,
} = require('@sourcegraph/build-config')

const { IS_PRODUCTION, IS_DEVELOPMENT, ENVIRONMENT_CONFIG } = require('./dev/utils')
const { getHTMLWebpackPlugins } = require('./dev/webpack/get-html-webpack-plugins')
const { isHotReloadEnabled } = require('./src/integration/environment')

const {
  NODE_ENV,
  CI: IS_CI,
  INTEGRATION_TESTS,
  ENTERPRISE,
  EMBED_DEVELOPMENT,
  ENABLE_SENTRY,
  ENABLE_OPEN_TELEMETRY,
  SOURCEGRAPH_API_URL,
  WEBPACK_SERVE_INDEX,
  WEBPACK_BUNDLE_ANALYZER,
  WEBPACK_STATS_NAME,
  WEBPACK_USE_NAMED_CHUNKS,
  SENTRY_UPLOAD_SOURCE_MAPS,
  COMMIT_SHA,
  VERSION,
  SENTRY_DOT_COM_AUTH_TOKEN,
  SENTRY_ORGANIZATION,
  SENTRY_PROJECT,
} = ENVIRONMENT_CONFIG

const IS_PERSISTENT_CACHE_ENABLED = IS_DEVELOPMENT && !IS_CI
const IS_EMBED_ENTRY_POINT_ENABLED = ENTERPRISE && (IS_PRODUCTION || (IS_DEVELOPMENT && EMBED_DEVELOPMENT))

const RUNTIME_ENV_VARIABLES = {
  NODE_ENV,
  ENABLE_SENTRY,
  ENABLE_OPEN_TELEMETRY,
  INTEGRATION_TESTS,
  COMMIT_SHA,
  ...(WEBPACK_SERVE_INDEX && { SOURCEGRAPH_API_URL }),
}

const hotLoadablePaths = ['branded', 'shared', 'web', 'wildcard'].map(workspace =>
  path.resolve(ROOT_PATH, 'client', workspace, 'src')
)

const enterpriseDirectory = path.resolve(__dirname, 'src', 'enterprise')
const styleLoader = IS_DEVELOPMENT ? 'style-loader' : MiniCssExtractPlugin.loader

const extensionHostWorker = /main\.worker\.ts$/

/** @type {import('webpack').Configuration} */
const config = {
  context: __dirname, // needed when running `gulp webpackDevServer` from the root dir
  mode: IS_PRODUCTION ? 'production' : 'development',
  stats: {
    // Minimize logging in case if Webpack is used along with multiple other services.
    // Use `normal` output preset in case of running standalone web server.
    preset: WEBPACK_SERVE_INDEX || IS_PRODUCTION ? 'normal' : 'errors-warnings',
    errorDetails: true,
    timings: true,
  },
  infrastructureLogging: {
    // Controls webpack-dev-server logging level.
    level: 'warn',
  },
  target: 'browserslist',
  // Use cache only in `development` mode to speed up production build.
  cache:
    IS_PERSISTENT_CACHE_ENABLED &&
    getCacheConfig({ invalidateCacheFiles: [path.resolve(__dirname, 'babel.config.js')] }),
  optimization: {
    minimize: IS_PRODUCTION,
    minimizer: [getTerserPlugin(), new CssMinimizerWebpackPlugin()],
    splitChunks: {
      cacheGroups: {
        react: {
          test: /[/\\]node_modules[/\\](react|react-dom)[/\\]/,
          name: 'react',
          chunks: 'all',
        },
        opentelemetry: {
          test: /[/\\]node_modules[/\\](@opentelemetry)[/\\]/,
          name: 'opentelemetry',
          chunks: 'all',
        },
      },
    },
    ...(IS_DEVELOPMENT && {
      // Running multiple entries on a single page that do not share a runtime chunk from the same compilation is not supported.
      // https://github.com/webpack/webpack-dev-server/issues/2792#issuecomment-808328432
      runtimeChunk: isHotReloadEnabled ? 'single' : false,
      removeAvailableModules: false,
      removeEmptyChunks: false,
      splitChunks: false,
    }),
  },
  entry: {
    // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
    // strict superset of the OSS entrypoint.
    app: ENTERPRISE ? path.join(enterpriseDirectory, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
    // Embedding entrypoint. It uses a small subset of the main webapp intended to be embedded into
    // iframes on 3rd party sites. Added only in production enterprise builds or if embed development is enabled.
    ...(IS_EMBED_ENTRY_POINT_ENABLED && { embed: path.join(enterpriseDirectory, 'embed', 'main.tsx') }),
  },
  output: {
    path: STATIC_ASSETS_PATH,
    // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
    // Note: [name] will vary depending on the Webpack chunk. If specified, it will use a provided chunk name, otherwise it will fallback to a deterministic id.
    filename:
      IS_PRODUCTION && !WEBPACK_USE_NAMED_CHUNKS
        ? 'scripts/[name].[contenthash].bundle.js'
        : 'scripts/[name].bundle.js',
    chunkFilename:
      IS_PRODUCTION && !WEBPACK_USE_NAMED_CHUNKS ? 'scripts/[name]-[contenthash].chunk.js' : 'scripts/[name].chunk.js',
    publicPath: '/.assets/',
    globalObject: 'self',
    pathinfo: false,
  },
  devtool: IS_PRODUCTION ? 'source-map' : 'eval-cheap-module-source-map',
  plugins: [
    new webpack.DefinePlugin({
      'process.env': mapValues(RUNTIME_ENV_VARIABLES, JSON.stringify),
    }),
    getProvidePlugin(),
    new MiniCssExtractPlugin({
      // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
      filename:
        IS_PRODUCTION && !WEBPACK_USE_NAMED_CHUNKS
          ? 'styles/[name].[contenthash].bundle.css'
          : 'styles/[name].bundle.css',
    }),
    getMonacoWebpackPlugin(),
    !WEBPACK_SERVE_INDEX &&
      new WebpackManifestPlugin({
        writeToFileEmit: true,
        fileName: 'webpack.manifest.json',
        // Only output files that are required to run the application.
        filter: ({ isInitial, name }) => isInitial || name?.includes('react'),
      }),
    ...(WEBPACK_SERVE_INDEX ? getHTMLWebpackPlugins() : []),
    WEBPACK_BUNDLE_ANALYZER && getStatoscopePlugin(WEBPACK_STATS_NAME),
    isHotReloadEnabled && new webpack.HotModuleReplacementPlugin(),
    isHotReloadEnabled && new ReactRefreshWebpackPlugin({ overlay: false }),
    IS_PRODUCTION &&
      new CompressionPlugin({
        filename: '[path][base].gz',
        algorithm: 'gzip',
        test: /\.(js|css|svg)$/,
        compressionOptions: {
          /** Maximum compression level for Gzip */
          level: 9,
        },
      }),
    IS_PRODUCTION &&
      new CompressionPlugin({
        filename: '[path][base].br',
        algorithm: 'brotliCompress',
        test: /\.(js|css|svg)$/,
        compressionOptions: {
          /** Maximum compression level for Brotli */
          level: 11,
        },
        /**
         * We get little/no benefits from compressing files that are already under this size.
         * We can fall back to dynamic gzip for these.
         */
        threshold: 10240,
      }),
    VERSION &&
      SENTRY_UPLOAD_SOURCE_MAPS &&
      new SentryWebpackPlugin({
        silent: true,
        org: SENTRY_ORGANIZATION,
        project: SENTRY_PROJECT,
        authToken: SENTRY_DOT_COM_AUTH_TOKEN,
        release: `frontend@${VERSION}`,
        include: path.join(STATIC_ASSETS_PATH, 'scripts', '*.map'),
      }),
  ].filter(Boolean),
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js', '.json'],
    mainFields: ['es2015', 'module', 'browser', 'main'],
    alias: {
      // react-visibility-sensor's main field points to a UMD bundle instead of ESM
      // https://github.com/joshwnj/react-visibility-sensor/issues/148
      'react-visibility-sensor': path.resolve(ROOT_PATH, 'node_modules/react-visibility-sensor/visibility-sensor.js'),
    },
  },
  module: {
    rules: [
      // Run hot-loading-related Babel plugins on our application code only (because they'd be
      // slow to run on all JavaScript code).
      {
        test: /\.[jt]sx?$/,
        include: hotLoadablePaths,
        exclude: extensionHostWorker,
        use: [
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: !IS_PRODUCTION,
              ...(isHotReloadEnabled && { plugins: ['react-refresh/babel'] }),
            },
          },
        ],
      },
      {
        test: /\.[jt]sx?$/,
        exclude: [...hotLoadablePaths, extensionHostWorker],
        use: [getBabelLoader()],
      },
      {
        test: /\.(sass|scss)$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(styleLoader, getCSSModulesLoader({ sourceMap: IS_DEVELOPMENT })),
      },
      {
        test: /\.(sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(styleLoader, getBasicCSSLoader()),
      },
      getMonacoCSSRule(),
      getMonacoTTFRule(),
      {
        test: extensionHostWorker,
        use: [{ loader: 'worker-loader', options: { inline: 'no-fallback' } }, getBabelLoader()],
      },
      { test: /\.ya?ml$/, type: 'asset/source' },
      { test: /\.(png|woff2)$/, type: 'asset/resource' },
      {
        test: /\.elm$/,
        exclude: /elm-stuff/,
        use: {
          loader: 'elm-webpack-loader',
          options: {
            cwd: path.resolve(ROOT_PATH, 'client/web/src/search/results/components/compute'),
            report: 'json',
            pathToElm: path.resolve(ROOT_PATH, 'node_modules/.bin/elm'),
          },
        },
      },
    ],
  },
}

module.exports = config
