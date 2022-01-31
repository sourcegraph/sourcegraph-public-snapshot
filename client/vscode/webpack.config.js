// @ts-check

'use strict'
const path = require('path')

const MiniCssExtractPlugin = require('mini-css-extract-plugin')

const {
  getMonacoWebpackPlugin,
  getCSSModulesLoader,
  getBasicCSSLoader,
  getMonacoCSSRule,
  getCSSLoaders,
} = require('@sourcegraph/build-config')

const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'

/**
 * The VS Code extension core needs to be built for two targets:
 * - Node.js for VS Code desktop
 * - Web Worker for VS Code web
 *
 * @param {*} targetType See https://webpack.js.org/configuration/target/
 */
function getExtensionCoreConfiguration(targetType) {
  return {
    context: __dirname, // needed when running `gulp` from the root dir
    mode,
    name: `extension:${targetType}`,
    target: targetType,
    entry: path.resolve(__dirname, 'src', 'extension.ts'), // the entry point of this extension, 📖 -> https://webpack.js.org/configuration/entry-context/
    output: {
      // the bundle is stored in the 'dist' folder (check package.json), 📖 -> https://webpack.js.org/configuration/output/
      path: path.resolve(__dirname, 'dist', `${targetType}`),
      filename: 'extension.js',
      library: {
        type: 'umd',
      },
      globalObject: 'globalThis',
      devtoolModuleFilenameTemplate: '../[resource-path]',
    },
    devtool: 'source-map',
    externals: {
      // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
      vscode: 'commonjs vscode',
    },
    resolve: {
      // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
      extensions: ['.ts', '.tsx', '.js', '.jsx'],
      alias: {},
      fallback:
        targetType === 'webworker'
          ? {
              process: require.resolve('process/browser'),
              path: require.resolve('path-browserify'),
              assert: require.resolve('assert'),
              util: require.resolve('util'),
            }
          : {},
    },
    module: {
      rules: [
        {
          test: /\.tsx?$/,
          exclude: /node_modules/,
          use: [
            {
              // TODO(tj): esbuild-loader https://github.com/privatenumber/esbuild-loader
              loader: 'ts-loader',
            },
          ],
        },
      ],
    },
  }
}

const rootPath = path.resolve(__dirname, '../../')
const vscodeWorkspacePath = path.resolve(rootPath, 'client', 'vscode')
const vscodeSourcePath = path.resolve(vscodeWorkspacePath, 'src')
const webviewSourcePath = path.resolve(vscodeSourcePath, 'webview')

const searchPanelWebviewPath = path.resolve(webviewSourcePath, 'search-panel')
const searchSidebarWebviewPath = path.resolve(webviewSourcePath, 'search-sidebar')

const extensionHostWorker = /main\.worker\.ts$/

const MONACO_EDITOR_PATH = path.resolve(rootPath, 'node_modules', 'monaco-editor')

/** @type {import('webpack').Configuration}*/

const webviewConfig = {
  context: __dirname, // needed when running `gulp` from the root dir
  mode,
  name: 'webviews',
  target: 'web',
  entry: {
    searchPanel: [path.resolve(searchPanelWebviewPath, 'index.tsx')],
    searchSidebar: [path.resolve(searchSidebarWebviewPath, 'index.tsx')],
    style: path.join(webviewSourcePath, 'index.scss'),
  },
  devtool: 'source-map',
  output: {
    path: path.resolve(__dirname, 'dist/webview'),
    filename: '[name].js',
  },
  plugins: [new MiniCssExtractPlugin(), getMonacoWebpackPlugin()],
  externals: {
    // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
    vscode: 'commonjs vscode',
  },
  resolve: {
    alias: {
      '../documentation/ModalVideo': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'ModalVideo'), // For NoResultsPage
    },
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
    fallback: {
      path: require.resolve('path-browserify'),
      process: require.resolve('process/browser'),
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

module.exports = function (targetType) {
  if (targetType) {
    return Promise.all([getExtensionCoreConfiguration(targetType), webviewConfig])
  }

  // If target type isn't specified, build both.
  return Promise.all([getExtensionCoreConfiguration('node'), getExtensionCoreConfiguration('webworker'), webviewConfig])
}
