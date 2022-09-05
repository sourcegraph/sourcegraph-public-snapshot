import path from 'path'

import StatoscopeWebpackPlugin from '@statoscope/webpack-plugin'
import TerserPlugin from 'terser-webpack-plugin'
import webpack, { StatsOptions } from 'webpack'

import { STATIC_ASSETS_PATH } from '../paths'

export const getTerserPlugin = (): TerserPlugin =>
    new TerserPlugin({
        terserOptions: {
            compress: {
                // Don't inline functions, which causes name collisions with uglify-es:
                // https://github.com/mishoo/UglifyJS2/issues/2842
                inline: 1,
            },
        },
    })

export const getProvidePlugin = (): webpack.ProvidePlugin =>
    new webpack.ProvidePlugin({
        // Adding the file extension is necessary to make importing this file
        // work inside JavaScript modules. The alternative is to set
        // `fullySpecified: false` (https://webpack.js.org/configuration/module/#resolvefullyspecified).
        process: 'process/browser.js',
        // Based on the issue: https://github.com/webpack/changelog-v5/issues/10
        Buffer: ['buffer', 'Buffer'],
    })

const STATOSCOPE_STATS: StatsOptions = {
    all: false, // disable all the stats
    hash: true, // compilation hash
    entrypoints: true,
    chunks: true,
    chunkModules: true, // modules
    reasons: true, // modules reasons
    ids: true, // IDs of modules and chunks (webpack 5)
    dependentModules: true, // dependent modules of chunks (webpack 5)
    chunkRelations: true, // chunk parents, children and siblings (webpack 5)
    cachedAssets: true, // information about the cached assets (webpack 5)

    nestedModules: true, // concatenated modules
    usedExports: true,
    providedExports: true, // provided imports
    assets: true,
    chunkOrigins: true, // chunks origins stats (to find out which modules require a chunk)
    version: true, // webpack version
    builtAt: true, // build at time
    timings: true, // modules timing information
    performance: true, // info about oversized assets
}

export const getStatoscopePlugin = (name = '[name]'): StatoscopeWebpackPlugin =>
    new StatoscopeWebpackPlugin({
        statsOptions: STATOSCOPE_STATS as Record<string, unknown>,
        saveStatsTo: path.join(STATIC_ASSETS_PATH, `stats-${name}-[hash].json`),
        saveReportTo: path.join(STATIC_ASSETS_PATH, `report-${name}-[hash].html`),
    })
