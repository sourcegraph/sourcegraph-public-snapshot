import path from 'path'

import CssMinimizerWebpackPlugin from 'css-minimizer-webpack-plugin'
import MiniCssExtractPlugin from 'mini-css-extract-plugin'
import type webpack from 'webpack'
import { optimize } from 'webpack'

import {
    getBabelLoader,
    getCSSLoaders,
    getProvidePlugin,
    getTerserPlugin,
    getCSSModulesLoader,
    getBasicCSSLoader,
} from '@sourcegraph/build-config'

import { browserWorkspacePath, entrypoints } from '../buildCommon'

const extensionHostWorker = /main\.worker\.ts$/

export const config = {
    target: 'browserslist',
    entry: entrypoints,
    output: {
        path: path.join(browserWorkspacePath, 'build/dist/js'),
        filename: '[name].bundle.js',
        chunkFilename: '[id].chunk.js',
    },
    devtool: 'inline-cheap-module-source-map',
    optimization: {
        minimizer: [getTerserPlugin() as webpack.WebpackPluginInstance, new CssMinimizerWebpackPlugin()],
    },

    plugins: [
        new MiniCssExtractPlugin({ filename: '../css/[name].bundle.css' }),
        // Code splitting doesn't make sense/work in the browser extension, but we still want to use dynamic import()
        new optimize.LimitChunkCountPlugin({ maxChunks: 1 }),
        getProvidePlugin(),
    ],
    resolve: {
        extensions: ['.ts', '.tsx', '.js'],
        alias: {
            path: require.resolve('path-browserify'),
        },
    },
    module: {
        rules: [
            {
                test: /\.[jt]sx?$/,
                exclude: extensionHostWorker,
                use: [getBabelLoader()],
            },
            {
                // SCSS rule for our own and third-party styles
                test: /\.(css|sass|scss)$/,
                exclude: /\.module\.(sass|scss)$/,
                use: getCSSLoaders(MiniCssExtractPlugin.loader, getBasicCSSLoader()),
            },
            {
                test: /\.(css|sass|scss)$/,
                include: /\.module\.(sass|scss)$/,
                use: getCSSLoaders(MiniCssExtractPlugin.loader, getCSSModulesLoader({ sourceMap: false, url: false })),
            },
            {
                test: /\.svg$/i,
                type: 'asset/inline',
            },
            {
                test: extensionHostWorker,
                use: [
                    {
                        loader: 'worker-loader',
                        options: { filename: 'extensionHostWorker.bundle.js' },
                    },
                    getBabelLoader(),
                ],
            },
        ],
    },
} satisfies webpack.Configuration
