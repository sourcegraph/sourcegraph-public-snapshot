// @ts-check

'use strict'

const path = require('path')

const MiniCssExtractPlugin = require('mini-css-extract-plugin')

/** @type {import('webpack').Configuration}*/
const extensionConfig = {
  target: 'node', // vscode extensions run in a Node.js-context 📖 -> https://webpack.js.org/configuration/node/

  entry: './src/extension.ts', // the entry point of this extension, 📖 -> https://webpack.js.org/configuration/entry-context/
  output: {
    // the bundle is stored in the 'dist' folder (check package.json), 📖 -> https://webpack.js.org/configuration/output/
    path: path.resolve(__dirname, 'dist'),
    filename: 'extension.js',
    library: {
      type: 'umd',
    },
    globalObject: 'globalThis',
    devtoolModuleFilenameTemplate: '../[resource-path]',
  },
  devtool: 'source-map',
  externals: {
    vscode: 'commonjs vscode', // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
  },
  resolve: {
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'ts-loader',
          },
        ],
      },
    ],
  },
}

const rootPath = path.resolve(__dirname, '../../')
const vscodeWorkspacePath = path.resolve(rootPath, 'client', 'vscode')
const vscodeSourcePath = path.resolve(vscodeWorkspacePath, 'src')
const webviewsSourcePath = path.resolve(vscodeSourcePath, 'webviews')

const getCSSLoaders = (...loaders) => [
  MiniCssExtractPlugin.loader,
  ...loaders,
  {
    loader: 'postcss-loader',
  },
  {
    loader: 'sass-loader',
    options: {
      sassOptions: {
        includePaths: [path.resolve(rootPath, 'node_modules'), path.resolve(rootPath, 'client')],
      },
    },
  },
]

/** @type {import('webpack').Configuration}*/
const webviewConfig = {
  target: 'web',
  entry: {
    search: [path.resolve(webviewsSourcePath, 'search', 'index.tsx')],
    // Styles
    style: path.join(webviewsSourcePath, 'app.scss'),
  },
  output: {
    path: path.join(vscodeWorkspacePath, 'dist/webviews'),
    filename: '[name].js',
  },
  plugins: [new MiniCssExtractPlugin()],
  resolve: {
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'ts-loader',
          },
        ],
      },
      // SCSS rule for our own styles and Bootstrap
      {
        test: /\.(css|sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders({ loader: 'css-loader', options: { url: false } }),
      },
      // For CSS modules
    ],
  },
}

module.exports = [webviewConfig, extensionConfig]
