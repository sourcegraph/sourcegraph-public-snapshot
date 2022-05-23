// @ts-check

'use strict'
const path = require('path')

const MiniCssExtractPlugin = require('mini-css-extract-plugin')
const webpack = require('webpack')

const {
  getMonacoWebpackPlugin,
  getCSSModulesLoader,
  getBasicCSSLoader,
  getMonacoCSSRule,
  getCSSLoaders,
} = require('@sourcegraph/build-config')

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'

const rootPath = path.resolve(__dirname, '../../')
const jetbrainsWorkspacePath = path.resolve(rootPath, 'client', 'jetbrains')
const webviewSourcePath = path.resolve(jetbrainsWorkspacePath, 'webview', 'src')

// Build artifacts are put directly into the JetBrains resources folder
const distPath = path.resolve(jetbrainsWorkspacePath, 'src', 'main', 'resources', 'dist')

const extensionHostWorker = /main\.worker\.ts$/

const MONACO_EDITOR_PATH = path.resolve(rootPath, 'node_modules', 'monaco-editor')

/** @type {import('webpack').Configuration}*/

const webviewConfig = {
  context: __dirname, // needed when running `gulp` from the root dir
  mode,
  name: 'webviews',
  target: 'web',
  entry: {
    search: path.resolve(webviewSourcePath, 'search', 'index.tsx'),
    style: path.join(webviewSourcePath, 'index.scss'),
  },
  devtool: 'source-map',
  output: {
    path: distPath,
    filename: '[name].js',
  },
  plugins: [
    new MiniCssExtractPlugin(),
    getMonacoWebpackPlugin(),
    new webpack.ProvidePlugin({
      Buffer: ['buffer', 'Buffer'],
      process: 'process/browser', // provide a shim for the global `process` variable
    }),
  ],
  resolve: {
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
    alias: {
      path: require.resolve('path-browserify'),
    },
    fallback: {
      path: require.resolve('path-browserify'),
      process: require.resolve('process/browser'),
      util: require.resolve('util'),
      http: require.resolve('stream-http'),
      https: require.resolve('https-browserify'),
    },
  },
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        exclude: [/node_modules/, extensionHostWorker],
        use: [
          {
            loader: 'ts-loader',
          },
        ],
      },
      {
        test: extensionHostWorker,
        use: [
          {
            loader: 'worker-loader',
            options: { inline: 'no-fallback' },
          },
          'ts-loader',
        ],
      },
      {
        test: /\.(sass|scss)$/,
        // CSS Modules loaders are only applied when the file is explicitly named as CSS module stylesheet using the extension `.module.scss`.
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(MiniCssExtractPlugin.loader, getCSSModulesLoader({})),
      },
      {
        test: /\.(sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders(MiniCssExtractPlugin.loader, getBasicCSSLoader()),
      },
      getMonacoCSSRule(),
      // Don't use shared getMonacoTFFRule(); we want to retain its name
      // to reference path in the extension when we load the font ourselves.
      {
        test: /\.ttf$/,
        include: [MONACO_EDITOR_PATH],
        type: 'asset/resource',
        generator: {
          filename: '[name][ext]',
        },
      },
    ],
  },
}

module.exports = function () {
  return Promise.resolve([webviewConfig])
}
