// @ts-check

'use strict'
const path = require('path')

const MiniCssExtractPlugin = require('mini-css-extract-plugin')

/**
 * The VS Code extension core needs to be built for two targets:
 * - Node.js for VS Code desktop
 * - Web Worker for VS Code web
 *
 * @param {*} targetType See https://webpack.js.org/configuration/target/
 */
function getExtensionCoreConfiguration(targetType) {
  return {
    name: `extension:${targetType}`,
    target: targetType,
    entry: './src/extension.ts', // the entry point of this extension, 📖 -> https://webpack.js.org/configuration/entry-context/
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

const searchPanelWebviewPath = path.resolve(webviewSourcePath, 'search-panel')
const searchSidebarWebviewPath = path.resolve(webviewSourcePath, 'search-sidebar')

const extensionHostWorker = /main\.worker\.ts$/

/** @type {import('webpack').Configuration}*/

const webviewConfig = {
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
  plugins: [new MiniCssExtractPlugin()],
  externals: {
    // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
    vscode: 'commonjs vscode',
  },
  resolve: {
    alias: {},
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
      // SCSS rule for our own styles and Bootstrap
      {
        test: /\.(css|sass|scss)$/,
        exclude: /\.module\.(sass|scss)$/,
        use: getCSSLoaders({ loader: 'css-loader', options: { url: false } }),
      },
      // For CSS modules
      {
        test: /\.(css|sass|scss)$/,
        include: /\.module\.(sass|scss)$/,
        use: getCSSLoaders({
          loader: 'css-loader',
          options: {
            sourceMap: false,
            modules: {
              exportLocalsConvention: 'camelCase',
              localIdentName: '[name]__[local]_[hash:base64:5]',
            },
            url: false,
          },
        }),
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
    ],
  },
}

module.exports = function () {
  return Promise.all([getExtensionCoreConfiguration('node'), getExtensionCoreConfiguration('webworker'), webviewConfig])
}
