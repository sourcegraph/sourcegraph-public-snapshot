// @ts-check
'use strict'

const path = require('path')

// Extension Host Worker Path
const extensionHostWorker = /main\.worker\.ts$/

// @ts-check
/** @typedef {import('webpack').Configuration} WebpackConfig **/

/** @type WebpackConfig */
const extensionConfig = {
  target: 'node', // VS Code extensions run in a Node.js-context 📖 -> https://webpack.js.org/configuration/node/
  mode: 'none', // this leaves the source code as close as possible to the original (when packaging we set this to 'production')
  entry: './src/extension.ts', // the entry point of this extension, 📖 -> https://webpack.js.org/configuration/entry-context/
  output: {
    // the bundle is stored in the 'dist' folder (check package.json), 📖 -> https://webpack.js.org/configuration/output/
    path: path.resolve(__dirname, 'dist'),
    filename: 'extension.js',
    library: {
      type: 'umd',
    },
    globalObject: 'globalThis',
  },
  externals: {
    // modules added here also need to be added in the .vscodeignore file
    vscode: 'commonjs vscode', // the vscode-module is created on-the-fly and must be excluded. Add other modules that cannot be webpack'ed, 📖 -> https://webpack.js.org/configuration/externals/
    // bufferutil: 'bufferutil',
    // 'utf-8-validate': 'utf-8-validate',
    // 'webworker-threads': 'webworker-threads',
  },
  resolve: {
    // support reading TypeScript and JavaScript files, 📖 -> https://github.com/TypeStrong/ts-loader
    extensions: ['.ts', '.js', '.tsx', '.jsx'],
  },
  module: {
    rules: [
      {
        test: /\.ts$|tsx/,
        exclude: [/node_modules/, extensionHostWorker],
        use: [
          {
            loader: 'babel-loader',
            options: {
              cacheDirectory: true,
              configFile: path.join('../../', 'babel.config.js'),
            },
          },
        ],
      },
    ],
  },
  devtool: 'nosources-source-map',
  infrastructureLogging: {
    level: 'log', // enables logging required for problem matchers
  },
}
module.exports = [extensionConfig]
