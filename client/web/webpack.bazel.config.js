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
  getCacheConfig,
  getMonacoWebpackPlugin,
  getBazelCSSLoaders: getCSSLoaders,
  getTerserPlugin,
  getProvidePlugin,
  getCSSModulesLoader,
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

const IS_PERSISTENT_CACHE_ENABLED = false // Disabled in Bazel
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

// Third-party npm dependencies that contain jsx syntax
const NPM_JSX = ['react-visibility-sensor/visibility-sensor.js']

const enterpriseDirectory = path.resolve(__dirname, 'src', 'enterprise')
const styleLoader = IS_DEVELOPMENT ? 'style-loader' : MiniCssExtractPlugin.loader

// Used to ensure that we include all initial chunks into the Webpack manifest.
const initialChunkNames = {
  react: 'react',
  opentelemetry: 'opentelemetry',
}

/** @type {import('webpack').Configuration} */
const config = {
  context: process.cwd(),
  mode: IS_PRODUCTION ? 'production' : 'development',
  stats: {
    // Minimize logging in case if Webpack is used along with multiple other services.
    // Use `normal` output preset in case of running standalone web server.
    preset: WEBPACK_SERVE_INDEX || IS_PRODUCTION ? 'normal' : 'errors',
    errorDetails: true,
    timings: true,
  },
  infrastructureLogging: {
    // Controls webpack-dev-server logging level.
    level: 'warn',
  },
  target: 'browserslist',
  // Use cache only in `development` mode to speed up production build.
  cache: IS_PERSISTENT_CACHE_ENABLED && getCacheConfig({ invalidateCacheFiles: [] }),
  performance: {
    hints: false,
  },
  optimization: {
    minimize: IS_PRODUCTION,
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
  // entry: { ... SET BY BAZEL RULE ... }
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
    // Change scss imports to the pre-compiled css files
    new webpack.NormalModuleReplacementPlugin(/.*\.scss$/, resource => {
      resource.request = resource.request.replace(/\.scss$/, '.css')
    }),

    new webpack.DefinePlugin({
      'process.env': mapValues(RUNTIME_ENV_VARIABLES, JSON.stringify),
    }),
    getProvidePlugin(),
    new MiniCssExtractPlugin({
      ignoreOrder: true,
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
    extensions: ['.mjs', '.jsx', '.js', '.json'],
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
      'react-visibility-sensor': require.resolve('react-visibility-sensor').replace('/dist/', '/'),
    },
  },
  module: {
    rules: [
      // Some third-party deps need jsx transpiled still
      {
        test: new RegExp('node_modules/' + NPM_JSX.join('|')),
        use: [{ loader: require.resolve('babel-loader') }],
      },
      {
        test: /\.m?js$/,
        type: 'javascript/auto',
        resolve: {
          // Allow importing without file extensions
          // https://webpack.js.org/configuration/module/#resolvefullyspecified
          fullySpecified: false,
        },
      },
      // TODO(bazel): why is this required when it is supposedly enabled by default?
      {
        test: /\.json$/,
        type: 'json',
      },
      {
        test: /\.css$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.css$/,
        use: getCSSLoaders(styleLoader, getCSSModulesLoader({ sourceMap: IS_DEVELOPMENT })),
      },
      {
        test: /\.css$/,
        exclude: /\.module\.css$/,
        use: getCSSLoaders(styleLoader, getBasicCSSLoader()),
      },
      getMonacoTTFRule(),
      { test: /\.ya?ml$/, type: 'asset/source' },
      { test: /\.(png|woff2)$/, type: 'asset/resource' },
    ],
  },
}

module.exports = config
