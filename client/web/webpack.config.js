// @ts-check

const path = require('path')

const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin')
const StatoscopeWebpackPlugin = require('@statoscope/webpack-plugin').default
const CompressionPlugin = require('compression-webpack-plugin')
const CssMinimizerWebpackPlugin = require('css-minimizer-webpack-plugin')
const logger = require('gulplog')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const webpack = require('webpack')
const { WebpackManifestPlugin } = require('webpack-manifest-plugin')

const {
  ROOT_PATH,
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
} = require('@sourcegraph/build-config')

const { getHTMLWebpackPlugins } = require('./dev/webpack/get-html-webpack-plugins')
const { isHotReloadEnabled } = require('./src/integration/environment')

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'
logger.info('Using mode', mode)

const isDevelopment = mode === 'development'
const isProduction = mode === 'production'
const isCI = process.env.CI === 'true'
const isCacheEnabled = isDevelopment && !isCI
const isEmbedDevelopment = isDevelopment && process.env.EMBED_DEVELOPMENT === 'true'

/** Allow overriding default Webpack naming behavior for debugging */
const useNamedChunks = process.env.WEBPACK_USE_NAMED_CHUNKS === 'true'

const devtool = isProduction ? 'source-map' : 'eval-cheap-module-source-map'

const shouldServeIndexHTML = process.env.WEBPACK_SERVE_INDEX === 'true'
if (shouldServeIndexHTML) {
  logger.info('Serving index.html with HTMLWebpackPlugin')
}
const webServerEnvironmentVariables = {
  WEBPACK_SERVE_INDEX: JSON.stringify(process.env.WEBPACK_SERVE_INDEX),
  SOURCEGRAPH_API_URL: JSON.stringify(process.env.SOURCEGRAPH_API_URL),
}

const shouldAnalyze = process.env.WEBPACK_ANALYZER === '1'

const STATOSCOPE_STATS = {
  all: false, // disable all the stats
  hash: true, // compilation hash
  entrypoints: true,
  chunks: true,
  chunkModules: true, // modules
  reasons: true, // modules reasons
  ids: true, // IDs of modules and chunks (webpack 5)
  dependentModules: true, // dependent modules of chunks (webpack 5)
  chunkRelations: true, // chunk parents, children and siblings (webpack 5)
  cachedAssets: true, // information about the cached assets (webpack 5)

  nestedModules: true, // concatenated modules
  usedExports: true,
  providedExports: true, // provided imports
  assets: true,
  chunkOrigins: true, // chunks origins stats (to find out which modules require a chunk)
  version: true, // webpack version
  builtAt: true, // build at time
  timings: true, // modules timing information
  performance: true, // info about oversized assets
}

if (shouldAnalyze) {
  logger.info('Running bundle analyzer')
}

const hotLoadablePaths = ['branded', 'shared', 'web', 'wildcard'].map(workspace =>
  path.resolve(ROOT_PATH, 'client', workspace, 'src')
)

const isEnterpriseBuild = process.env.ENTERPRISE && Boolean(JSON.parse(process.env.ENTERPRISE))
const isEmbedEntrypointEnabled = isEnterpriseBuild && (isProduction || isEmbedDevelopment)
const enterpriseDirectory = path.resolve(__dirname, 'src', 'enterprise')

const styleLoader = isDevelopment ? 'style-loader' : MiniCssExtractPlugin.loader

const extensionHostWorker = /main\.worker\.ts$/

/** @type {import('webpack').Configuration} */
const config = {
  context: __dirname, // needed when running `gulp webpackDevServer` from the root dir
  mode,
  stats: shouldAnalyze
    ? STATOSCOPE_STATS
    : {
        // Minimize logging in case if Webpack is used along with multiple other services.
        // Use `normal` output preset in case of running standalone web server.
        preset: shouldServeIndexHTML || isProduction ? 'normal' : 'errors-warnings',
        errorDetails: true,
        timings: true,
      },
  infrastructureLogging: {
    // Controls webpack-dev-server logging level.
    level: 'warn',
  },
  target: 'browserslist',
  // Use cache only in `development` mode to speed up production build.
  cache: isCacheEnabled && getCacheConfig({ invalidateCacheFiles: [path.resolve(__dirname, 'babel.config.js')] }),
  optimization: {
    minimize: isProduction,
    minimizer: [getTerserPlugin(), new CssMinimizerWebpackPlugin()],
    splitChunks: {
      cacheGroups: {
        react: {
          test: /[/\\]node_modules[/\\](react|react-dom)[/\\]/,
          name: 'react',
          chunks: 'all',
        },
      },
    },
    ...(isDevelopment && {
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
    app: isEnterpriseBuild ? path.join(enterpriseDirectory, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
    // Embedding entrypoint. It uses a small subset of the main webapp intended to be embedded into
    // iframes on 3rd party sites. Added only in production enterprise builds or if embed development is enabled.
    ...(isEmbedEntrypointEnabled && { embed: path.join(enterpriseDirectory, 'embed', 'main.tsx') }),
  },
  output: {
    path: path.join(ROOT_PATH, 'ui', 'assets'),
    // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
    // Note: [name] will vary depending on the Webpack chunk. If specified, it will use a provided chunk name, otherwise it will fallback to a deterministic id.
    filename:
      mode === 'production' && !useNamedChunks ? 'scripts/[name].[contenthash].bundle.js' : 'scripts/[name].bundle.js',
    chunkFilename:
      mode === 'production' && !useNamedChunks ? 'scripts/[name]-[contenthash].chunk.js' : 'scripts/[name].chunk.js',
    publicPath: '/.assets/',
    globalObject: 'self',
    pathinfo: false,
  },
  devtool,
  plugins: [
    // Needed for React
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify(mode),
        ...(shouldServeIndexHTML && webServerEnvironmentVariables),
      },
    }),
    getProvidePlugin(),
    new MiniCssExtractPlugin({
      // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
      filename: mode === 'production' ? 'styles/[name].[contenthash].bundle.css' : 'styles/[name].bundle.css',
    }),
    getMonacoWebpackPlugin(),
    !shouldServeIndexHTML &&
      new WebpackManifestPlugin({
        writeToFileEmit: true,
        fileName: 'webpack.manifest.json',
        // Only output files that are required to run the application.
        filter: ({ isInitial, name }) => isInitial || name?.includes('react'),
      }),
    ...(shouldServeIndexHTML ? getHTMLWebpackPlugins() : []),
    shouldAnalyze && new StatoscopeWebpackPlugin(),
    isHotReloadEnabled && new webpack.HotModuleReplacementPlugin(),
    isHotReloadEnabled && new ReactRefreshWebpackPlugin({ overlay: false }),
    isProduction &&
      new CompressionPlugin({
        filename: '[path][base].gz',
        algorithm: 'gzip',
        test: /\.(js|css|svg)$/,
        compressionOptions: {
          /** Maximum compression level for Gzip */
          level: 9,
        },
      }),
    isProduction &&
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
          ...(isProduction ? ['thread-loader'] : []),
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: true,
              ...(isHotReloadEnabled && { plugins: ['react-refresh/babel'] }),
            },
          },
        ],
      },
      {
        test: /\.[jt]sx?$/,
        exclude: [...hotLoadablePaths, extensionHostWorker],
        use: [...(isProduction ? ['thread-loader'] : []), getBabelLoader()],
      },
      {
        test: /\.(sass|scss)$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(styleLoader, getCSSModulesLoader({ sourceMap: isDevelopment })),
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
    ],
  },
}

module.exports = config
