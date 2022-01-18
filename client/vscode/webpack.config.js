// @ts-check

'use strict'
const path = require('path')

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
              stream: require.resolve('stream-browserify'),
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

module.exports = function () {
  return Promise.all([getExtensionCoreConfiguration('node'), getExtensionCoreConfiguration('webworker')])
}
