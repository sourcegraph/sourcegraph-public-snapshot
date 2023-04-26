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
const { StatsWriterPlugin } = require('webpack-stats-plugin')

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

const { IS_PRODUCTION, IS_DEVELOPMENT, ENVIRONMENT_CONFIG, writeIndexHTMLPlugin } = require('./dev/utils')
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
  WEBPACK_BUNDLE_ANALYZER,
  WEBPACK_EXPORT_STATS,
  WEBPACK_SERVE_INDEX,
  WEBPACK_STATS_NAME,
  WEBPACK_USE_NAMED_CHUNKS,
  WEBPACK_DEVELOPMENT_DEVTOOL,
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

// Used to ensure that we include all initial chunks into the Webpack manifest.
const initialChunkNames = {
  react: 'react',
  opentelemetry: 'opentelemetry',
}

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
    minimize: IS_PRODUCTION && !INTEGRATION_TESTS,
    minimizer: [getTerserPlugin(), new CssMinimizerWebpackPlugin()],
    splitChunks: {
      cacheGroups: {
        [initialChunkNames.react]: {
          test: /[/\\]node_modules[/\\](react|react-dom)[/\\]/,
          name: initialChunkNames.react,
          chunks: 'all',
        },
        [initialChunkNames.opentelemetry]: {
          test: /[/\\]node_modules[/\\](@opentelemetry|zone.js)[/\\]/,
          name: initialChunkNames.opentelemetry,
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
  // Inline source maps for integration tests to preserve readable stack traces.
  // See related issue here: https://github.com/puppeteer/puppeteer/issues/985
  devtool: IS_PRODUCTION ? (INTEGRATION_TESTS ? 'inline-source-map' : 'source-map') : WEBPACK_DEVELOPMENT_DEVTOOL,
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
    new WebpackManifestPlugin({
      writeToFileEmit: true,
      fileName: 'webpack.manifest.json',
      seed: {
        environment: NODE_ENV,
      },
      // Only output files that are required to run the application.
      filter: ({ isInitial, name }) =>
        isInitial || Object.values(initialChunkNames).some(initialChunkName => name?.includes(initialChunkName)),
    }),
    ...(WEBPACK_SERVE_INDEX && IS_PRODUCTION ? [writeIndexHTMLPlugin] : []),
    WEBPACK_BUNDLE_ANALYZER && getStatoscopePlugin(WEBPACK_STATS_NAME),
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
    WEBPACK_EXPORT_STATS &&
      new StatsWriterPlugin({
        filename: `stats-${process.env.BUILDKITE_COMMIT || 'unknown-commit'}.json`,
        stats: {
          all: false, // disable all the stats
          hash: true, // compilation hash
          entrypoints: true,
          chunks: true,
          chunkModules: true, // modules
          ids: true, // IDs of modules and chunks (webpack 5)
          cachedAssets: true, // information about the cached assets (webpack 5)
          nestedModules: true, // concatenated modules
          usedExports: true,
          assets: true,
          chunkOrigins: true, // chunks origins stats (to find out which modules require a chunk)
          timings: true, // modules timing information
          performance: true, // info about oversized assets
        },
      }),
  ].filter(Boolean),
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js', '.json'],
    mainFields: ['es2015', 'module', 'browser', 'main'],
    fallback: {
      path: require.resolve('path-browserify'),
      punycode: require.resolve('punycode'),
      util: require.resolve('util'),
      events: require.resolve('events'),
    },
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
        exclude: hotLoadablePaths,
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
      {
        test: /\.css$/,
        include: [path.resolve(__dirname, '../cody-ui')],
        exclude: /\.module\.css$/,
        use: getCSSLoaders(styleLoader, getBasicCSSLoader()),
      },
      {
        test: /\.module\.css$/,
        include: [path.resolve(__dirname, '../cody-ui')],
        use: getCSSLoaders(styleLoader, getCSSModulesLoader({ sourceMap: IS_DEVELOPMENT })),
      },
      getMonacoCSSRule(),
      getMonacoTTFRule(),
      { test: /\.ya?ml$/, type: 'asset/source' },
      { test: /\.(png|woff2)$/, type: 'asset/resource' },
    ],
  },
}

module.exports = config
