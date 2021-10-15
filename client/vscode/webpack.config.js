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

/** @type {import('webpack').Configuration}*/
const webviewConfig = {
  target: 'web',
  entry: {
    searchPanel: [path.resolve(searchPanelWebviewPath, 'index.tsx')],
    searchSidebar: [path.resolve(searchSidebarWebviewPath, 'index.tsx')],
    style: path.join(webviewSourcePath, 'index.scss'),
  },
  output: {
    path: path.join(vscodeWorkspacePath, 'dist/webview'),
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
    ],
  },
}

module.exports = [webviewConfig, extensionConfig]
