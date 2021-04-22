// @ts-check

const path = require('path')

const CssMinimizerWebpackPlugin = require('css-minimizer-webpack-plugin')
const logger = require('gulplog')
const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')
const TerserPlugin = require('terser-webpack-plugin')
const webpack = require('webpack')
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
const { WebpackManifestPlugin } = require('webpack-manifest-plugin')

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'
logger.info('Using mode', mode)

const isDevelopment = mode === 'development'
const isProduction = mode === 'production'
const devtool = isProduction ? 'source-map' : 'cheap-module-eval-source-map'

const shouldAnalyze = process.env.WEBPACK_ANALYZER === '1'
if (shouldAnalyze) {
  logger.info('Running bundle analyzer')
}

const rootPath = path.resolve(__dirname, '..', '..')
const nodeModulesPath = path.resolve(rootPath, 'node_modules')
const monacoEditorPaths = [path.resolve(nodeModulesPath, 'monaco-editor')]

const isEnterpriseBuild = !!process.env.ENTERPRISE
const enterpriseDirectory = path.resolve(__dirname, 'src', 'enterprise')

const babelLoader = {
  loader: 'babel-loader',
  options: {
    cacheDirectory: true,
    configFile: path.join(__dirname, 'babel.config.js'),
  },
}

const extensionHostWorker = /main\.worker\.ts$/

/**
 * Generates array of CSS loaders both for regular CSS and CSS modules.
 * Useful to ensure that we use the same configuration for shared loaders: postcss-loader, sass-loader, etc.
 *
 * @param {import('webpack').RuleSetUseItem[]} loaders additional CSS loaders
 * @returns {import('webpack').RuleSetUseItem[]} array of CSS loaders
 */
const getCSSLoaders = (...loaders) => [
  // Use style-loader for local development as it is significantly faster.
  isDevelopment ? 'style-loader' : MiniCssExtractPlugin.loader,
  ...loaders,
  'postcss-loader',
  {
    loader: 'sass-loader',
    options: {
      sassOptions: {
        implementation: require('sass'),
        includePaths: [nodeModulesPath, path.resolve(rootPath, 'client')],
      },
    },
  },
]

/** @type {import('webpack').Configuration} */
const config = {
  context: __dirname, // needed when running `gulp webpackDevServer` from the root dir
  mode,
  optimization: {
    minimize: isProduction,
    minimizer: [
      new TerserPlugin({
        sourceMap: true,
        terserOptions: {
          compress: {
            // // Don't inline functions, which causes name collisions with uglify-es:
            // https://github.com/mishoo/UglifyJS2/issues/2842
            inline: 1,
          },
        },
      }),
      new CssMinimizerWebpackPlugin(),
    ],
    namedModules: false,

    ...(isDevelopment
      ? {
          removeAvailableModules: false,
          removeEmptyChunks: false,
          splitChunks: false,
        }
      : {}),
  },
  entry: {
    // Enterprise vs. OSS builds use different entrypoints. The enterprise entrypoint imports a
    // strict superset of the OSS entrypoint.
    app: [
      'react-hot-loader/patch',
      isEnterpriseBuild ? path.join(enterpriseDirectory, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
    ],

    'editor.worker': 'monaco-editor/esm/vs/editor/editor.worker.js',
    'json.worker': 'monaco-editor/esm/vs/language/json/json.worker',
  },
  output: {
    path: path.join(rootPath, 'ui', 'assets'),
    // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
    filename: mode === 'production' ? 'scripts/[name].[contenthash].bundle.js' : 'scripts/[name].bundle.js',
    chunkFilename: mode === 'production' ? 'scripts/[id]-[contenthash].chunk.js' : 'scripts/[id].chunk.js',
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
      },
    }),
    new MiniCssExtractPlugin({
      // Do not [hash] for development -- see https://github.com/webpack/webpack-dev-server/issues/377#issuecomment-241258405
      filename: mode === 'production' ? 'styles/[name].[contenthash].bundle.css' : 'styles/[name].bundle.css',
    }),
    new MonacoWebpackPlugin({
      languages: ['json'],
      features: [
        'bracketMatching',
        'clipboard',
        'coreCommands',
        'cursorUndo',
        'find',
        'format',
        'hover',
        'inPlaceReplace',
        'iPadShowKeyboard',
        'links',
        'suggest',
      ],
    }),
    new webpack.IgnorePlugin(/\.flow$/, /.*/),
    new WebpackManifestPlugin({
      writeToFileEmit: true,
      fileName: 'webpack.manifest.json',
      // Only output files that are required to run the application
      filter: ({ isInitial }) => isInitial,
    }),
    ...(shouldAnalyze ? [new BundleAnalyzerPlugin()] : []),
  ],
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js'],
    mainFields: ['es2015', 'module', 'browser', 'main'],
    alias: {
      // react-visibility-sensor's main field points to a UMD bundle instead of ESM
      // https://github.com/joshwnj/react-visibility-sensor/issues/148
      'react-visibility-sensor': path.resolve(rootPath, 'node_modules/react-visibility-sensor/visibility-sensor.js'),
    },
  },
  module: {
    rules: [
      // Run hot-loading-related Babel plugins on our application code only (because they'd be
      // slow to run on all JavaScript code).
      {
        test: /\.[jt]sx?$/,
        include: path.join(__dirname, 'src'),
        exclude: extensionHostWorker,
        use: [
          ...(isProduction ? ['thread-loader'] : []),
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: true,
              plugins: [
                'react-hot-loader/babel',
                [
                  '@sourcegraph/babel-plugin-transform-react-hot-loader-wrapper',
                  {
                    modulePattern: 'web/src/.*\\.tsx$',
                    componentNamePattern: '(Page|Area)$',
                  },
                ],
              ],
            },
          },
        ],
      },
      {
        test: /\.[jt]sx?$/,
        exclude: [path.join(__dirname, 'src'), extensionHostWorker],
        use: [...(isProduction ? ['thread-loader'] : []), babelLoader],
      },
      {
        test: /\.(sass|scss)$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders({
          loader: 'css-loader',
          options: {
            sourceMap: isDevelopment,
            url: false,
            modules: {
              exportLocalsConvention: 'camelCase',
              localIdentName: '[name]__[local]_[hash:base64:5]',
            },
          },
        }),
      },
      {
        test: /\.(sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders({ loader: 'css-loader', options: { url: false } }),
      },
      {
        // CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
        test: /\.css$/,
        include: monacoEditorPaths,
        use: ['style-loader', { loader: 'css-loader', options: { url: false } }],
      },
      {
        test: extensionHostWorker,
        use: [{ loader: 'worker-loader', options: { inline: 'no-fallback' } }, babelLoader],
      },
      { test: /\.ya?ml$/, use: ['raw-loader'] },
    ],
  },
}

module.exports = config
