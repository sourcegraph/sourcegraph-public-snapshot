import signale from 'signale'
import { Configuration, DllPlugin } from 'webpack'
import { WebpackManifestPlugin } from 'webpack-manifest-plugin'

import {
    getBasicCSSLoader,
    getMonacoCSSRule,
    getMonacoTTFRule,
    getMonacoWebpackPlugin,
} from '@sourcegraph/build-config'

import { getVendorModules, getWebpackStats } from './dllPlugin'
import { dllBundleManifestPath, dllPluginConfig, monacoEditorPath } from './webpack.config.common'

const webpackStats = getWebpackStats()
signale.await('Waiting for Webpack to build DLL bundle based on vendor stats.')

const config: Configuration = {
    mode: 'development',
    stats: 'errors-warnings',
    entry: {
        dll: [...getVendorModules(webpackStats), 'monaco-editor'],
    },
    output: {
        filename: '[name].bundle.[contenthash].js',
        path: dllPluginConfig.context,
        library: dllPluginConfig.name,
        // Required to fix the `WebpackManifestPlugin` output: https://github.com/shellscape/webpack-manifest-plugin/issues/229
        publicPath: '',
    },
    module: {
        rules: [
            getMonacoCSSRule(),
            getMonacoTTFRule(),
            {
                test: /\.css$/,
                exclude: [monacoEditorPath],
                use: [getBasicCSSLoader()],
            },
        ],
    },
    plugins: [
        getMonacoWebpackPlugin(),
        new DllPlugin(dllPluginConfig),
        new WebpackManifestPlugin({ fileName: dllBundleManifestPath }),
    ],
}

module.exports = config
