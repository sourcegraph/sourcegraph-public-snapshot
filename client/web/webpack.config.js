// @ts-check

const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin')
const OptimizeCssAssetsPlugin = require('optimize-css-assets-webpack-plugin')
const path = require('path')
const webpack = require('webpack')
const logger = require('gulplog')
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
const { ESBuildPlugin, ESBuildMinifyPlugin } = require('esbuild-loader')

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'
logger.info('Using mode', mode)

const devtool = mode === 'production' ? 'source-map' : 'eval'

const shouldAnalyze = process.env.WEBPACK_ANALYZER === '1'
if (shouldAnalyze) {
  logger.info('Running bundle analyzer')
}

const rootDirectory = path.resolve(__dirname, '..', '..')
const nodeModulesPath = path.resolve(rootDirectory, 'node_modules')
const monacoEditorPaths = [path.resolve(nodeModulesPath, 'monaco-editor')]

const isEnterpriseBuild = !!process.env.ENTERPRISE
const enterpriseDirectory = path.resolve(__dirname, 'src', 'enterprise')

const extensionHostWorker = /main\.worker\.ts$/

const esbuildTarget = ['chrome87', 'chrome86', 'edge87', 'firefox84', 'firefox83', 'safari14', 'safari13.1', 'safari13']

/** @type {import('webpack').Configuration} */
const config = {
  context: __dirname, // needed when running `gulp webpackDevServer` from the root dir
  mode,
  optimization: {
    minimize: mode === 'production',
    minimizer: [new ESBuildMinifyPlugin()],
    namedModules: false,
    ...(mode === 'development'
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
    app: isEnterpriseBuild ? path.join(enterpriseDirectory, 'main.tsx') : path.join(__dirname, 'src', 'main.tsx'),
  },
  output: {
    path: path.join(rootDirectory, 'ui', 'assets'),
    filename: 'scripts/[name].bundle.js',
    chunkFilename: 'scripts/[id]-[contenthash].chunk.js',
    publicPath: '/.assets/',
    globalObject: 'self',
    pathinfo: false,
  },
  devtool,
  plugins: [
    new ESBuildPlugin(),
    // Needed for React
    new webpack.DefinePlugin({
      'process.env': {
        NODE_ENV: JSON.stringify(mode),
      },
    }),
    new MiniCssExtractPlugin({ filename: 'styles/[name].bundle.css' }),
    new OptimizeCssAssetsPlugin(),
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
    ...(shouldAnalyze ? [new BundleAnalyzerPlugin()] : []),
  ],
  resolve: {
    extensions: ['.mjs', '.ts', '.tsx', '.js'],
    mainFields: ['es2015', 'module', 'browser', 'main'],
    // alias: {
    //   // react-visibility-sensor's main field points to a UMD bundle instead of ESM
    //   // https://github.com/joshwnj/react-visibility-sensor/issues/148
    //   'react-visibility-sensor': path.resolve(
    //     rootDirectory,
    //     'node_modules/react-visibility-sensor/visibility-sensor.js'
    //   ),
    // },
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        include: path.resolve(__dirname, '..'),
        exclude: extensionHostWorker,
        use: [
          {
            loader: 'esbuild-loader',
            options: {
              loader: 'tsx',
              target: esbuildTarget,
            },
          },
        ],
      },
      {
        test: /\.(sass|scss)$/,
        include: path.resolve(__dirname, 'src'),
        use: [
          mode === 'production' ? MiniCssExtractPlugin.loader : 'style-loader',
          'css-loader',
          'postcss-loader',
          {
            loader: 'sass-loader',
            options: {
              sassOptions: {
                implementation: require('sass'),
              },
            },
          },
        ],
      },
      {
        // CSS rule for monaco-editor and other external plain CSS (skip SASS and PostCSS for build perf)
        test: /\.css$/,
        include: monacoEditorPaths,
        use: ['style-loader', 'css-loader'],
      },
      {
        test: extensionHostWorker,
        exclude: /node_modules/,
        use: [
          { loader: 'worker-loader', options: { inline: 'no-fallback' } },
          {
            loader: 'esbuild-loader',
            options: {
              loader: 'ts',
              target: esbuildTarget,
            },
          },
        ],
      },
      { test: /\.ya?ml$/, exclude: /node_modules/, use: ['raw-loader'] },
    ],
  },
}

module.exports = config
