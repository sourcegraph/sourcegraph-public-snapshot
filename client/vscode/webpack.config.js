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
/**
 * The VS Code extension core needs to be built for two targets:
 * - Node.js for VS Code desktop
 * - Web Worker for VS Code web
 *
 * @param {*} targetType See https://webpack.js.org/configuration/target/
 */
// Node Envs
const mode = process.env.NODE_ENV === 'production' ? 'production' : 'development'
// Core Configuration
function getExtensionCoreConfiguration(targetType) {
  if (typeof targetType !== 'string') {
    return
  }
  return {
    context: __dirname, // needed when running `gulp` from the root dir
    mode,
    name: `extension:${targetType}`,
    target: targetType,
    entry: path.resolve(__dirname, 'src', 'extension.ts'), // the entry point of this extension, 📖 -> https://webpack.js.org/configuration/entry-context/
    output: {
      // the bundle is stored in the 'dist' folder (check package.json), 📖 -> https://webpack.js.org/configuration/output/
      path: path.resolve(__dirname, 'dist', `${targetType}`),
      filename: '[name].js',
      library: {
        type: 'umd',
      },
      globalObject: 'globalThis',
      devtoolModuleFilenameTemplate: '../[resource-path]',
    },
    performance: {
      hints: false,
    },
    optimization: {
      splitChunks: {
        minSize: 10000,
        maxSize: 240000,
      },
    },
    devtool: mode === 'development' ? 'source-map' : false,
    externals: {
      // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
      vscode: 'commonjs vscode',
    },
    resolve: {
      // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
      extensions: ['.ts', '.tsx', '.js', '.jsx'],
      alias:
        targetType === 'webworker'
          ? {
              path: require.resolve('path-browserify'),
              './browserActionsNode': path.resolve(__dirname, 'src', 'link-commands', 'browserActionsWeb'),
            }
          : {
              path: require.resolve('path-browserify'),
            },
      fallback:
        targetType === 'webworker'
          ? {
              process: require.resolve('process/browser'),
              path: require.resolve('path-browserify'),
              assert: require.resolve('assert'),
              util: require.resolve('util'),
              http: require.resolve('stream-http'),
              https: require.resolve('https-browserify'),
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
    plugins: [
      new webpack.ProvidePlugin({
        Buffer: ['buffer', 'Buffer'],
        process: 'process/browser', // provide a shim for the global `process` variable
      }),
      ...(process.env.IS_TEST
        ? [
            new webpack.DefinePlugin({
              'process.env': {
                IS_TEST: true,
              },
            }),
          ]
        : []),
    ],
  }
}
/**
 * Configuration for Webviews
 */
// PATHS
const rootPath = path.resolve(__dirname, '../../')
const vscodeWorkspacePath = path.resolve(rootPath, 'client', 'vscode')
const vscodeSourcePath = path.resolve(vscodeWorkspacePath, 'src')
const webviewSourcePath = path.resolve(vscodeSourcePath, 'webview')
// Webview Panels Paths
const searchPanelWebviewPath = path.resolve(webviewSourcePath, 'search-panel')
const searchSidebarWebviewPath = path.resolve(webviewSourcePath, 'sidebars', 'search')
const helpSidebarWebviewPath = path.resolve(webviewSourcePath, 'sidebars', 'help')
// Extension Host Worker Path
const extensionHostWorker = /main\.worker\.ts$/
// Monaco Editor Path
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
    helpSidebar: [path.resolve(helpSidebarWebviewPath, 'index.tsx')],
    style: path.join(webviewSourcePath, 'index.scss'),
  },
  devtool: mode === 'development' ? 'source-map' : false,
  output: {
    path: path.resolve(__dirname, 'dist/webview'),
    filename: '[name].js',
  },
  performance: {
    hints: false,
  },
  optimization: {
    splitChunks: {
      minSize: 10000,
      maxSize: 250000,
    },
  },
  plugins: [
    new MiniCssExtractPlugin(),
    getMonacoWebpackPlugin(),
    new webpack.ProvidePlugin({
      Buffer: ['buffer', 'Buffer'],
      process: 'process/browser', // provide a shim for the global `process` variable
    }),
  ],
  externals: {
    // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
    vscode: 'commonjs vscode',
  },
  resolve: {
    alias: {
      path: require.resolve('path-browserify'),
      './RepoSearchResult': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'RepoSearchResult'),
      './CommitSearchResult': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'CommitSearchResult'),
      './FileMatchChildren': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'FileMatchChildren'),
      './RepoFileLink': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'RepoFileLink'),
      '../documentation/ModalVideo': path.resolve(__dirname, 'src', 'webview', 'search-panel', 'alias', 'ModalVideo'),
    },
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.tsx', '.js', '.jsx'],
    fallback: {
      path: require.resolve('path-browserify'),
      process: require.resolve('process/browser'),
      http: require.resolve('stream-http'), // for stream search - event source polyfills
      https: require.resolve('https-browserify'), // for stream search - event source polyfills
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
  if (process.env.TARGET_TYPE) {
    return Promise.all([getExtensionCoreConfiguration(process.env.TARGET_TYPE), webviewConfig])
  }
  // If target type isn't specified, build both.
  return Promise.all([getExtensionCoreConfiguration('node'), getExtensionCoreConfiguration('webworker'), webviewConfig])
}
