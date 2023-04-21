// @ts-check

// const path = require('path')

// const ReactRefreshWebpackPlugin = require('@pmmmwh/react-refresh-webpack-plugin')
// const SentryWebpackPlugin = require('@sentry/webpack-plugin')
// const CompressionPlugin = require('compression-webpack-plugin')
// const CssMinimizerWebpackPlugin = require('css-minimizer-webpack-plugin')
// const mapValues = require('lodash/mapValues')
// const MiniCssExtractPlugin = require('mini-css-extract-plugin')
// const webpack = require('webpack')
// const { WebpackManifestPlugin } = require('webpack-manifest-plugin')
// const { StatsWriterPlugin } = require('webpack-stats-plugin')

// const {
//   ROOT_PATH,
//   STATIC_ASSETS_PATH,
//   getCacheConfig,
//   getMonacoWebpackPlugin,
//   getBazelCSSLoaders: getCSSLoaders,
//   getTerserPlugin,
//   getProvidePlugin,
//   getCSSModulesLoader,
//   getMonacoTTFRule,
//   getBasicCSSLoader,
//   getStatoscopePlugin,
// } = require('@sourcegraph/build-config')

// const { IS_PRODUCTION, IS_DEVELOPMENT, ENVIRONMENT_CONFIG, writeIndexHTMLPlugin } = require('./dev/utils')
// const { isHotReloadEnabled } = require('./src/integration/environment')

console.log('BUILDKITE_COMMIT', process.env.BUILDKITE_COMMIT)
console.log('MY_VALUE', process.env.MY_VALUE)

const config = {}

module.exports = config
